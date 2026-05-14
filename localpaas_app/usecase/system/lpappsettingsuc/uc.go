package lpappsettingsuc

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
)

type UC struct {
	db            *database.DB
	settingRepo   repository.SettingRepo
	lpAppService  lpappservice.Service
	taskService   taskservice.Service
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager
}

func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	lpAppService lpappservice.Service,
	taskService taskservice.Service,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:            db,
		settingRepo:   settingRepo,
		lpAppService:  lpAppService,
		taskService:   taskService,
		taskQueue:     taskQueue,
		dockerManager: dockerManager,
	}
}
