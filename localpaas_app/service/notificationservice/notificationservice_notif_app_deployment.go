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

const (
	buffSizeXs = 1000
	buffSizeMd = 5000
)

type BaseMsgDataAppDeploymentNotification struct {
	ProjectName   string
	AppName       string
	Succeeded     bool
	Method        base.DeploymentMethod
	RepoURL       string
	RepoRef       string
	CommitMsg     string
	Image         string
	StartedAt     time.Time
	Duration      time.Duration
	DashboardLink string
}

type EmailMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (s *notificationService) EmailSendAppDeploymentNotification(
	ctx context.Context,
	db database.IDB,
	data *EmailMsgDataAppDeploymentNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeEmail, TemplateAppDeploymentNotification)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, buffSizeMd))
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "Deployment notification")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type SlackMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Setting *entity.Slack
}

func (s *notificationService) SlackSendAppDeploymentNotification(
	ctx context.Context,
	db database.IDB,
	data *SlackMsgDataAppDeploymentNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeSlack, TemplateAppDeploymentNotification)
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

type DiscordMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Setting *entity.Discord
}

func (s *notificationService) DiscordSendAppDeploymentNotification(
	ctx context.Context,
	db database.IDB,
	data *DiscordMsgDataAppDeploymentNotification,
) error {
	template, err := s.GetTemplate(ctx, db, TemplateTypeDiscord, TemplateAppDeploymentNotification)
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
