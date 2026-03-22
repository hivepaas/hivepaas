package notificationservice

import (
	"bytes"
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/email"
)

type BaseMsgDataSSLRenewalNotification struct {
	ProjectName   string
	AppName       string
	Succeeded     bool
	SSLName       string
	SSLType       string
	Domain        string
	CreatedAt     time.Time
	ExpireAt      time.Time
	NextRenewalIn timeutil.Duration
	DashboardLink string
}

type EmailMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (s *notificationService) EmailSendSSLRenewalNotification(
	ctx context.Context,
	db database.IDB,
	data *EmailMsgDataSSLRenewalNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeEmail, TemplateSSLRenewalNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeMd))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "SSL renewal notification")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type SlackMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Setting *entity.Slack
}

func (s *notificationService) SlackSendSSLRenewalNotification(
	ctx context.Context,
	db database.IDB,
	data *SlackMsgDataSSLRenewalNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeSlack, TemplateSSLRenewalNotification)
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

type DiscordMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Setting *entity.Discord
}

func (s *notificationService) DiscordSendSSLRenewalNotification(
	ctx context.Context,
	db database.IDB,
	data *DiscordMsgDataSSLRenewalNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeDiscord, TemplateSSLRenewalNotification)
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
