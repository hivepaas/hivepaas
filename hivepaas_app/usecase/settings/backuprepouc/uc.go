package backuprepouc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeBackupRepo
	currentSettingVersion = entity.CurrentBackupRepoVersion
)

type UC struct {
	*settings.BaseUC

	backupRepoService backupreposervice.Service
}

func New(
	baseUC *settings.BaseUC,

	backupRepoService backupreposervice.Service,
) *UC {
	return &UC{
		BaseUC: baseUC,

		backupRepoService: backupRepoService,
	}
}
