package volumeuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db *database.DB

	projectService projectservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,

	projectService projectservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		projectService: projectService,

		dockerManager: dockerManager,
	}
}
