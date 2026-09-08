package auditloguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type UC struct {
	db *database.DB

	auditLogRepo repository.AuditLogRepo
}

func New(
	db *database.DB,

	auditLogRepo repository.AuditLogRepo,
) *UC {
	return &UC{
		db: db,

		auditLogRepo: auditLogRepo,
	}
}
