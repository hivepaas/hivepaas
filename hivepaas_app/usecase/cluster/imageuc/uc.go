package imageuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db *database.DB

	settingRepo repository.SettingRepo

	dockerManager docker.Manager
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db: db,

		settingRepo: settingRepo,

		dockerManager: dockerManager,
	}
}
