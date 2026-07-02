package emailservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	GetDefaultSystemEmail(ctx context.Context, db database.IDB) (*entity.Setting, error)

	// User emailing
	SendMailPasswordReset(ctx context.Context, db database.IDB, data *EmailDataPasswordReset) error
	SendMailUserInvite(ctx context.Context, db database.IDB, data *EmailDataUserInvite) error
}
