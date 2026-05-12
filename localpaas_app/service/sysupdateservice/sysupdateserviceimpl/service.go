package sysupdateserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/service/dbservice"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sysupdateservice"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/services/docker"
)

type service struct {
	logger              logging.Logger
	settingService      settingservice.Service
	lpAppService        lpappservice.Service
	traefikService      traefikservice.Service
	dbService           dbservice.Service
	userService         userservice.Service
	taskService         taskservice.Service
	notificationService notificationservice.Service
	dockerManager       docker.Manager
}

func New(
	logger logging.Logger,
	settingService settingservice.Service,
	lpAppService lpappservice.Service,
	traefikService traefikservice.Service,
	dbService dbservice.Service,
	userService userservice.Service,
	taskService taskservice.Service,
	notificationService notificationservice.Service,
	dockerManager docker.Manager,
) sysupdateservice.Service {
	return &service{
		logger:              logger,
		settingService:      settingService,
		lpAppService:        lpAppService,
		traefikService:      traefikService,
		dbService:           dbService,
		userService:         userService,
		taskService:         taskService,
		notificationService: notificationService,
		dockerManager:       dockerManager,
	}
}
