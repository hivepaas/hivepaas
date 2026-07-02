package settingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	healthcheckSettingsRepo cacherepository.HealthcheckSettingsRepo,
	settingRepo repository.SettingRepo,

	sslService sslservice.Service,
	userService userservice.Service,

	dockerManager docker.Manager,
	permissionManager permission.Manager,
) settingservice.Service {
	return &service{
		appRepo:                 appRepo,
		healthcheckSettingsRepo: healthcheckSettingsRepo,
		settingRepo:             settingRepo,

		sslService:  sslService,
		userService: userService,

		dockerManager:     dockerManager,
		permissionManager: permissionManager,
	}
}

type service struct {
	appRepo                 repository.AppRepo
	healthcheckSettingsRepo cacherepository.HealthcheckSettingsRepo
	settingRepo             repository.SettingRepo

	sslService  sslservice.Service
	userService userservice.Service

	dockerManager     docker.Manager
	permissionManager permission.Manager
}
