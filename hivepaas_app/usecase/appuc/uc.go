package appuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db            *database.DB
	dockerManager docker.Manager
	taskQueue     queue.TaskQueue

	appRepo       repository.AppRepo
	binObjectRepo repository.BinObjectRepo
	settingRepo   repository.SettingRepo

	auditService           auditservice.Service
	appCloneService        appcloneservice.Service
	appDeploymentService   appdeploymentservice.Service
	appProvisionService    appprovisionservice.Service
	appService             appservice.Service
	schedJobTriggerService schedjobtriggerservice.Service
	clusterService         clusterservice.Service
	containerExecService   containerexecservice.Service
	functionService        functionservice.Service
	settingService         settingservice.Service
	loggingService         loggingservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	binObjectRepo repository.BinObjectRepo,
	settingRepo repository.SettingRepo,

	auditService auditservice.Service,
	appCloneService appcloneservice.Service,
	appDeploymentService appdeploymentservice.Service,
	appProvisionService appprovisionservice.Service,
	appService appservice.Service,
	schedJobTriggerService schedjobtriggerservice.Service,
	clusterService clusterservice.Service,
	containerExecService containerexecservice.Service,
	functionService functionservice.Service,
	settingService settingservice.Service,
	loggingService loggingservice.Service,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,
		taskQueue:     taskQueue,

		appRepo:       appRepo,
		binObjectRepo: binObjectRepo,
		settingRepo:   settingRepo,

		auditService:           auditService,
		appCloneService:        appCloneService,
		appDeploymentService:   appDeploymentService,
		appProvisionService:    appProvisionService,
		appService:             appService,
		schedJobTriggerService: schedJobTriggerService,
		clusterService:         clusterService,
		containerExecService:   containerExecService,
		functionService:        functionService,
		settingService:         settingService,
		loggingService:         loggingService,
	}
}
