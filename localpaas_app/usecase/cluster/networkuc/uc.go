package networkuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/services/docker"
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
