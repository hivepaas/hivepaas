package permissionimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type manager struct {
	aclPermissionRepo repository.ACLPermissionRepo
	appRepo           repository.AppRepo
	userRepo          repository.UserRepo
	projectRepo       repository.ProjectRepo
}

func NewManager(
	aclPermissionRepo repository.ACLPermissionRepo,
	appRepo repository.AppRepo,
	userRepo repository.UserRepo,
	projectRepo repository.ProjectRepo,
) permission.Manager {
	return &manager{
		aclPermissionRepo: aclPermissionRepo,
		appRepo:           appRepo,
		userRepo:          userRepo,
		projectRepo:       projectRepo,
	}
}
