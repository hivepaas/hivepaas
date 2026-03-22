package sslrenewaluc

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/cronjobservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type SSLRenewalUC struct {
	*settings.BaseSettingUC
	appRepo        repository.AppRepo
	taskRepo       repository.TaskRepo
	taskService    taskservice.TaskService
	cronJobService cronjobservice.CronJobService
	taskQueue      queue.TaskQueue
}

func NewSSLRenewalUC(
	baseSettingUC *settings.BaseSettingUC,
	appRepo repository.AppRepo,
	taskRepo repository.TaskRepo,
	taskService taskservice.TaskService,
	cronJobService cronjobservice.CronJobService,
	taskQueue queue.TaskQueue,
) *SSLRenewalUC {
	return &SSLRenewalUC{
		BaseSettingUC:  baseSettingUC,
		appRepo:        appRepo,
		taskRepo:       taskRepo,
		taskService:    taskService,
		cronJobService: cronJobService,
		taskQueue:      taskQueue,
	}
}
