package settingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,

	appService appservice.Service,
	projectService projectservice.Service,
	systemEventBus systemeventbusservice.Service,
	userService userservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) settingservice.Service {
	return &service{
		appRepo:              appRepo,
		projectEnvRepo:       projectEnvRepo,
		projectRepo:          projectRepo,
		settingRepo:          settingRepo,
		periodicSettingsRepo: periodicSettingsRepo,

		appService:     appService,
		projectService: projectService,
		systemEventBus: systemEventBus,
		userService:    userService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	appRepo              repository.AppRepo
	projectEnvRepo       repository.ProjectEnvRepo
	projectRepo          repository.ProjectRepo
	settingRepo          repository.SettingRepo
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo

	appService     appservice.Service
	projectService projectservice.Service
	systemEventBus systemeventbusservice.Service
	userService    userservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
