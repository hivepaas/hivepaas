package builduc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
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
