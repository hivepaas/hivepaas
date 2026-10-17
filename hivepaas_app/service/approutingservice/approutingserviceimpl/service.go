package approutingserviceimpl

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	db            *database.DB
	dockerManager docker.Manager

	appRepo repository.AppRepo

	appService     appservice.Service
	networkService networkservice.Service
	settingService settingservice.Service
	sslService     sslservice.Service
	traefikService traefikservice.Service

	// routingApplyRetryDelay is how long ApplyRoutingSettings waits before trying
	// again an update swarm refused as out of sequence.
	routingApplyRetryDelay time.Duration
}

func New(
	db *database.DB,
	dockerManager docker.Manager,

	appRepo repository.AppRepo,

	appService appservice.Service,
	networkService networkservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	traefikService traefikservice.Service,
) approutingservice.Service {
	return &service{
		db:            db,
		dockerManager: dockerManager,

		appRepo: appRepo,

		appService:     appService,
		networkService: networkService,
		settingService: settingService,
		sslService:     sslService,
		traefikService: traefikService,

		routingApplyRetryDelay: routingApplyRetryDelay,
	}
}
