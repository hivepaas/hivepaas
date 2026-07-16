package hpappsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db        *database.DB
	taskQueue queue.TaskQueue

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appService     appservice.Service
	domainService  domainservice.Service
	hpAppService   hpappservice.Service
	networkService networkservice.Service
	settingService settingservice.Service
	sslService     sslservice.Service
	taskService    taskservice.Service
	traefikService traefikservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	domainService domainservice.Service,
	hpAppService hpappservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	taskService taskservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:        db,
		taskQueue: taskQueue,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appService:     appService,
		domainService:  domainService,
		hpAppService:   hpAppService,
		networkService: networkService,
		settingService: settingService,
		sslService:     sslService,
		taskService:    taskService,
		traefikService: traefikService,

		dockerManager: dockerManager,
	}
}
