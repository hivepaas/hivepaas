package appsettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
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

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appDeploymentService appdeploymentservice.Service
	appService           appservice.Service
	clusterService       clusterservice.Service
	domainService        domainservice.Service
	envVarService        envvarservice.Service
	networkService       networkservice.Service
	settingService       settingservice.Service
	sslService           sslservice.Service
	traefikService       traefikservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appDeploymentService appdeploymentservice.Service,
	appService appservice.Service,
	clusterService clusterservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appDeploymentService: appDeploymentService,
		appService:           appService,
		clusterService:       clusterService,
		domainService:        domainService,
		envVarService:        envVarService,
		networkService:       networkService,
		settingService:       settingService,
		sslService:           sslService,
		traefikService:       traefikService,

		dockerManager: dockerManager,
	}
}
