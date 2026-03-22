package taskappdeploy

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

func (e *Executor) notifyForDeployment(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	notifConfig := data.Deployment.Settings.Notification
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
			return e.notifyForDeploymentViaEmail(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.notifyForDeploymentViaSlack(ctx, db, notification, data)
		})
	}
	if notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			return e.notifyForDeploymentViaDiscord(ctx, db, notification, data)
		})
	}
	if len(execFuncs) == 0 {
		return nil
	}

	e.buildDeploymentNotifMsgData(data)

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
	scope := data.App.GetSettingScope()
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

func (e *Executor) buildDeploymentNotifMsgData(
	data *taskData,
) {
	deployment := data.Deployment

	msgData := &notificationservice.BaseMsgDataAppDeploymentNotification{
		ProjectName:   data.Project.Name,
		AppName:       data.App.Name,
		Succeeded:     deployment.IsDone(),
		Method:        deployment.Settings.ActiveMethod,
		StartedAt:     deployment.StartedAt,
		Duration:      deployment.GetDuration(),
		DashboardLink: config.Current.DashboardDeploymentDetailsURL(deployment.ID),
	}
	data.NotifMsgData = msgData

	switch deployment.Settings.ActiveMethod {
	case base.DeploymentMethodRepo:
		msgData.RepoURL = deployment.Settings.RepoSource.RepoURL
		msgData.RepoRef = deployment.Settings.RepoSource.RepoRef
		if deployment.Output != nil {
			msgData.CommitMsg = deployment.Output.CommitMessage
		}
	case base.DeploymentMethodImage:
		msgData.Image = deployment.Settings.ImageSource.Image
	}
}

func (e *Executor) notifyForDeploymentViaEmail(
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

	subject := fmt.Sprintf("[%s][%s]", data.Project.Name, data.App.Name)
	subject += gofn.If(data.Deployment.IsDone(), " Deployment succeeded", " Deployment failed")

	err = e.notificationService.EmailSendAppDeploymentNotification(ctx, db,
		&notificationservice.EmailMsgDataAppDeploymentNotification{
			BaseMsgDataAppDeploymentNotification: data.NotifMsgData,
			Email:                                emailAcc,
			Recipients:                           userEmails,
			Subject:                              subject,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) notifyForDeploymentViaSlack(
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

	err := e.notificationService.SlackSendAppDeploymentNotification(ctx, db,
		&notificationservice.SlackMsgDataAppDeploymentNotification{
			BaseMsgDataAppDeploymentNotification: data.NotifMsgData,
			Setting:                              imService.Slack,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) notifyForDeploymentViaDiscord(
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

	err := e.notificationService.DiscordSendAppDeploymentNotification(ctx, db,
		&notificationservice.DiscordMsgDataAppDeploymentNotification{
			BaseMsgDataAppDeploymentNotification: data.NotifMsgData,
			Setting:                              imService.Discord,
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
