package emailserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bbpool"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/services/email"
)

func (s *service) SendMailUserInvite(
	ctx context.Context,
	db database.IDB,
	data *emailservice.EmailDataUserInvite,
) error {
	template, err := s.GetTemplate(ctx, db, emailservice.TemplateNameUserInvite)
	if err != nil {
		return hperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, *data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	subject := gofn.Coalesce(data.Subject, "You’ve been invited to join HivePaaS")
	err = email.SendMail(ctx, data.Email, data.Recipients, subject, buf.String())
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
