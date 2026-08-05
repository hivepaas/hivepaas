package clustersecretserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustersecretservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	dockerManager docker.Manager,
) clustersecretservice.Service {
	return &service{
		appRepo:     appRepo,
		settingRepo: settingRepo,

		dockerManager: dockerManager,
	}
}

type service struct {
	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	dockerManager docker.Manager
}
