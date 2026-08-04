package volumeserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	dockerManager docker.Manager,
	hpAppService hpappservice.Service,

	settingRepo repository.SettingRepo,
) volumeservice.Service {
	return &service{
		dockerManager: dockerManager,
		hpAppService:  hpAppService,

		settingRepo: settingRepo,
	}
}

type service struct {
	dockerManager docker.Manager
	hpAppService  hpappservice.Service

	settingRepo repository.SettingRepo
}
