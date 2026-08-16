package scopeserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func New(
	appService appservice.Service,
	projectService projectservice.Service,
	userService userservice.Service,
) scopeservice.Service {
	return &service{
		appService:     appService,
		projectService: projectService,
		userService:    userService,
	}
}

type service struct {
	appService     appservice.Service
	projectService projectservice.Service
	userService    userservice.Service
}
