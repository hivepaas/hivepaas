package dbserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dbservice"
)

func New(
	dataMigrationRepo repository.DataMigrationRepo,
	settingRepo repository.SettingRepo,
) dbservice.Service {
	return &service{
		dataMigrationRepo: dataMigrationRepo,
		settingRepo:       settingRepo,
	}
}

type service struct {
	dataMigrationRepo repository.DataMigrationRepo
	settingRepo       repository.SettingRepo
}
