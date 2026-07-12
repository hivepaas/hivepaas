package email

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/email/http"
	"github.com/hivepaas/hivepaas/services/email/smtp"
)

func SendMail(
	ctx context.Context,
	email *entity.Email,
	recipients []string,
	subject string,
	content string,
) (err error) {
	switch { //nolint
	case email.SMTP != nil:
		err = smtp.SendMail(ctx, email.SMTP, recipients, subject, content)
	case email.HTTP != nil:
		err = http.SendMail(ctx, email.HTTP, recipients, subject, content)
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
