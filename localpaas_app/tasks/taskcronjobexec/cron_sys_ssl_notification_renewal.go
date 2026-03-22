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

func (e *Executor) sslNotifyOfRenewal(
	ctx context.Context,
	db database.IDB,
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) (err error) {
	notifSetting, err := e.sslGetNotification(ctx, db, item.Setting, item.RenewalError == nil, data)
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
			return e.sslSendRenewalNotificationViaEmail(ctx, db, notification, item, data)
		})
	}
	if notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sslSendRenewalNotificationViaSlack(ctx, db, notification, item, data)
		})
	}
	if notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.sslSendRenewalNotificationViaDiscord(ctx, db, notification, item, data)
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.sslBuildRenewalNotificationMsgData(item, data)

	err = gofn.ExecTasks(ctx, 0, execFuncs...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslBuildRenewalNotificationMsgData(
	item *sslRenewalTaskItem,
	data *sslRenewalTaskData,
) {
	ssl := item.Setting.MustAsSSL()
	timeNow := timeutil.NowUTC()
	msgData := &notificationservice.BaseMsgDataSSLRenewalNotification{
		Succeeded: item.RenewalError == nil,
		SSLName:   item.Setting.Name,
		SSLType:   string(ssl.Provider),
		Domain:    ssl.Domain,
		CreatedAt: item.Setting.CreatedAt,
		ExpireAt:  ssl.ExpireAt,
	}
	project := item.Setting.BelongToProject
	app := item.Setting.BelongToApp
	if project != nil {
		msgData.ProjectName = project.Name
	}
	if app != nil {
		msgData.AppName = app.Name
	}
	if !ssl.RenewableFrom.IsZero() && ssl.RenewableFrom.After(timeNow) {
		msgData.NextRenewalIn = timeutil.Duration(ssl.RenewableFrom.Sub(timeNow).Truncate(time.Hour))
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
	item.RenewalNotifMsgData = msgData
}

func (e *Executor) sslSendRenewalNotificationViaEmail(
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
	subject += gofn.If(item.RenewalError == nil, " SSL renewal succeeded", " SSL renewal failed")

	err = e.notificationService.EmailSendSSLRenewalNotification(ctx, db,
		&notificationservice.EmailMsgDataSSLRenewalNotification{
			BaseMsgDataSSLRenewalNotification: item.RenewalNotifMsgData,
			Email:                             emailAcc,
			Recipients:                        userEmails,
			Subject:                           subject,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslSendRenewalNotificationViaSlack(
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

	err := e.notificationService.SlackSendSSLRenewalNotification(ctx, db,
		&notificationservice.SlackMsgDataSSLRenewalNotification{
			BaseMsgDataSSLRenewalNotification: item.RenewalNotifMsgData,
			Setting:                           imService.Slack,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sslSendRenewalNotificationViaDiscord(
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

	err := e.notificationService.DiscordSendSSLRenewalNotification(ctx, db,
		&notificationservice.DiscordMsgDataSSLRenewalNotification{
			BaseMsgDataSSLRenewalNotification: item.RenewalNotifMsgData,
			Setting:                           imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
