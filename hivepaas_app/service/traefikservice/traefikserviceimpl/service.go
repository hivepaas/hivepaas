package traefikserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
) traefikservice.Service {
	return &service{
		dockerManager: dockerManager,
		settingRepo:   settingRepo,
	}
}

type service struct {
	dockerManager docker.Manager
	settingRepo   repository.SettingRepo
}
