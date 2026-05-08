package notificationserviceimpl

import (
	"context"
	htmltemplate "html/template"
	"io"
	"sync"
	texttemplate "text/template"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
)

const (
	emailTemplateDir   = "config/email/templates/" // NOTE: must end with /
	slackTemplateDir   = "config/slack/templates/"
	discordTemplateDir = "config/discord/templates/"
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
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "app_deployment_notification.html")
	case notificationservice.TemplateCronTaskNotification:
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "cron_task_notification.html")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "healthcheck_notification.html")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "ssl_expiring_notification.html")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "ssl_renewal_notification.html")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = htmltemplate.ParseFiles(emailTemplateDir + "system_update_notification.html")
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
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "app_deployment_notification.tpl")
	case notificationservice.TemplateCronTaskNotification:
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "cron_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = texttemplate.ParseFiles(slackTemplateDir + "system_update_notification.tpl")
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
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "app_deployment_notification.tpl")
	case notificationservice.TemplateCronTaskNotification:
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "cron_task_notification.tpl")
	case notificationservice.TemplateHealthcheckNotification:
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "healthcheck_notification.tpl")
	case notificationservice.TemplateSSLExpiringNotification:
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "ssl_expiring_notification.tpl")
	case notificationservice.TemplateSSLRenewalNotification:
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "ssl_renewal_notification.tpl")
	case notificationservice.TemplateSystemUpdateNotification:
		tpl, err = texttemplate.ParseFiles(discordTemplateDir + "system_update_notification.tpl")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return tpl, nil
}
