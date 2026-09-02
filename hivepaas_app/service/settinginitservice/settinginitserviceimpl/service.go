package settinginitserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func New(
	lockRepo repository.LockRepo,
	settingRepo repository.SettingRepo,

	sslService sslservice.Service,

	dockerManager docker.Manager,
) settinginitservice.Service {
	return &service{
		lockRepo:    lockRepo,
		settingRepo: settingRepo,

		sslService: sslService,

		dockerManager: dockerManager,
	}
}

type service struct {
	lockRepo    repository.LockRepo
	settingRepo repository.SettingRepo

	sslService sslservice.Service

	dockerManager docker.Manager
}
