package sslrenewaluc

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
