package schedjobuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeSchedJob
	currentSettingVersion = entity.CurrentSchedJobVersion
)

type UC struct {
	*settings.BaseUC
	appRepo         repository.AppRepo
	taskRepo        repository.TaskRepo
	taskService     taskservice.Service
	schedJobService schedjobservice.Service
	taskQueue       queue.TaskQueue
}

func New(
	baseUC *settings.BaseUC,
	appRepo repository.AppRepo,
	taskRepo repository.TaskRepo,
	taskService taskservice.Service,
	schedJobService schedjobservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC:          baseUC,
		appRepo:         appRepo,
		taskRepo:        taskRepo,
		taskService:     taskService,
		schedJobService: schedJobService,
		taskQueue:       taskQueue,
	}
}
