// Package imagescanuc is image scanning's system settings: the addon, its
// daily scan, and what every app's images carry.
package imagescanuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UC struct {
	*settings.BaseUC

	appRepo       repository.AppRepo
	imageScanRepo repository.ImageScanRepo
	taskRepo      repository.TaskRepo

	imageScanService imagescanservice.Service
	schedJobService  schedjobservice.Service
	taskQueue        queue.TaskQueue
}

func New(
	baseUC *settings.BaseUC,

	appRepo repository.AppRepo,
	imageScanRepo repository.ImageScanRepo,
	taskRepo repository.TaskRepo,

	imageScanService imagescanservice.Service,
	schedJobService schedjobservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC:           baseUC,
		appRepo:          appRepo,
		imageScanRepo:    imageScanRepo,
		taskRepo:         taskRepo,
		imageScanService: imageScanService,
		schedJobService:  schedJobService,
		taskQueue:        taskQueue,
	}
}
