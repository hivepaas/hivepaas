package specuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db *database.DB

	projectRepo repository.ProjectRepo

	permissionManager permission.Manager
	auditService      auditservice.Service
	composeService    composeservice.Service
	projectService    projectservice.Service
	specService       specservice.Service
	taskQueue         queue.TaskQueue
}

func New(
	db *database.DB,

	projectRepo repository.ProjectRepo,

	permissionManager permission.Manager,
	auditService auditservice.Service,
	composeService composeservice.Service,
	projectService projectservice.Service,
	specService specservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		db: db,

		projectRepo: projectRepo,

		permissionManager: permissionManager,
		auditService:      auditService,
		composeService:    composeService,
		projectService:    projectService,
		specService:       specService,
		taskQueue:         taskQueue,
	}
}
