package networkserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
) networkservice.Service {
	return &service{
		dockerManager: dockerManager,

		settingRepo: settingRepo,
	}
}

type service struct {
	dockerManager docker.Manager

	settingRepo repository.SettingRepo
}
