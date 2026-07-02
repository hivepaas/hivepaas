package sysupdateserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dbservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	dbService           dbservice.Service
	hpAppService        hpappservice.Service
	notificationService notificationservice.Service
	traefikService      traefikservice.Service

	dockerManager docker.Manager
}

func New(
	dbService dbservice.Service,
	hpAppService hpappservice.Service,
	notificationService notificationservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) sysupdateservice.Service {
	return &service{
		dbService:           dbService,
		hpAppService:        hpAppService,
		notificationService: notificationService,
		traefikService:      traefikService,

		dockerManager: dockerManager,
	}
}
