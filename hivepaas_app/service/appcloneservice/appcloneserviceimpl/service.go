package appcloneserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apphttpservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandpipeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appHttpService         apphttpservice.Service
	appService             appservice.Service
	clusterSecretService   clustersecretservice.Service
	clusterService         clusterservice.Service
	commandPipeExecService commandpipeexecservice.Service
	domainService          domainservice.Service
	envVarService          envvarservice.Service
	networkService         networkservice.Service
	settingService         settingservice.Service
	sslService             sslservice.Service
	traefikService         traefikservice.Service
	volumeService          volumeservice.Service
}

func New(
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appHttpService apphttpservice.Service,
	appService appservice.Service,
	clusterSecretService clustersecretservice.Service,
	clusterService clusterservice.Service,
	commandPipeExecService commandpipeexecservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,
	volumeService volumeservice.Service,
) appcloneservice.Service {
	return &service{
		taskQueue:     taskQueue,
		dockerManager: dockerManager,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appHttpService:         appHttpService,
		appService:             appService,
		clusterSecretService:   clusterSecretService,
		clusterService:         clusterService,
		commandPipeExecService: commandPipeExecService,
		domainService:          domainService,
		envVarService:          envVarService,
		networkService:         networkService,
		settingService:         settingService,
		sslService:             sslService,
		traefikService:         traefikService,
		volumeService:          volumeService,
	}
}
