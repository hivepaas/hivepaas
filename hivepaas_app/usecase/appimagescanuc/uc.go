// Package appimagescanuc is an app's image scan: what its images carry, and a
// scan asked for. See service/imagescanservice.
package appimagescanuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appRepo       repository.AppRepo
	imageScanRepo repository.ImageScanRepo
	settingRepo   repository.SettingRepo
	taskRepo      repository.TaskRepo

	appService       appservice.Service
	imageScanService imagescanservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	imageScanRepo repository.ImageScanRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	appService appservice.Service,
	imageScanService imagescanservice.Service,
) *UC {
	return &UC{
		db:               db,
		taskQueue:        taskQueue,
		appRepo:          appRepo,
		imageScanRepo:    imageScanRepo,
		settingRepo:      settingRepo,
		taskRepo:         taskRepo,
		appService:       appService,
		imageScanService: imageScanService,
	}
}
