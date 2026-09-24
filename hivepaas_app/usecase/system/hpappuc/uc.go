package hpappuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

type UC struct {
	db *database.DB

	lockRepo repository.LockRepo

	auditService     auditservice.Service
	hpAppService     hpappservice.Service
	sysUpdateService sysupdateservice.Service
}

func New(
	db *database.DB,

	lockRepo repository.LockRepo,

	auditService auditservice.Service,
	hpAppService hpappservice.Service,
	sysUpdateService sysupdateservice.Service,
) *UC {
	return &UC{
		db: db,

		lockRepo: lockRepo,

		auditService:     auditService,
		hpAppService:     hpAppService,
		sysUpdateService: sysUpdateService,
	}
}
