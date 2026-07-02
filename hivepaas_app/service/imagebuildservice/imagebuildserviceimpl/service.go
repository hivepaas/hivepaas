package imagebuildserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	settingRepo repository.SettingRepo

	envVarService envvarservice.Service

	dockerManager docker.Manager
}

func New(
	settingRepo repository.SettingRepo,

	envVarService envvarservice.Service,

	dockerManager docker.Manager,
) imagebuildservice.Service {
	return &service{
		settingRepo: settingRepo,

		envVarService: envVarService,

		dockerManager: dockerManager,
	}
}
