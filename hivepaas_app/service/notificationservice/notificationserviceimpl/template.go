package notificationserviceimpl

import (
	"context"
	htmltemplate "html/template"
	"io"
	"sync"
	texttemplate "text/template"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

const (
	emailTemplateDir    = "email/templates/" // NOTE: must end with /
	slackTemplateDir    = "slack/templates/"
	discordTemplateDir  = "discord/templates/"
	telegramTemplateDir = "telegram/templates/"
	larkTemplateDir     = "lark/templates/"
)

type Template interface {
	Execute(wr io.Writer, data any) error
}

var (
	templateMap = map[notificationservice.TemplateType]map[notificationservice.TemplateName]Template{}
	mu          sync.Mutex
)

func (s *service) GetTemplate(
	ctx context.Context,
	db database.IDB,
	typ notificationservice.TemplateType,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	mu.Lock()
	defer mu.Unlock()

	mapTplByName, exists := templateMap[typ]
	if !exists {
		mapTplByName = make(map[notificationservice.TemplateName]Template, 5) //nolint:mnd
		templateMap[typ] = mapTplByName
	}

	tpl, exists = mapTplByName[name]
	if exists {
		return tpl, nil
	}

	switch typ {
	case notificationservice.TemplateTypeEmail:
		tpl, err = s.loadEmailTemplate(ctx, db, name)
	case notificationservice.TemplateTypeSlack:
		tpl, err = s.loadSlackTemplate(ctx, db, name)
	case notificationservice.TemplateTypeDiscord:
		tpl, err = s.loadDiscordTemplate(ctx, db, name)
	case notificationservice.TemplateTypeTelegram:
		tpl, err = s.loadTelegramTemplate(ctx, db, name)
	case notificationservice.TemplateTypeLark:
		tpl, err = s.loadLarkTemplate(ctx, db, name)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	mapTplByName[name] = tpl

	return tpl, nil
}

func (s *service) loadEmailTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	switch name {
	case notificationservice.TemplateAppDeploymentNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"app_deployment_notification.html")
	case notificationservice.TemplateSchedTaskNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"sched_task_notification.html")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"healthcheck_notification.html")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"ssl_expiring_notification.html")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"ssl_renewal_notification.html")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), emailTemplateDir+"system_update_notification.html")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}

func (s *service) loadSlackTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	switch name {
	case notificationservice.TemplateAppDeploymentNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"app_deployment_notification.tpl")
	case notificationservice.TemplateSchedTaskNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"sched_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), slackTemplateDir+"system_update_notification.tpl")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}

func (s *service) loadDiscordTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	switch name {
	case notificationservice.TemplateAppDeploymentNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"app_deployment_notification.tpl")
	case notificationservice.TemplateSchedTaskNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"sched_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), discordTemplateDir+"system_update_notification.tpl")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}

func (s *service) loadTelegramTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	switch name {
	case notificationservice.TemplateAppDeploymentNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"app_deployment_notification.tpl")
	case notificationservice.TemplateSchedTaskNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"sched_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), telegramTemplateDir+"system_update_notification.tpl")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}

func (s *service) loadLarkTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	switch name {
	case notificationservice.TemplateAppDeploymentNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"app_deployment_notification.tpl")
	case notificationservice.TemplateSchedTaskNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"sched_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = texttemplate.ParseFS(assets.GetTemplatesFS(), larkTemplateDir+"system_update_notification.tpl")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}
