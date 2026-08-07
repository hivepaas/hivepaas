package apppreviewserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type service struct {
	appRepo        repository.AppRepo
	deploymentRepo repository.DeploymentRepo
	resLinkRepo    repository.ResLinkRepo
	settingRepo    repository.SettingRepo
	taskRepo       repository.TaskRepo

	appCloneService      appcloneservice.Service
	appDeploymentService appdeploymentservice.Service
	appService           appservice.Service
	domainService        domainservice.Service
	envVarService        envvarservice.Service
}

func New(
	appRepo repository.AppRepo,
	deploymentRepo repository.DeploymentRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	appCloneService appcloneservice.Service,
	appDeploymentService appdeploymentservice.Service,
	appService appservice.Service,
	domainService domainservice.Service,
	envVarService envvarservice.Service,
) apppreviewservice.Service {
	return &service{
		appRepo:        appRepo,
		deploymentRepo: deploymentRepo,
		resLinkRepo:    resLinkRepo,
		settingRepo:    settingRepo,
		taskRepo:       taskRepo,

		appCloneService:      appCloneService,
		appDeploymentService: appDeploymentService,
		appService:           appService,
		domainService:        domainService,
		envVarService:        envVarService,
	}
}
