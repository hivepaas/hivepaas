package taskhealthcheck

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

const (
	eventSuccess = "success"
	eventFailure = "failure"
)

//nolint:unparam
func (e *Executor) sendNotification(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notifConfig := data.Healthcheck.Notification
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
	minSendingInterval := notification.MinSendInterval.ToDuration()

	currEvent := gofn.If(data.Task.IsDone(), eventSuccess, eventFailure)
	lastNotifEvent := data.NotifEventMap[data.HealthcheckSetting.ID]
	timeNow := timeutil.NowUTC()
	shouldSkipNotifEmail := false
	shouldSkipNotifSlack := false
	shouldSkipNotifDiscord := false

	if minSendingInterval > 0 && lastNotifEvent != nil && lastNotifEvent.Event == currEvent {
		if timeNow.Sub(lastNotifEvent.Ts) < minSendingInterval {
			shouldSkipNotifEmail = lastNotifEvent.EmailSent
			shouldSkipNotifSlack = lastNotifEvent.SlackSent
			shouldSkipNotifDiscord = lastNotifEvent.DiscordSent
		}
	}

	var execFuncs []func(ctx context.Context) error
	emailSent := false
	slackSent := false
	discordSent := false

	if !shouldSkipNotifEmail && notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := e.sendNotificationViaEmail(ctx, db, notification, data)
			emailSent = err == nil
			return err
		})
	}
	if !shouldSkipNotifSlack && notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := e.sendNotificationViaSlack(ctx, db, notification, data)
			slackSent = err == nil
			return err
		})
	}
	if !shouldSkipNotifDiscord && notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := e.sendNotificationViaDiscord(ctx, db, notification, data)
			discordSent = err == nil
			return err
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.buildNotificationMsgData(data)

	_ = gofn.ExecTasksEx(ctx, 20, false, execFuncs...) //nolint

	// Update notification events in redis
	if minSendingInterval > 0 {
		_ = e.notifEventRepo.Set(ctx, data.HealthcheckSetting.ID, &cacheentity.HealthcheckNotifEvent{
			Event:       currEvent,
			Ts:          timeNow,
			EmailSent:   emailSent,
			SlackSent:   slackSent,
			DiscordSent: discordSent,
		}, minSendingInterval)
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
	msgData := &notificationservice.BaseMsgDataHealthcheckNotification{
		Succeeded:       data.Task.IsDone(),
		HealthcheckName: data.HealthcheckSetting.Name,
		HealthcheckType: data.Healthcheck.HealthcheckType,
		StartedAt:       data.Task.StartedAt,
		Duration:        data.Task.GetDuration(),
		Retries:         data.Task.Config.Retry,
	}
	if data.Project != nil {
		msgData.ProjectName = data.Project.Name
	}
	if data.App != nil {
		msgData.AppName = data.App.Name
	}
	switch {
	case data.App != nil:
		msgData.DashboardLink = config.Current.DashboardAppHealthcheckDetailsURL(data.App.ID, data.App.ProjectID,
			data.HealthcheckSetting.ID, data.Task.ID)
	case data.Project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectHealthcheckDetailsURL(data.Project.ID,
			data.HealthcheckSetting.ID, data.Task.ID)
	}

	output, _ := data.Task.OutputAsHealthcheck()
	if output.REST != nil && data.Healthcheck.REST != nil {
		input := data.Healthcheck.REST
		maxLen := 100
		pad := "..."
		if output.REST.ReturnCode != 0 {
			msgData.Expect = fmt.Sprintf("Status code = %v", input.ReturnCode)
			msgData.Actual = fmt.Sprintf("Status code = %v", output.REST.ReturnCode)
		}
		if output.REST.ReturnText != "" {
			msgData.Expect = fmt.Sprintf("Text = %v", strutil.CutShort(input.ReturnText, maxLen, pad))
			msgData.Actual = fmt.Sprintf("Text = %v", strutil.CutShort(output.REST.ReturnText, maxLen, pad))
		}
		if output.REST.ReturnJSON != "" {
			msgData.Expect = fmt.Sprintf("JSON = %v", strutil.CutShort(input.ReturnJSON, maxLen, pad))
			msgData.Actual = fmt.Sprintf("JSON = %v", strutil.CutShort(output.REST.ReturnJSON, maxLen, pad))
		}
	}
	if output.GRPC != nil && data.Healthcheck.GRPC != nil {
		if output.GRPC.ReturnStatus != 0 {
			msgData.Expect = fmt.Sprintf("Status = %v", data.Healthcheck.GRPC.ReturnStatus)
			msgData.Actual = fmt.Sprintf("Status = %v", output.GRPC.ReturnStatus)
		}
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
	emailAcc := emailSetting.MustAsEmail() //nolint
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

	subject := "[System]"
	if data.Project != nil {
		subject = fmt.Sprintf("[%s]", data.Project.Name)
	}
	if data.App != nil {
		subject += fmt.Sprintf("[%s]", data.App.Name)
	}
	subject += gofn.If(data.Task.IsDone(), " Healthcheck succeeded", " Healthcheck failed")

	err = e.notificationService.EmailSendHealthcheckNotification(ctx, db,
		&notificationservice.EmailMsgDataHealthcheckNotification{
			BaseMsgDataHealthcheckNotification: data.NotifMsgData,
			Email:                              emailAcc,
			Recipients:                         userEmails,
			Subject:                            subject,
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
	imService := imSetting.MustAsIMService() //nolint
	if imService == nil || imService.Slack == nil {
		return apperrors.NewMissing("Slack webhook")
	}

	err := e.notificationService.SlackSendHealthcheckNotification(ctx, db,
		&notificationservice.SlackMsgDataHealthcheckNotification{
			BaseMsgDataHealthcheckNotification: data.NotifMsgData,
			Setting:                            imService.Slack,
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
	imService := imSetting.MustAsIMService() //nolint
	if imService == nil || imService.Discord == nil {
		return apperrors.NewMissing("Discord webhook")
	}

	err := e.notificationService.DiscordSendHealthcheckNotification(ctx, db,
		&notificationservice.DiscordMsgDataHealthcheckNotification{
			BaseMsgDataHealthcheckNotification: data.NotifMsgData,
			Setting:                            imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
