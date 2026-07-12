package emailserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bbpool"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/services/email"
)

func (s *service) SendMailPasswordReset(
	ctx context.Context,
	db database.IDB,
	data *emailservice.EmailDataPasswordReset,
) error {
	template, err := s.GetTemplate(ctx, db, emailservice.TemplateNamePasswordReset)
	if err != nil {
		return apperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "[HivePaaS] Password reset")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
