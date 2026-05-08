package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/services/email"
)

func (s *service) NotifyForTaskResult(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) (resp *notificationservice.TaskResultNotificationResp, err error) {
	resp = &notificationservice.TaskResultNotificationResp{}
	notification := data.Notification
	if notification == nil {
		return resp, nil
	}

	minSendingInterval := notification.MinSendInterval.ToDuration()
	currEvent := gofn.If(data.ActionSucceeded, "success", "failure")
	shouldSkipNotifEmail := false
	shouldSkipNotifSlack := false
	shouldSkipNotifDiscord := false

	if minSendingInterval > 0 && !data.LastSendTimestamp.IsZero() && data.LastSendEvent == currEvent {
		if time.Since(data.LastSendTimestamp) < minSendingInterval {
			shouldSkipNotifEmail = data.LastEmailSent
			shouldSkipNotifSlack = data.LastSlackSent
			shouldSkipNotifDiscord = data.LastDiscordSent
		}
	}

	var execFuncs []func(ctx context.Context) error

	if !shouldSkipNotifEmail && notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaEmail(ctx, db, data)
			resp.EmailSent = err == nil
			return err
		})
	}
	if !shouldSkipNotifSlack && notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaSlack(ctx, db, data)
			resp.SlackSent = err == nil
			return err
		})
	}
	if !shouldSkipNotifDiscord && notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaDiscord(ctx, db, data)
			resp.DiscordSent = err == nil
			return err
		})
	}
	if len(execFuncs) == 0 {
		return resp, nil
	}

	err = gofn.ExecTasks(ctx, 0, execFuncs...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (s *service) notifyForTaskResultViaEmail(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
	if notification == nil || notification.ViaEmail == nil {
		return nil
	}

	viaEmail := notification.ViaEmail
	emailSetting := data.RefObjects.RefSettings[viaEmail.Sender.ID]
	if emailSetting == nil {
		return apperrors.NewMissing("Sender email account")
	}
	emailAcc := emailSetting.MustAsEmail()
	if emailAcc == nil {
		return apperrors.NewMissing("Sender email account")
	}

	userMap, err := s.userService.LoadNotificationUsers(ctx, db, data.ScopeProject,
		viaEmail.ToProjectMembers, viaEmail.ToProjectOwners, viaEmail.ToAllAdmins)
	if err != nil {
		return apperrors.Wrap(err)
	}

	userEmails := make([]string, 0, len(userMap))
	for _, user := range userMap {
		userEmails = append(userEmails, user.Email)
	}
	if len(viaEmail.ToAddresses) > 0 {
		userEmails = gofn.ToSet(append(userEmails, viaEmail.ToAddresses...))
	}
	if len(userEmails) == 0 {
		return nil
	}

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeEmail, data.TemplateName)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf, cleanup := s.getEmailBuildBuf()
	defer cleanup()
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := data.TemplateData.GetTitle()
	if subject == "" {
		subject = gofn.If(data.ActionSucceeded, "Action succeeded", "Action failed")
	}

	err = email.SendMail(ctx, emailAcc, userEmails, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) notifyForTaskResultViaSlack(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
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

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeSlack, data.TemplateName)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf, cleanup := s.getSlackBuildBuf()
	defer cleanup()
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.slackSendMsg(ctx, imService.Slack, strutil.RemoveEmptyLines(buf.String(), false))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) notifyForTaskResultViaDiscord(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
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

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeDiscord, data.TemplateName)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf, cleanup := s.getDiscordBuildBuf()
	defer cleanup()
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.discordSendMsg(ctx, imService.Discord, strutil.RemoveEmptyLines(buf.String(), false))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
