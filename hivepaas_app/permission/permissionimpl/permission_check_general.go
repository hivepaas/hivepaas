package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (p *manager) checkResourceAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.GeneralResourceAccessCheck,
) (hasPerm bool, allowedResources map[base.ResourceType][]string, err error) {
	resources := make([]*base.PermissionResource, 0, 2) //nolint:mnd
	if check.ResourceType != "" {
		resources = append(resources, &base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: check.ResourceType,
			ResourceID:   check.ResourceID,
		})
	}
	if check.Module != "" {
		resources = append(resources,
			&base.PermissionResource{
				SubjectType:  check.SubjectType,
				SubjectID:    check.SubjectID,
				ResourceType: base.ResourceTypeModule,
				ResourceID:   string(check.Module),
			})
	}

	return p.checkAccess(ctx, db, &check.BaseAccessCheck, resources)
}

func (p *manager) checkAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.BaseAccessCheck,
	resources []*base.PermissionResource,
) (hasPerm bool, allowedResources map[base.ResourceType][]string, err error) {
	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, resources)
	if err != nil || len(perms) == 0 {
		return false, nil, hperrors.Wrap(err)
	}

	// Check permission on a specific resource (ResourceID must be not empty)
	for _, res := range resources {
		if res.ResourceID == "" {
			continue
		}
		for _, perm := range perms {
			if res.ResourceType != perm.ResourceType {
				continue
			}
			if res.ResourceID == perm.ResourceID {
				return p.hasPermission(perm, check), nil, nil
			}
		}
	}

	// When user has no permission, collect IDs of all resources of resource types the user has permissions on.
	// This is usually to allow users seeing individual accessible objects of a resource type.
	allowedResources = make(map[base.ResourceType][]string)
	for _, res := range resources {
		if res.ResourceID != "" {
			continue
		}
		for _, perm := range perms {
			if res.ResourceType != perm.ResourceType {
				continue
			}
			allowedResources[res.ResourceType] = append(allowedResources[res.ResourceType], perm.ResourceID)
		}
	}

	return hasPerm, allowedResources, nil
}

func (p *manager) hasPermission(
	perm *entity.ACLPermission,
	check *permission.BaseAccessCheck,
) bool {
	switch {
	case check.Action != "":
		if perm.Actions.Allows(check.Action) {
			return true
		}
	case len(check.AllOf) > 0:
		if perm.Actions.AllowsAll(check.AllOf) {
			return true
		}
	case len(check.AnyOf) > 0:
		if perm.Actions.AllowsAny(check.AnyOf) {
			return true
		}
	}
	return false
}
