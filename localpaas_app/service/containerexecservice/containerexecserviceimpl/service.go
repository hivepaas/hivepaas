package containerexecserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
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
	sysErrorRepo             repository.SysErrorRepo
	loginTrustedDeviceRepo   repository.LoginTrustedDeviceRepo

	schedJobService     schedjobservice.Service
	appService          appservice.Service
	settingService      settingservice.Service
	sslService          sslservice.Service
	userService         userservice.Service
	notificationService notificationservice.Service
	traefikService      traefikservice.Service
	dockerManager       docker.Manager
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
	sysErrorRepo repository.SysErrorRepo,
	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo,
	schedJobService schedjobservice.Service,
	appService appservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	userService userservice.Service,
	notificationService notificationservice.Service,
	traefikService traefikservice.Service,
	dockerManager docker.Manager,
) containerexecservice.Service {
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
		sysErrorRepo:             sysErrorRepo,
		loginTrustedDeviceRepo:   loginTrustedDeviceRepo,
		schedJobService:          schedJobService,
		appService:               appService,
		settingService:           settingService,
		sslService:               sslService,
		userService:              userService,
		notificationService:      notificationService,
		traefikService:           traefikService,
		dockerManager:            dockerManager,
	}
}
