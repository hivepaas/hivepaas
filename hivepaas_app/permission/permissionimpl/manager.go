package permissionimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type manager struct {
	aclPermissionRepo repository.ACLPermissionRepo
	userRepo          repository.UserRepo
	projectRepo       repository.ProjectRepo
}

func NewManager(
	aclPermissionRepo repository.ACLPermissionRepo,
	userRepo repository.UserRepo,
	projectRepo repository.ProjectRepo,
) permission.Manager {
	return &manager{
		aclPermissionRepo: aclPermissionRepo,
		userRepo:          userRepo,
		projectRepo:       projectRepo,
	}
}
