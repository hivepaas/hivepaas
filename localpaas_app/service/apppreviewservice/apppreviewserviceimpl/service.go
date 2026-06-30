package apppreviewserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/appcopyservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/service/domainservice"
)

type service struct {
	appRepo        repository.AppRepo
	deploymentRepo repository.DeploymentRepo
	taskRepo       repository.TaskRepo

	appCopyService       appcopyservice.Service
	appDeploymentService appdeploymentservice.Service
	appService           appservice.Service
	domainService        domainservice.Service
}

func New(
	appRepo repository.AppRepo,
	deploymentRepo repository.DeploymentRepo,
	taskRepo repository.TaskRepo,

	appCopyService appcopyservice.Service,
	appDeploymentService appdeploymentservice.Service,
	appService appservice.Service,
	domainService domainservice.Service,
) apppreviewservice.Service {
	return &service{
		appRepo:        appRepo,
		deploymentRepo: deploymentRepo,
		taskRepo:       taskRepo,

		appCopyService:       appCopyService,
		appDeploymentService: appDeploymentService,
		appService:           appService,
		domainService:        domainService,
	}
}
