package syscleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	appRepo        repository.AppRepo
	auditLogRepo   repository.AuditLogRepo
	deploymentRepo repository.DeploymentRepo
	fileRepo       repository.FileRepo
	lockRepo       repository.LockRepo
	resLinkRepo    repository.ResLinkRepo
	sysErrorRepo   repository.SysErrorRepo
	taskLogRepo    repository.TaskLogRepo
	taskRepo       repository.TaskRepo

	clusterCleanupService clustercleanupservice.Service
	agentService          agentservice.Service
	appService            appservice.Service
	auditService          auditservice.Service
	fileService           fileservice.Service

	dockerManager docker.Manager
}

func New(
	appRepo repository.AppRepo,
	auditLogRepo repository.AuditLogRepo,
	deploymentRepo repository.DeploymentRepo,
	fileRepo repository.FileRepo,
	lockRepo repository.LockRepo,
	resLinkRepo repository.ResLinkRepo,
	sysErrorRepo repository.SysErrorRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,

	clusterCleanupService clustercleanupservice.Service,
	agentService agentservice.Service,
	appService appservice.Service,
	auditService auditservice.Service,
	fileService fileservice.Service,

	dockerManager docker.Manager,
) syscleanupservice.Service {
	return &service{
		appRepo:        appRepo,
		auditLogRepo:   auditLogRepo,
		deploymentRepo: deploymentRepo,
		fileRepo:       fileRepo,
		lockRepo:       lockRepo,
		resLinkRepo:    resLinkRepo,
		sysErrorRepo:   sysErrorRepo,
		taskLogRepo:    taskLogRepo,
		taskRepo:       taskRepo,

		clusterCleanupService: clusterCleanupService,
		agentService:          agentService,
		appService:            appService,
		auditService:          auditService,
		fileService:           fileService,

		dockerManager: dockerManager,
	}
}
