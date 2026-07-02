package nodeuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db *database.DB

	settingRepo repository.SettingRepo

	hpAppService hpappservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	hpAppService hpappservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		settingRepo: settingRepo,

		hpAppService: hpAppService,

		dockerManager: dockerManager,
	}
}
