package settingserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func New(
	appRepo repository.AppRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,

	userService userservice.Service,

	permissionManager permission.Manager,
) settingservice.Service {
	return &service{
		appRepo:        appRepo,
		projectEnvRepo: projectEnvRepo,
		projectRepo:    projectRepo,
		settingRepo:    settingRepo,

		userService: userService,

		permissionManager: permissionManager,
	}
}

type service struct {
	appRepo        repository.AppRepo
	projectEnvRepo repository.ProjectEnvRepo
	projectRepo    repository.ProjectRepo
	settingRepo    repository.SettingRepo

	userService userservice.Service

	permissionManager permission.Manager
}
