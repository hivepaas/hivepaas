package imagebuildserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
	"github.com/localpaas/localpaas/localpaas_app/service/imagebuildservice"
	"github.com/localpaas/localpaas/services/docker"
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
