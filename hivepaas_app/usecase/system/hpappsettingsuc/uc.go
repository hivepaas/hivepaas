package hpappsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/datakeyservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db            *database.DB
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appRoutingService approutingservice.Service
	appService        appservice.Service
	dataKeyService    datakeyservice.Service
	domainService     domainservice.Service
	hpAppService      hpappservice.Service
	networkService    networkservice.Service
	settingService    settingservice.Service
	sslService        sslservice.Service
	systemEventBus    systemeventbusservice.Service
	taskService       taskservice.Service
	traefikService    traefikservice.Service
}

func New(
	db *database.DB,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appRoutingService approutingservice.Service,
	appService appservice.Service,
	dataKeyService datakeyservice.Service,
	domainService domainservice.Service,
	hpAppService hpappservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	systemEventBus systemeventbusservice.Service,
	taskService taskservice.Service,
	traefikService traefikservice.Service,

) *UC {
	return &UC{
		db:            db,
		taskQueue:     taskQueue,
		dockerManager: dockerManager,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appRoutingService: appRoutingService,
		appService:        appService,
		dataKeyService:    dataKeyService,
		domainService:     domainService,
		hpAppService:      hpAppService,
		networkService:    networkService,
		settingService:    settingService,
		sslService:        sslService,
		systemEventBus:    systemEventBus,
		taskService:       taskService,
		traefikService:    traefikService,
	}
}
