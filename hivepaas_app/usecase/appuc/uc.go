package appuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db            *database.DB
	dockerManager docker.Manager

	appRepo       repository.AppRepo
	binObjectRepo repository.BinObjectRepo
	projectRepo   repository.ProjectRepo
	settingRepo   repository.SettingRepo

	appCloneService      appcloneservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	containerExecService containerexecservice.Service
	networkService       networkservice.Service
	placementService     placementservice.Service
	settingService       settingservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	appCloneService appcloneservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	networkService networkservice.Service,
	placementService placementservice.Service,
	settingService settingservice.Service,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,

		appRepo:       appRepo,
		binObjectRepo: binObjectRepo,
		projectRepo:   projectRepo,
		settingRepo:   settingRepo,

		appCloneService:      appCloneService,
		appService:           appService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		networkService:       networkService,
		placementService:     placementService,
		settingService:       settingService,
	}
}
