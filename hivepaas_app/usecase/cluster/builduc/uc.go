package builduc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
)

type UC struct {
	db *database.DB

	auditService      auditservice.Service
	sysCleanupService syscleanupservice.Service
}

func New(
	db *database.DB,

	auditService auditservice.Service,
	sysCleanupService syscleanupservice.Service,
) *UC {
	return &UC{
		db: db,

		auditService:      auditService,
		sysCleanupService: sysCleanupService,
	}
}
