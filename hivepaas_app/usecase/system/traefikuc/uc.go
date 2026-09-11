package traefikuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

type UC struct {
	db *database.DB

	auditService   auditservice.Service
	traefikService traefikservice.Service
}

func New(
	db *database.DB,

	auditService auditservice.Service,
	traefikService traefikservice.Service,
) *UC {
	return &UC{
		db: db,

		auditService:   auditService,
		traefikService: traefikService,
	}
}
