package startupserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/startupservice"
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
