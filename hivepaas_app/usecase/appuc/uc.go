package appuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db *database.DB

	appRepo       repository.AppRepo
	binObjectRepo repository.BinObjectRepo
	projectRepo   repository.ProjectRepo

	appCloneService      appcloneservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	containerExecService containerexecservice.Service
	networkService       networkservice.Service
	settingService       settingservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	projectRepo repository.ProjectRepo,

	appCloneService appcloneservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		appRepo:       appRepo,
		binObjectRepo: binObjectRepo,
		projectRepo:   projectRepo,

		appCloneService:      appCloneService,
		appService:           appService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		networkService:       networkService,
		settingService:       settingService,

		dockerManager: dockerManager,
	}
}
