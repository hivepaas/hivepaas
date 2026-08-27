package volumeuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	currentSettingType    = base.SettingTypeClusterVolume
	currentSettingVersion = entity.CurrentClusterVolumeVersion
)

type UC struct {
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager

	taskRepo repository.TaskRepo

	nodeExecService nodeexecservice.Service
	schedJobService schedjobservice.Service
	taskService     taskservice.Service
	volumeService   volumeservice.Service

	*settings.BaseUC
}

func New(
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,

	taskRepo repository.TaskRepo,

	nodeExecService nodeexecservice.Service,
	schedJobService schedjobservice.Service,
	taskService taskservice.Service,
	volumeService volumeservice.Service,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		taskQueue:     taskQueue,
		dockerManager: dockerManager,

		taskRepo: taskRepo,

		nodeExecService: nodeExecService,
		schedJobService: schedJobService,
		taskService:     taskService,
		volumeService:   volumeService,

		BaseUC: baseUC,
	}
}
