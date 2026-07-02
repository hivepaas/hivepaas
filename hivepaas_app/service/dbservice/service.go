package dbservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	MigrateData(ctx context.Context, db database.IDB) error
}
