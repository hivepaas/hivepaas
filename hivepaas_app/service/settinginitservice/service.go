package settinginitservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	InitDefaults(ctx context.Context, db database.IDB) error
	InitDefaultsWithTx(ctx context.Context, db database.Tx) error
}
