package schedjobuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeSchedJob
	currentSettingVersion = entity.CurrentSchedJobVersion
)

type UC struct {
	taskQueue queue.TaskQueue

	taskRepo repository.TaskRepo

	schedJobService schedjobservice.Service
	taskService     taskservice.Service

	*settings.BaseUC
}

func New(
	taskQueue queue.TaskQueue,

	taskRepo repository.TaskRepo,

	schedJobService schedjobservice.Service,
	taskService taskservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskQueue: taskQueue,

		taskRepo: taskRepo,

		schedJobService: schedJobService,
		taskService:     taskService,

		BaseUC: baseUC,
	}
}
