package apppreviewserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type service struct {
	db          *database.DB
	redisClient rediscache.Client
	taskQueue   queue.TaskQueue

	appRepo        repository.AppRepo
	deploymentRepo repository.DeploymentRepo
	resLinkRepo    repository.ResLinkRepo
	settingRepo    repository.SettingRepo
	taskLogRepo    repository.TaskLogRepo
	taskRepo       repository.TaskRepo

	appCloneService      appcloneservice.Service
	appDeploymentService appdeploymentservice.Service
	appService           appservice.Service
	commandService       commandservice.Service
	containerExecService containerexecservice.Service
	domainService        domainservice.Service
	envVarService        envvarservice.Service
	settingService       settingservice.Service
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	deploymentRepo repository.DeploymentRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,

	appCloneService appcloneservice.Service,
	appDeploymentService appdeploymentservice.Service,
	appService appservice.Service,
	commandService commandservice.Service,
	containerExecService containerexecservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
	settingService settingservice.Service,
) apppreviewservice.Service {
	return &service{
		db:          db,
		redisClient: redisClient,
		taskQueue:   taskQueue,

		appRepo:        appRepo,
		deploymentRepo: deploymentRepo,
		resLinkRepo:    resLinkRepo,
		settingRepo:    settingRepo,
		taskLogRepo:    taskLogRepo,
		taskRepo:       taskRepo,

		appCloneService:      appCloneService,
		appDeploymentService: appDeploymentService,
		appService:           appService,
		commandService:       commandService,
		containerExecService: containerExecService,
		domainService:        domainService,
		envVarService:        envVarService,
		settingService:       settingService,
	}
}
