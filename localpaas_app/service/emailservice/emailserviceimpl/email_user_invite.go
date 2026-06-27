package emailserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bbpool"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/services/email"
)

func (s *service) SendMailUserInvite(
	ctx context.Context,
	db database.IDB,
	data *emailservice.EmailDataUserInvite,
) error {
	template, err := s.GetTemplate(ctx, db, emailservice.TemplateNameUserInvite)
	if err != nil {
		return apperrors.New(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, *data)
	if err != nil {
		return apperrors.New(err)
	}

	subject := gofn.Coalesce(data.Subject, "You’ve been invited to join LocalPaaS")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
