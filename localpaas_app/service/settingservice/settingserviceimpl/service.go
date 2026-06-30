package settingserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
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
