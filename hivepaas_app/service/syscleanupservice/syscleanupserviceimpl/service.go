package syscleanupserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	// db is the database apart from the cleanup's own transaction: the system
	// apps' sync commits there, feature by feature.
	db database.IDB

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
	loggingService        loggingservice.Service
	registryService       registryservice.Service

	obiSettings   cacherepository.OBISettingsRepo
	taskQueue     queue.TaskQueue
	dockerManager docker.Manager
}

func New(
	db *database.DB,

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
	loggingService loggingservice.Service,
	registryService registryservice.Service,

	obiSettings cacherepository.OBISettingsRepo,
	taskQueue queue.TaskQueue,
	dockerManager docker.Manager,
) syscleanupservice.Service {
	return &service{
		db: db,

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
		loggingService:        loggingService,
		registryService:       registryService,

		obiSettings:   obiSettings,
		taskQueue:     taskQueue,
		dockerManager: dockerManager,
	}
}
