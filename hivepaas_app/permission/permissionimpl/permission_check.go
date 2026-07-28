package permissionimpl

import (
	"context"
	"maps"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (p *manager) CheckAccess(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	check permission.AccessCheck,
) (hasPerm bool, err error) {
	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return true, nil
	}

	check.InitSubject(auth)

	var allowedResources map[base.ResourceType][]string
	switch chk := check.(type) {
	case *permission.ModuleAccessCheck:
		hasPerm, err = p.checkModuleAccess(ctx, db, chk)
	case *permission.GeneralResourceAccessCheck:
		hasPerm, allowedResources, err = p.checkResourceAccess(ctx, db, chk)
	case *permission.AppAccessCheck:
		hasPerm, allowedResources, err = p.checkAppAccess(ctx, db, chk)
	case *permission.ProjectAccessCheck:
		hasPerm, allowedResources, err = p.checkProjectAccess(ctx, db, chk)
	}
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	if len(allowedResources) > 0 {
		hasPerm = true
		if auth.AllowedResources == nil {
			auth.AllowedResources = allowedResources
		} else {
			auth.AllowedResources = make(map[base.ResourceType][]string, len(allowedResources))
			maps.Copy(auth.AllowedResources, allowedResources)
		}
	}

	return hasPerm, nil
}
