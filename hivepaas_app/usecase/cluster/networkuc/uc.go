package networkuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType    = base.SettingTypeClusterNetwork
	currentSettingVersion = entity.CurrentClusterNetworkVersion
)

type UC struct {
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager

	taskRepo repository.TaskRepo

	schedJobService schedjobservice.Service
	taskService     taskservice.Service

	*settings.BaseUC
}

func New(
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,

	taskRepo repository.TaskRepo,

	schedJobService schedjobservice.Service,
	taskService taskservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskQueue:     taskQueue,
		dockerManager: dockerManager,

		taskRepo: taskRepo,

		schedJobService: schedJobService,
		taskService:     taskService,

		BaseUC: baseUC,
	}
}
