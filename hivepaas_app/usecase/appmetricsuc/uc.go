package appmetricsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// UC reads an app's metrics: a function's calls, an app's requests through
// the proxy, its containers' resources, and what it served and called.
type UC struct {
	db            *database.DB
	dockerManager docker.Manager

	appRepo     repository.AppRepo
	settingRepo repository.SettingRepo

	appService          appservice.Service
	appAutoscaleService appautoscaleservice.Service
	hpAppService        hpappservice.Service
	loggingService      loggingservice.Service
	traefikService      traefikservice.Service
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,

	appService appservice.Service,
	appAutoscaleService appautoscaleservice.Service,
	hpAppService hpappservice.Service,
	loggingService loggingservice.Service,
	traefikService traefikservice.Service,
) *UC {
	return &UC{
		db:            db,
		dockerManager: dockerManager,

		appRepo:     appRepo,
		settingRepo: settingRepo,

		appService:          appService,
		appAutoscaleService: appAutoscaleService,
		hpAppService:        hpAppService,
		loggingService:      loggingService,
		traefikService:      traefikService,
	}
}
