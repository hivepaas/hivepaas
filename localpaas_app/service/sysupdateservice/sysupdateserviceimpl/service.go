package sysupdateserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sysupdateservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	dbService           dbservice.Service
	lpAppService        lpappservice.Service
	notificationService notificationservice.Service
	traefikService      traefikservice.Service

	dockerManager docker.Manager
}

func New(
	dbService dbservice.Service,
	lpAppService lpappservice.Service,
	notificationService notificationservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) sysupdateservice.Service {
	return &service{
		dbService:           dbService,
		lpAppService:        lpAppService,
		notificationService: notificationService,
		traefikService:      traefikService,

		dockerManager: dockerManager,
	}
}
