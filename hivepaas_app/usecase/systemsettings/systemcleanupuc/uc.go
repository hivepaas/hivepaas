package systemcleanupuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UC struct {
	taskQueue queue.TaskQueue

	taskRepo repository.TaskRepo

	schedJobService schedjobservice.Service

	*settings.BaseUC
}

func New(
	taskQueue queue.TaskQueue,

	taskRepo repository.TaskRepo,

	schedJobService schedjobservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskQueue: taskQueue,

		taskRepo: taskRepo,

		schedJobService: schedJobService,

		BaseUC: baseUC,
	}
}
