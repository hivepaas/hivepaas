package builduc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
)

type UC struct {
	db *database.DB

	sysCleanupService syscleanupservice.Service
}

func New(
	db *database.DB,

	sysCleanupService syscleanupservice.Service,
) *UC {
	return &UC{
		db: db,

		sysCleanupService: sysCleanupService,
	}
}
