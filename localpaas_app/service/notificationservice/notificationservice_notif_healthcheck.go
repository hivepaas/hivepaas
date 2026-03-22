package notificationservice

import (
	"bytes"
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/services/email"
)

type BaseMsgDataHealthcheckNotification struct {
	ProjectName     string
	AppName         string
	Succeeded       bool
	HealthcheckName string
	HealthcheckType base.HealthcheckType
	StartedAt       time.Time
	Duration        time.Duration
	Retries         int
	Expect          string
	Actual          string
	DashboardLink   string
}

type EmailMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (s *notificationService) EmailSendHealthcheckNotification(
	ctx context.Context,
	db database.IDB,
	data *EmailMsgDataHealthcheckNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeEmail, TemplateHealthcheckNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeMd))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "Healthcheck notification")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type SlackMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Setting *entity.Slack
}

func (s *notificationService) SlackSendHealthcheckNotification(
	ctx context.Context,
	db database.IDB,
	data *SlackMsgDataHealthcheckNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeSlack, TemplateHealthcheckNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeXs))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.slackSendMsg(ctx, data.Setting, strutil.RemoveEmptyLines(buf.String(), false))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type DiscordMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Setting *entity.Discord
}

func (s *notificationService) DiscordSendHealthcheckNotification(
	ctx context.Context,
	db database.IDB,
	data *DiscordMsgDataHealthcheckNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeDiscord, TemplateHealthcheckNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeXs))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.discordSendMsg(ctx, data.Setting, strutil.RemoveEmptyLines(buf.String(), false))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
