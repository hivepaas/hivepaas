package emailservice

import (
	"bytes"
	"context"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	GetDefaultSystemEmail(ctx context.Context, db database.IDB) (*entity.Setting, error)
	GetBuildBuf() (buf *bytes.Buffer, cleanup func())

	// User emailing
	SendMailPasswordReset(ctx context.Context, db database.IDB, data *EmailDataPasswordReset) error
	SendMailUserInvite(ctx context.Context, db database.IDB, data *EmailDataUserInvite) error
}
