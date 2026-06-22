package appuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appcopyservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/clusterservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/reslinkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db                   *database.DB
	projectRepo          repository.ProjectRepo
	appRepo              repository.AppRepo
	settingRepo          repository.SettingRepo
	deploymentRepo       repository.DeploymentRepo
	userService          userservice.Service
	appService           appservice.Service
	appCopyService       appcopyservice.Service
	settingService       settingservice.Service
	resLinkService       reslinkservice.Service
	sslService           sslservice.Service
	projectService       projectservice.Service
	networkService       networkservice.Service
	envVarService        envvarservice.Service
	traefikService       traefikservice.Service
	clusterService       clusterservice.Service
	containerExecService containerexecservice.Service
	dockerManager        docker.Manager
	taskQueue            queue.TaskQueue
}

func New(
	db *database.DB,
	projectRepo repository.ProjectRepo,
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
	deploymentRepo repository.DeploymentRepo,
	userService userservice.Service,
	appService appservice.Service,
	appCopyService appcopyservice.Service,
	settingService settingservice.Service,
	resLinkService reslinkservice.Service,
	sslService sslservice.Service,
	projectService projectservice.Service,
	networkService networkservice.Service,
	envVarService envvarservice.Service,
	traefikService traefikservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	dockerManager docker.Manager,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		db:                   db,
		projectRepo:          projectRepo,
		appRepo:              appRepo,
		settingRepo:          settingRepo,
		deploymentRepo:       deploymentRepo,
		userService:          userService,
		appService:           appService,
		appCopyService:       appCopyService,
		settingService:       settingService,
		resLinkService:       resLinkService,
		sslService:           sslService,
		projectService:       projectService,
		networkService:       networkService,
		envVarService:        envVarService,
		traefikService:       traefikService,
		clusterService:       clusterService,
		containerExecService: containerExecService,
		dockerManager:        dockerManager,
		taskQueue:            taskQueue,
	}
}
