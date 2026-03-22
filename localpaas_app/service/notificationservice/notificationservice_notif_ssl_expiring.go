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

type BaseMsgDataSSLExpiringNotification struct {
	ProjectName   string
	AppName       string
	SSLName       string
	SSLType       string
	Domain        string
	CreatedAt     time.Time
	ExpireAt      time.Time
	ExpireIn      timeutil.Duration
	DashboardLink string
}

type EmailMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (s *notificationService) EmailSendSSLExpiringNotification(
	ctx context.Context,
	db database.IDB,
	data *EmailMsgDataSSLExpiringNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeEmail, TemplateSSLExpiringNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeMd))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "Your SSL is expiring")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type SlackMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Setting *entity.Slack
}

func (s *notificationService) SlackSendSSLExpiringNotification(
	ctx context.Context,
	db database.IDB,
	data *SlackMsgDataSSLExpiringNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeSlack, TemplateSSLExpiringNotification)
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

type DiscordMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Setting *entity.Discord
}

func (s *notificationService) DiscordSendSSLExpiringNotification(
	ctx context.Context,
	db database.IDB,
	data *DiscordMsgDataSSLExpiringNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeDiscord, TemplateSSLExpiringNotification)
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
