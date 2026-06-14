package builduc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db                *database.DB
	sysCleanupService syscleanupservice.Service
	dockerManager     docker.Manager
}

func New(
	db *database.DB,
	sysCleanupService syscleanupservice.Service,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:                db,
		sysCleanupService: sysCleanupService,
		dockerManager:     dockerManager,
	}
}
