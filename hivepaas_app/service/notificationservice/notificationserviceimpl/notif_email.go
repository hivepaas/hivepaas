package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bbpool"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/services/email"
)

const (
	emailSendRetryMax   = 2
	emailSendRetryDelay = 3 * time.Second
)

func (s *service) emailSendMsg(
	ctx context.Context,
	db database.IDB,
	emailSetting *entity.Setting,
	recipients []string,
	subject string,
	templateName notificationservice.TemplateName,
	templateData notificationservice.TemplateData,
) (err error) {
	if len(recipients) == 0 {
		return nil
	}
	if emailSetting == nil {
		return hperrors.NewMissing("Sender email account")
	}
	emailAcc := emailSetting.MustAsEmail()
	if emailAcc == nil {
		return hperrors.NewMissing("Sender email account")
	}

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeEmail, templateName)
	if err != nil {
		return hperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, templateData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = email.SendMailRetry(ctx, emailAcc, recipients, subject, buf.String(),
		emailSendRetryMax, emailSendRetryDelay)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
