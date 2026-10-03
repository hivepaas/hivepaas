package logginguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	*settings.BaseUC

	settingRepo   repository.SettingRepo
	dockerManager docker.Manager

	loggingService loggingservice.Service
	settingService settingservice.Service
	taskQueue      queue.TaskQueue
}

func New(
	baseUC *settings.BaseUC,

	settingRepo repository.SettingRepo,
	dockerManager docker.Manager,

	loggingService loggingservice.Service,
	settingService settingservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC: baseUC,

		settingRepo:   settingRepo,
		dockerManager: dockerManager,

		loggingService: loggingService,
		settingService: settingService,
		taskQueue:      taskQueue,
	}
}
