package taskcronjobexec

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

func (e *Executor) sslNotifyOfExpiration(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) (err error) {
	notifSetting, err := e.sslGetNotification(ctx, db, item.Setting, false, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if notifSetting == nil {
		return nil
	}
	notification := notifSetting.MustAsNotification()

	var execFuncs []func(ctx context.Context) error

	if notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sslSendExpiringNotificationViaEmail(ctx, db, notification, item, data)
		})
	}
	if notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sslSendExpiringNotificationViaSlack(ctx, db, notification, item, data)
		})
	}
	if notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sslSendExpiringNotificationViaDiscord(ctx, db, notification, item, data)
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.sslBuildExpiringNotificationMsgData(item, data)

	err = gofn.ExecTasks(ctx, 0, execFuncs...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslBuildExpiringNotificationMsgData(
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) {
	ssl := item.Setting.MustAsSSL()
	msgData := &notificationservice.BaseMsgDataSSLExpiringNotification{
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.Provider),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
		ExpireIn:  timeutil.Duration(ssl.ExpireAt.Sub(timeutil.NowUTC()).Truncate(time.Hour)),
	}
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
	}
	switch {
	case app != nil:
		msgData.DashboardLink = config.Current.DashboardAppCronTaskDetailsURL(app.ID, app.ProjectID,
			data.CronJobSetting.ID, data.Task.ID)
	case project != nil:
		msgData.DashboardLink = config.Current.DashboardProjectCronTaskDetailsURL(project.ID,
			data.CronJobSetting.ID, data.Task.ID)
	default:
		msgData.DashboardLink = config.Current.DashboardGlobalCronTaskDetailsURL(
			data.CronJobSetting.ID, data.Task.ID)
	}
	item.ExpiringNotifMsgData = msgData
}

func (e *Executor) sslSendExpiringNotificationViaEmail(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
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

	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp

	userMap, err := e.userService.LoadProjectUsers(ctx, db, project, notification.ViaEmail.ToProjectMembers,
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
	if project != nil {
		subject = fmt.Sprintf("[%s]", project.Name)
	}
	if app != nil {
		subject += fmt.Sprintf("[%s]", app.Name)
	}
	subject += fmt.Sprintf(" Your SSL expiring in %v", item.ExpiringNotifMsgData.ExpireIn)

	err = e.notificationService.EmailSendSSLExpiringNotification(ctx, db,
		&notificationservice.EmailMsgDataSSLExpiringNotification{
			BaseMsgDataSSLExpiringNotification: item.ExpiringNotifMsgData,
			Email:                              emailAcc,
			Recipients:                         userEmails,
			Subject:                            subject,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslSendExpiringNotificationViaSlack(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
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

	err := e.notificationService.SlackSendSSLExpiringNotification(ctx, db,
		&notificationservice.SlackMsgDataSSLExpiringNotification{
			BaseMsgDataSSLExpiringNotification: item.ExpiringNotifMsgData,
			Setting:                            imService.Slack,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslSendExpiringNotificationViaDiscord(
	ctx context.Context,
	db database.IDB,
	notification *entity.Notification,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
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

	err := e.notificationService.DiscordSendSSLExpiringNotification(ctx, db,
		&notificationservice.DiscordMsgDataSSLExpiringNotification{
			BaseMsgDataSSLExpiringNotification: item.ExpiringNotifMsgData,
			Setting:                            imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
