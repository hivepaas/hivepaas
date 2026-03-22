package taskcronjobexec

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

const (
	subjectPrefixSystem = "[System]"
)

func (e *Executor) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notifConfig := data.CronJob.Notification
	if notifConfig == nil {
		return nil
	}

	isSuccess := data.Task.IsDone()
	notifSettingID := gofn.If(isSuccess, notifConfig.Success.ID, notifConfig.Failure.ID)
	var notification *entity.Notification
	if notifSettingID == "" {
		if (isSuccess && !notifConfig.SuccessUseDefault) || (!isSuccess && !notifConfig.FailureUseDefault) {
			return nil
		}
		notification, err = e.getDefaultNotification(ctx, db, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		notifSetting := data.RefObjects.RefSettings[notifSettingID]
		if notifSetting == nil {
			return nil
		}
		notification = notifSetting.MustAsNotification()
	}
	if notification == nil {
		return nil
	}

	var execFuncs []func(ctx context.Context) error

	if notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sendNotificationViaEmail(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sendNotificationViaSlack(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sendNotificationViaDiscord(ctx, db, notification, data)
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.buildNotificationMsgData(data)

	err = gofn.ExecTasks(ctx, 0, execFuncs...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) getDefaultNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (*entity.Notification, error) {
	var scope *base.SettingScope
	switch {
	case data.App != nil:
		scope = data.App.GetSettingScope()
	case data.Project != nil:
		scope = data.Project.GetSettingScope()
	default:
		scope = &base.SettingScope{}
	}

	setting, err := e.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
		bunex.SelectWhere("setting.is_default = TRUE"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}
	notification := setting.MustAsNotification()

	// Load ref objects of the setting (otherwise we will have error of missing ref objects)
	refObjects, err := e.settingService.LoadReferenceObjects(ctx, db, scope, true,
		false, setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)
	data.RefObjects.RefSettings[setting.ID] = setting

	return notification, nil
}

func (e *Executor) buildNotificationMsgData(
	data *taskData,
) {
	msgData := &notificationservice.BaseMsgDataCronTaskNotification{
		Succeeded:   data.Task.IsDone(),
		CronJobName: data.CronJobSetting.Name,
		CreatedAt:   data.CronJob.Schedule.InitialTime,
		StartedAt:   data.Task.StartedAt,
		Duration:    data.Task.GetDuration(),
		Retries:     data.Task.Config.Retry,
	}
	if data.CronJob.Schedule.Interval > 0 {
		msgData.Schedule = fmt.Sprintf("every %v", data.CronJob.Schedule.Interval.String())
	} else {
		msgData.Schedule = fmt.Sprintf("cron expression %v", data.CronJob.Schedule.CronExpr)
	}
	if data.Project != nil {
		msgData.ProjectName = data.Project.Name
	}
	if data.App != nil {
		msgData.AppName = data.App.Name
	}
	switch {
	case data.App != nil:
		msgData.DashboardLink = config.Current.DashboardAppCronTaskDetailsURL(data.App.ID, data.App.ProjectID,
			data.CronJobSetting.ID, data.Task.ID)
	case data.Project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectCronTaskDetailsURL(data.Project.ID,
			data.CronJobSetting.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalCronTaskDetailsURL(
			data.CronJobSetting.ID, data.Task.ID)
	}
	data.NotifMsgData = msgData
}

func (e *Executor) sendNotificationViaEmail(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	data *taskData,
) error {
	if notification == nil || notification.ViaEmail == nil {
		return nil
	}

	emailSetting := data.RefObjects.RefSettings[notification.ViaEmail.Sender.ID]
	if emailSetting == nil {
		return apperrors.NewMissing("Sender email account")
	}
	emailAcc := emailSetting.MustAsEmail()
	if emailAcc == nil {
		return apperrors.NewMissing("Sender email account")
	}

	userMap, err := e.userService.LoadProjectUsers(ctx, db, data.Project, notification.ViaEmail.ToProjectMembers,
		notification.ViaEmail.ToProjectOwners, notification.ViaEmail.ToAllAdmins)
	if err != nil {
		return apperrors.Wrap(err)
	}

	userEmails := make([]string, 0, len(userMap))
	for _, user := range userMap {
		userEmails = append(userEmails, user.Email)
	}
	if len(notification.ViaEmail.ToAddresses) > 0 {
		userEmails = gofn.ToSet(append(userEmails, notification.ViaEmail.ToAddresses...))
	}
	if len(userEmails) == 0 {
		return nil
	}

	subject := subjectPrefixSystem
	if data.Project != nil {
		subject = fmt.Sprintf("[%s]", data.Project.Name)
	}
	if data.App != nil {
		subject += fmt.Sprintf("[%s]", data.App.Name)
	}
	subject += gofn.If(data.Task.IsDone(), " Scheduled task succeeded", " Scheduled task failed")

	err = e.notificationService.EmailSendCronTaskNotification(ctx, db,
		&notificationservice.EmailMsgDataCronTaskNotification{
			BaseMsgDataCronTaskNotification: data.NotifMsgData,
			Email:                           emailAcc,
			Recipients:                      userEmails,
			Subject:                         subject,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sendNotificationViaSlack(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	data *taskData,
) error {
	if notification == nil || notification.ViaSlack == nil {
		return nil
	}

	imSetting := data.RefObjects.RefSettings[notification.ViaSlack.Webhook.ID]
	if imSetting == nil {
		return apperrors.NewMissing("Slack webhook")
	}
	imService := imSetting.MustAsIMService()
	if imService == nil || imService.Slack == nil {
		return apperrors.NewMissing("Slack webhook")
	}

	err := e.notificationService.SlackSendCronTaskNotification(ctx, db,
		&notificationservice.SlackMsgDataCronTaskNotification{
			BaseMsgDataCronTaskNotification: data.NotifMsgData,
			Setting:                         imService.Slack,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sendNotificationViaDiscord(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	data *taskData,
) error {
	if notification == nil || notification.ViaDiscord == nil {
		return nil
	}

	imSetting := data.RefObjects.RefSettings[notification.ViaDiscord.Webhook.ID]
	if imSetting == nil {
		return apperrors.NewMissing("Discord webhook")
	}
	imService := imSetting.MustAsIMService()
	if imService == nil || imService.Discord == nil {
		return apperrors.NewMissing("Discord webhook")
	}

	err := e.notificationService.DiscordSendCronTaskNotification(ctx, db,
		&notificationservice.DiscordMsgDataCronTaskNotification{
			BaseMsgDataCronTaskNotification: data.NotifMsgData,
			Setting:                         imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
