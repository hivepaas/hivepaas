package tasksystemupdate

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (e *Executor) notifyForSystemUpdate(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notification, err := e.getDefaultNotification(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notification == nil {
		return nil
	}

	var execFuncs []func(ctx context.Context) error

	if notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.notifyForSystemUpdateViaEmail(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.notifyForSystemUpdateViaSlack(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.notifyForSystemUpdateViaDiscord(ctx, db, notification, data)
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.buildSystemUpdateNotifMsgData(data)

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
	scope := base.NewSettingScopeGlobal()
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
	data.RefObjects.AddRefObjects(refObjects)
	data.RefObjects.RefSettings[setting.ID] = setting

	return notification, nil
}

func (e *Executor) buildSystemUpdateNotifMsgData(
	data *taskData,
) {
	task := data.Task
	args := gofn.Must(task.ArgsAsSystemUpdate())

	msgData := &notificationservice.BaseMsgDataSystemUpdateNotification{
		CurrentVersion: args.CurrentVersion.AppVersion,
		TargetVersion:  args.TargetVersion.AppVersion,
		Succeeded:      task.IsDone(),
		StartedAt:      task.StartedAt,
		Duration:       task.GetDuration(),
		DashboardLink:  config.Current.DashboardTaskDetailsURL(task.ID),
	}
	data.NotifMsgData = msgData
}

func (e *Executor) notifyForSystemUpdateViaEmail(
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

	userMap, err := e.userService.LoadProjectUsers(ctx, db, nil, notification.ViaEmail.ToProjectMembers,
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

	subject := gofn.If(data.Task.IsDone(), "System update succeeded", "System update failed")

	err = e.notificationService.EmailSendSystemUpdateNotification(ctx, db,
		&notificationservice.EmailMsgDataSystemUpdateNotification{
			BaseMsgDataSystemUpdateNotification: data.NotifMsgData,
			Email:                               emailAcc,
			Recipients:                          userEmails,
			Subject:                             subject,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) notifyForSystemUpdateViaSlack(
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

	err := e.notificationService.SlackSendSystemUpdateNotification(ctx, db,
		&notificationservice.SlackMsgDataSystemUpdateNotification{
			BaseMsgDataSystemUpdateNotification: data.NotifMsgData,
			Setting:                             imService.Slack,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) notifyForSystemUpdateViaDiscord(
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

	err := e.notificationService.DiscordSendSystemUpdateNotification(ctx, db,
		&notificationservice.DiscordMsgDataSystemUpdateNotification{
			BaseMsgDataSystemUpdateNotification: data.NotifMsgData,
			Setting:                             imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
