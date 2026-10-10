package auditloguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

type UC struct {
	db *database.DB

	auditLogRepo repository.AuditLogRepo

	scopeService   scopeservice.Service
	settingService settingservice.Service
}

func New(
	db *database.DB,

	auditLogRepo repository.AuditLogRepo,

	scopeService scopeservice.Service,
	settingService settingservice.Service,
) *UC {
	return &UC{
		db: db,

		auditLogRepo: auditLogRepo,

		scopeService:   scopeService,
		settingService: settingService,
	}
}
