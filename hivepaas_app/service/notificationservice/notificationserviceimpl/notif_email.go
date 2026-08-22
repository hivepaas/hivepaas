package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/email"
)

const (
	emailSendRetryMax   = 2
	emailSendRetryDelay = 3 * time.Second
)

func (s *service) emailSendMsg(
	ctx context.Context,
	emailAcc *entity.Email,
	recipients []string,
	subject string,
	content string,
) (err error) {
	if emailAcc == nil || len(recipients) == 0 {
		return nil
	}

	err = email.SendMailRetry(ctx, emailAcc, recipients, subject, content,
		emailSendRetryMax, emailSendRetryDelay)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
