package startupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/startupservice"
)

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,
) startupservice.Service {
	return &service{
		db: db,

		settingRepo: settingRepo,
	}
}

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo
}
