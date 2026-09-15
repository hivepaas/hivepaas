package permissionimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

type manager struct {
	aclPermissionRepo repository.ACLPermissionRepo
	appRepo           repository.AppRepo
	userRepo          repository.UserRepo
	projectRepo       repository.ProjectRepo

	auditService auditservice.Service
}

func NewManager(
	aclPermissionRepo repository.ACLPermissionRepo,
	appRepo repository.AppRepo,
	userRepo repository.UserRepo,
	projectRepo repository.ProjectRepo,

	auditService auditservice.Service,
) permission.Manager {
	return &manager{
		aclPermissionRepo: aclPermissionRepo,
		appRepo:           appRepo,
		userRepo:          userRepo,
		projectRepo:       projectRepo,

		auditService: auditService,
	}
}
