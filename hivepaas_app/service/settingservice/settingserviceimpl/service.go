package settingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	projectService projectservice.Service,
	sslService sslservice.Service,
	userService userservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) settingservice.Service {
	return &service{
		appRepo:              appRepo,
		periodicSettingsRepo: periodicSettingsRepo,
		projectEnvRepo:       projectEnvRepo,
		projectRepo:          projectRepo,
		settingRepo:          settingRepo,

		appService:     appService,
		projectService: projectService,
		sslService:     sslService,
		userService:    userService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	appRepo              repository.AppRepo
	periodicSettingsRepo cacherepository.PeriodicSettingsRepo
	projectEnvRepo       repository.ProjectEnvRepo
	projectRepo          repository.ProjectRepo
	settingRepo          repository.SettingRepo

	appService     appservice.Service
	projectService projectservice.Service
	sslService     sslservice.Service
	userService    userservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
