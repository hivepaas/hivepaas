package appuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appcopyservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db *database.DB

	appRepo     repository.AppRepo
	projectRepo repository.ProjectRepo

	appCopyService       appcopyservice.Service
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
	projectRepo repository.ProjectRepo,

	appCopyService appcopyservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		appRepo:     appRepo,
		projectRepo: projectRepo,

		appCopyService:       appCopyService,
		appService:           appService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		networkService:       networkService,
		settingService:       settingService,

		dockerManager: dockerManager,
	}
}
