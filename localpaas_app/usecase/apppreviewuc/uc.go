package apppreviewuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appPreviewService apppreviewservice.Service
	appService        appservice.Service
	settingService    settingservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appPreviewService apppreviewservice.Service,
	appService appservice.Service,
	settingService settingservice.Service,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appPreviewService: appPreviewService,
		appService:        appService,
		settingService:    settingService,
	}
}
