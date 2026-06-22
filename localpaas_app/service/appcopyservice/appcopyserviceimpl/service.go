package appcopyserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appcopyservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/domainservice"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/networkservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	logger      logging.Logger
	db          *database.DB
	redisClient rediscache.Client

	userRepo                 repository.UserRepo
	aclPermissionRepo        repository.ACLPermissionRepo
	projectRepo              repository.ProjectRepo
	projectTagRepo           repository.ProjectTagRepo
	projectSharedSettingRepo repository.ProjectSharedSettingRepo
	appRepo                  repository.AppRepo
	appTagRepo               repository.AppTagRepo
	deploymentRepo           repository.DeploymentRepo
	taskLogRepo              repository.TaskLogRepo
	settingRepo              repository.SettingRepo
	taskRepo                 repository.TaskRepo

	appService     appservice.Service
	settingService settingservice.Service
	sslService     sslservice.Service
	envVarService  envvarservice.Service
	userService    userservice.Service
	networkService networkservice.Service
	domainService  domainservice.Service
	traefikService traefikservice.Service
	taskQueue      queue.TaskQueue
	dockerManager  docker.Manager
}

func New(
	logger logging.Logger,
	db *database.DB,
	redisClient rediscache.Client,
	userRepo repository.UserRepo,
	aclPermissionRepo repository.ACLPermissionRepo,
	projectRepo repository.ProjectRepo,
	projectTagRepo repository.ProjectTagRepo,
	projectSharedSettingRepo repository.ProjectSharedSettingRepo,
	appRepo repository.AppRepo,
	appTagRepo repository.AppTagRepo,
	deploymentRepo repository.DeploymentRepo,
	taskLogRepo repository.TaskLogRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	appService appservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	envVarService envvarservice.Service,
	userService userservice.Service,
	networkService networkservice.Service,
	domainService domainservice.Service,
	traefikService traefikservice.Service,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,
) appcopyservice.Service {
	return &service{
		logger:                   logger,
		db:                       db,
		redisClient:              redisClient,
		userRepo:                 userRepo,
		aclPermissionRepo:        aclPermissionRepo,
		projectRepo:              projectRepo,
		projectTagRepo:           projectTagRepo,
		projectSharedSettingRepo: projectSharedSettingRepo,
		appRepo:                  appRepo,
		appTagRepo:               appTagRepo,
		deploymentRepo:           deploymentRepo,
		taskLogRepo:              taskLogRepo,
		settingRepo:              settingRepo,
		taskRepo:                 taskRepo,
		appService:               appService,
		settingService:           settingService,
		sslService:               sslService,
		envVarService:            envVarService,
		userService:              userService,
		networkService:           networkService,
		domainService:            domainService,
		traefikService:           traefikService,
		taskQueue:                taskQueue,
		dockerManager:            dockerManager,
	}
}
