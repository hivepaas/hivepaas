package backupsnapshotuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// UC lists, reads, deletes and restores the snapshots a scope's view reaches.
type UC struct {
	*settings.BaseUC

	appRepo           repository.AppRepo
	taskRepo          repository.TaskRepo
	taskQueue         queue.TaskQueue
	backupRepoService backupreposervice.Service
	dataBackupService databackupservice.Service
}

func New(
	baseUC *settings.BaseUC,
	appRepo repository.AppRepo,
	taskRepo repository.TaskRepo,
	taskQueue queue.TaskQueue,
	backupRepoService backupreposervice.Service,
	dataBackupService databackupservice.Service,
) *UC {
	return &UC{
		BaseUC:            baseUC,
		appRepo:           appRepo,
		taskRepo:          taskRepo,
		taskQueue:         taskQueue,
		backupRepoService: backupRepoService,
		dataBackupService: dataBackupService,
	}
}
