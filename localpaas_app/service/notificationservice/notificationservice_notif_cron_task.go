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
	"github.com/localpaas/localpaas/services/email"
)

type BaseMsgDataCronTaskNotification struct {
	ProjectName   string
	AppName       string
	Succeeded     bool
	CronJobName   string
	Schedule      string
	CreatedAt     time.Time
	StartedAt     time.Time
	Duration      time.Duration
	Retries       int
	DashboardLink string
}

type EmailMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (s *notificationService) EmailSendCronTaskNotification(
	ctx context.Context,
	db database.IDB,
	data *EmailMsgDataCronTaskNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeEmail, TemplateCronTaskNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeMd))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "Scheduled task notification")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type SlackMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Setting *entity.Slack
}

func (s *notificationService) SlackSendCronTaskNotification(
	ctx context.Context,
	db database.IDB,
	data *SlackMsgDataCronTaskNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeSlack, TemplateCronTaskNotification)
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

type DiscordMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Setting *entity.Discord
}

func (s *notificationService) DiscordSendCronTaskNotification(
	ctx context.Context,
	db database.IDB,
	data *DiscordMsgDataCronTaskNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeDiscord, TemplateCronTaskNotification)
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
