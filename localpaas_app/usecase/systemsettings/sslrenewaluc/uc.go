package sslrenewaluc

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UC struct {
	*settings.BaseUC
	appRepo        repository.AppRepo
	taskRepo       repository.TaskRepo
	taskService    taskservice.Service
	cronJobService cronjobservice.Service
	taskQueue      queue.TaskQueue
}

func New(
	baseSettingUC *settings.BaseUC,
	appRepo repository.AppRepo,
	taskRepo repository.TaskRepo,
	taskService taskservice.Service,
	cronJobService cronjobservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		BaseUC:         baseSettingUC,
		appRepo:        appRepo,
		taskRepo:       taskRepo,
		taskService:    taskService,
		cronJobService: cronJobService,
		taskQueue:      taskQueue,
	}
}
