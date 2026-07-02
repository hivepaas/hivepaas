package hpappsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	settingRepo repository.SettingRepo

	hpAppService hpappservice.Service
	taskService  taskservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	settingRepo repository.SettingRepo,

	hpAppService hpappservice.Service,
	taskService taskservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		settingRepo: settingRepo,

		hpAppService: hpAppService,
		taskService:  taskService,

		dockerManager: dockerManager,
	}
}
