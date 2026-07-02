package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bbpool"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/services/email"
)

func (s *service) NotifyForTaskResult(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) (resp *notificationservice.TaskResultNotificationResp, err error) {
	resp = &notificationservice.TaskResultNotificationResp{
		SendTs: timeutil.NowUTC(),
	}
	notification := data.Notification
	if notification == nil {
		return resp, nil
	}

	err = s.loadDefaultNotificationSourceSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	currEvent := gofn.If(data.ActionSucceeded, "success", "failure")
	minSendingInterval := notification.MinSendInterval.ToDuration()
	shouldSkipNotif := minSendingInterval > 0 && data.LastEvent == currEvent &&
		!data.LastSendTs.IsZero() && time.Since(data.LastSendTs) < minSendingInterval

	var execFuncs []func(ctx context.Context) error

	if !shouldSkipNotif && notification.HasNotificationViaEmail() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaEmail(ctx, db, data)
			resp.EmailSent = err == nil
			return err
		})
	}
	if !shouldSkipNotif && notification.HasNotificationViaSlack() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaSlack(ctx, db, data)
			resp.SlackSent = err == nil
			return err
		})
	}
	if !shouldSkipNotif && notification.HasNotificationViaDiscord() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaDiscord(ctx, db, data)
			resp.DiscordSent = err == nil
			return err
		})
	}
	if !shouldSkipNotif && notification.HasNotificationViaTelegram() {
		execFuncs = append(execFuncs, func(ctx context.Context) error {
			err := s.notifyForTaskResultViaTelegram(ctx, db, data)
			resp.TelegramSent = err == nil
			return err
		})
	}
	if len(execFuncs) == 0 {
		return resp, nil
	}

	err = gofn.ExecTasks(ctx, 0, execFuncs...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

//nolint:gocognit
func (s *service) loadDefaultNotificationSourceSettings(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) (err error) {
	notif := data.Notification
	var settingTypes []base.SettingType
	needLoadEmail := false
	if notif.ViaEmail != nil && notif.ViaEmail.UseDefault {
		needLoadEmail = true
		settingTypes = append(settingTypes, base.SettingTypeEmail)
	}
	needLoadSlack := notif.ViaSlack != nil && notif.ViaSlack.UseDefault
	needLoadDiscord := notif.ViaDiscord != nil && notif.ViaDiscord.UseDefault
	needLoadTelegram := notif.ViaTelegram != nil && notif.ViaTelegram.UseDefault
	if needLoadSlack || needLoadDiscord || needLoadTelegram {
		settingTypes = append(settingTypes, base.SettingTypeIMService)
	}

	if len(settingTypes) == 0 {
		return nil
	}

	var scope *base.ObjectScope
	var objectID, parentObjectID string
	switch {
	case data.ScopeApp != nil:
		objectID, parentObjectID = data.ScopeApp.ID, data.ScopeApp.ProjectID
		scope = base.NewObjectScopeApp(objectID, parentObjectID)
	case data.ScopeProject != nil:
		objectID = data.ScopeProject.ID
		scope = base.NewObjectScopeProject(objectID)
	case data.ScopeUser != nil:
		objectID = data.ScopeUser.ID
		scope = base.NewObjectScopeUser(objectID)
	default:
		scope = base.NewObjectScopeGlobal()
	}

	settings, _, err := s.settingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhereIn("setting.type IN (?)", settingTypes...),
		bunex.SelectWhere("setting.is_default IS TRUE"),
	)
	if err != nil {
		return apperrors.New(err)
	}

	findFunc := func(typ base.SettingType, kind *string) *entity.Setting {
		var defaultInParent *entity.Setting
		var defaultInGlobal *entity.Setting
		for _, setting := range settings {
			if setting.Type != typ || (kind != nil && setting.Kind != *kind) {
				continue
			}
			if setting.ObjectID == objectID {
				return setting // found the default setting in current scope
			}
			if setting.ObjectID == parentObjectID {
				defaultInParent = setting
				continue
			}
			if setting.Scope == base.ObjectScopeGlobal {
				defaultInGlobal = setting
				continue
			}
		}
		return gofn.Coalesce(defaultInParent, defaultInGlobal)
	}

	if needLoadEmail {
		if setting := findFunc(base.SettingTypeEmail, nil); setting != nil {
			data.RefObjects.RefSettings[setting.ID] = setting
			notif.ViaEmail.Sender.ID = setting.ID
		}
	}
	if needLoadSlack {
		if setting := findFunc(base.SettingTypeIMService, new(string(base.IMServiceKindSlack))); setting != nil {
			data.RefObjects.RefSettings[setting.ID] = setting
			notif.ViaSlack.Webhook.ID = setting.ID
		}
	}
	if needLoadDiscord {
		if setting := findFunc(base.SettingTypeIMService, new(string(base.IMServiceKindDiscord))); setting != nil {
			data.RefObjects.RefSettings[setting.ID] = setting
			notif.ViaDiscord.Webhook.ID = setting.ID
		}
	}
	if needLoadTelegram {
		if setting := findFunc(base.SettingTypeIMService, new(string(base.IMServiceKindTelegram))); setting != nil {
			data.RefObjects.RefSettings[setting.ID] = setting
			notif.ViaTelegram.Setting.ID = setting.ID
		}
	}

	return nil
}

func (s *service) notifyForTaskResultViaEmail(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
	if notification == nil || notification.ViaEmail == nil || !notification.ViaEmail.Enabled {
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
		return apperrors.New(err)
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
		return apperrors.New(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.New(err)
	}

	subject := data.TemplateData.GetTitle()
	if subject == "" {
		subject = gofn.If(data.ActionSucceeded, "Action succeeded", "Action failed")
	}

	err = email.SendMail(ctx, emailAcc, userEmails, subject, buf.String())
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) notifyForTaskResultViaSlack(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
	if notification == nil || notification.ViaSlack == nil || !notification.ViaSlack.Enabled {
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
		return apperrors.New(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.New(err)
	}

	err = s.slackSendMsg(ctx, imService.Slack, buf.String())
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) notifyForTaskResultViaDiscord(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
	if notification == nil || notification.ViaDiscord == nil || !notification.ViaDiscord.Enabled {
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
		return apperrors.New(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.New(err)
	}

	err = s.discordSendMsg(ctx, imService.Discord, buf.String())
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) notifyForTaskResultViaTelegram(
	ctx context.Context,
	db database.IDB,
	data *notificationservice.TaskResultNotificationReq,
) error {
	notification := data.Notification
	if notification == nil || notification.ViaTelegram == nil || !notification.ViaTelegram.Enabled {
		return nil
	}

	imSetting := data.RefObjects.RefSettings[notification.ViaTelegram.Setting.ID]
	if imSetting == nil {
		return apperrors.NewMissing("Telegram configuration")
	}
	imService := imSetting.MustAsIMService()
	if imService == nil || imService.Telegram == nil {
		return apperrors.NewMissing("Telegram configuration")
	}

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeTelegram, data.TemplateName)
	if err != nil {
		return apperrors.New(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, data.TemplateData)
	if err != nil {
		return apperrors.New(err)
	}

	err = s.telegramSendMsg(ctx, imService.Telegram, buf.String())
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
