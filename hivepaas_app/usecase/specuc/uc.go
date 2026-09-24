package specuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db *database.DB

	permissionManager permission.Manager
	auditService      auditservice.Service
	specService       specservice.Service
	taskQueue         queue.TaskQueue
}

func New(
	db *database.DB,

	permissionManager permission.Manager,
	auditService auditservice.Service,
	specService specservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		db: db,

		permissionManager: permissionManager,
		auditService:      auditService,
		specService:       specService,
		taskQueue:         taskQueue,
	}
}
