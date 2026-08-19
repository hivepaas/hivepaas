package commandserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,

	settingRepo repository.SettingRepo,

	envVarService envvarservice.Service,
	settingService settingservice.Service,
) commandservice.Service {
	return &service{
		dockerManager: dockerManager,

		settingRepo: settingRepo,

		envVarService:  envVarService,
		settingService: settingService,
	}
}

type service struct {
	dockerManager docker.Manager

	settingRepo repository.SettingRepo

	envVarService  envvarservice.Service
	settingService settingservice.Service
}
