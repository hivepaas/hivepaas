package appcontaineruc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db *database.DB

	appService appservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,

	appService appservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		appService: appService,

		dockerManager: dockerManager,
	}
}
