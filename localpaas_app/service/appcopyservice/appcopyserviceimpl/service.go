package appcopyserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appcopyservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/domainservice"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	taskQueue queue.TaskQueue

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	clusterService clusterservice.Service
	domainService  domainservice.Service
	envVarService  envvarservice.Service
	networkService networkservice.Service
	settingService settingservice.Service
	sslService     sslservice.Service
	traefikService traefikservice.Service

	dockerManager docker.Manager
}

func New(
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	clusterService clusterservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) appcopyservice.Service {
	return &service{
		taskQueue: taskQueue,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		clusterService: clusterService,
		domainService:  domainService,
		envVarService:  envVarService,
		networkService: networkService,
		settingService: settingService,
		sslService:     sslService,
		traefikService: traefikService,

		dockerManager: dockerManager,
	}
}
