package backupsnapshotuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// UC lists, reads and deletes the snapshots a scope's view reaches.
type UC struct {
	*settings.BaseUC

	appRepo           repository.AppRepo
	backupRepoService backupreposervice.Service
}

func New(
	baseUC *settings.BaseUC,
	appRepo repository.AppRepo,
	backupRepoService backupreposervice.Service,
) *UC {
	return &UC{
		BaseUC:            baseUC,
		appRepo:           appRepo,
		backupRepoService: backupRepoService,
	}
}
