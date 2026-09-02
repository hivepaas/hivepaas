package backuprepocleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backuprepocleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
)

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo

	backupRepoService backupreposervice.Service
	scopeService      scopeservice.Service
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	backupRepoService backupreposervice.Service,
	scopeService scopeservice.Service,
) backuprepocleanupservice.Service {
	return &service{
		db: db,

		settingRepo: settingRepo,

		backupRepoService: backupRepoService,
		scopeService:      scopeService,
	}
}
