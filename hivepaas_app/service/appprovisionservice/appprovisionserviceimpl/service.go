package appprovisionserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	projectRepo repository.ProjectRepo,

	appDeploymentService appdeploymentservice.Service,
	appRoutingService approutingservice.Service,
	appService appservice.Service,
	clusterSecretService clustersecretservice.Service,
	clusterService clusterservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,
	placementService placementservice.Service,
) appprovisionservice.Service {
	return &service{
		taskQueue:     taskQueue,
		dockerManager: dockerManager,

		appRepo:     appRepo,
		projectRepo: projectRepo,

		appDeploymentService: appDeploymentService,
		appRoutingService:    appRoutingService,
		appService:           appService,
		clusterSecretService: clusterSecretService,
		clusterService:       clusterService,
		domainService:        domainService,
		envVarService:        envVarService,
		networkService:       networkService,
		placementService:     placementService,
	}
}

type service struct {
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	projectRepo repository.ProjectRepo

	appDeploymentService appdeploymentservice.Service
	appRoutingService    approutingservice.Service
	appService           appservice.Service
	clusterSecretService clustersecretservice.Service
	clusterService       clusterservice.Service
	domainService        domainservice.Service
	envVarService        envvarservice.Service
	networkService       networkservice.Service
	placementService     placementservice.Service
}
