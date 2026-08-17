package appcontaineruc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db            *database.DB
	dockerManager docker.Manager

	agentService         agentservice.Service
	appService           appservice.Service
	containerFileService containerfileservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	agentService agentservice.Service,
	appService appservice.Service,
	containerFileService containerfileservice.Service,

) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,

		agentService:         agentService,
		appService:           appService,
		containerFileService: containerFileService,
	}
}
