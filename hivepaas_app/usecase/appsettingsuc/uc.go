package appsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
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

	permissionManager permission.Manager
	dockerManager     docker.Manager
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

	permissionManager permission.Manager,
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

		permissionManager: permissionManager,
		dockerManager:     dockerManager,
	}
}
