package schedjobtriggerserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	schedJobService schedjobservice.Service
	taskQueue       queue.TaskQueue
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	schedJobService schedjobservice.Service,
	taskQueue queue.TaskQueue,
) schedjobtriggerservice.Service {
	return &service{
		db:              db,
		settingRepo:     settingRepo,
		taskRepo:        taskRepo,
		schedJobService: schedJobService,
		taskQueue:       taskQueue,
	}
}
