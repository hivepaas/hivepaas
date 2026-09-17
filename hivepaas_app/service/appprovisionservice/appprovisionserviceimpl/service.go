package appprovisionserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	projectRepo repository.ProjectRepo,

	appService appservice.Service,
	clusterService clusterservice.Service,
	networkService networkservice.Service,
	placementService placementservice.Service,
) appprovisionservice.Service {
	return &service{
		dockerManager: dockerManager,

		appRepo:     appRepo,
		projectRepo: projectRepo,

		appService:       appService,
		clusterService:   clusterService,
		networkService:   networkService,
		placementService: placementService,
	}
}

type service struct {
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	projectRepo repository.ProjectRepo

	appService       appservice.Service
	clusterService   clusterservice.Service
	networkService   networkservice.Service
	placementService placementservice.Service
}
