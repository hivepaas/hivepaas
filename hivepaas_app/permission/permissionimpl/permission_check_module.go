package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (p *manager) checkModuleAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.ModuleAccessCheck,
) (bool, error) {
	resources := []*base.PermissionResource{ // order matters, the check will go from current scope to higher
		{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: base.ResourceTypeModule,
			ResourceID:   string(check.Module),
		},
	}

	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, resources)
	if err != nil || len(perms) == 0 {
		return false, apperrors.Wrap(err)
	}

	for _, res := range resources {
		for _, perm := range perms {
			if res.ResourceType != perm.ResourceType {
				continue
			}
			if res.ResourceID != "" && res.ResourceID == perm.ResourceID {
				return p.hasPermission(perm, &check.BaseAccessCheck), nil
			}
		}
	}

	return false, nil
}
