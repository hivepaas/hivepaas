package appuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
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

	auditService         auditservice.Service
	appCloneService      appcloneservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	containerExecService containerexecservice.Service
	networkService       networkservice.Service
	placementService     placementservice.Service
	settingService       settingservice.Service
	loggingService       loggingservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	auditService auditservice.Service,
	appCloneService appcloneservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	networkService networkservice.Service,
	placementService placementservice.Service,
	settingService settingservice.Service,
	loggingService loggingservice.Service,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,

		appRepo:       appRepo,
		binObjectRepo: binObjectRepo,
		projectRepo:   projectRepo,
		settingRepo:   settingRepo,

		auditService:         auditService,
		appCloneService:      appCloneService,
		appService:           appService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		networkService:       networkService,
		placementService:     placementService,
		settingService:       settingService,
		loggingService:       loggingService,
	}
}
