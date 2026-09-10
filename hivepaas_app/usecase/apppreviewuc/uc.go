package apppreviewuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	taskRepo repository.TaskRepo

	appPreviewService apppreviewservice.Service
	auditService      auditservice.Service
	appService        appservice.Service
	settingService    settingservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	taskRepo repository.TaskRepo,

	appPreviewService apppreviewservice.Service,
	auditService auditservice.Service,
	appService appservice.Service,
	settingService settingservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		taskRepo: taskRepo,

		appPreviewService: appPreviewService,
		auditService:      auditService,
		appService:        appService,
		settingService:    settingService,
	}
}
