package permissionimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (p *manager) CheckAccess(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	check *permission.AccessCheck,
) (hasPerm bool, err error) {
	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return true, nil
	}

	if check.SubjectID == "" {
		check.SubjectType = base.SubjectTypeUser
		check.SubjectID = auth.User.ID
	}

	// Special project/app access check
	hasPerm, err = p.checkProjectAccess(ctx, db, check)
	if err != nil {
		return false, apperrors.New(err)
	}
	if hasPerm {
		return true, nil
	}

	modPerms, parentPerms, objPerms, err := p.loadPermissions(ctx, db, check)
	if err != nil {
		return false, apperrors.New(err)
	}
	defer func() {
		// When user has no permission on the given resource, collect IDs of all other resources user has permission on.
		// This is usually to allow users listing accessible objects when they don't have permission on the module.
		if !hasPerm && check.ResourceID == "" {
			for _, perm := range objPerms {
				if p.hasPermission(perm, check) {
					auth.AllowObjectIDs = append(auth.AllowObjectIDs, perm.ResourceID)
				}
			}
			hasPerm = len(auth.AllowObjectIDs) > 0
		}
	}()

	if check.ResourceType != "" && check.ResourceID != "" {
		for _, perm := range objPerms {
			hasPerm = p.hasPermission(perm, check)
			return hasPerm, nil
		}
	}

	if check.ParentResourceType != "" && check.ParentResourceID != "" {
		for _, perm := range parentPerms {
			hasPerm = p.hasPermission(perm, check)
			return hasPerm, nil
		}
	}

	if check.ResourceModule != "" {
		for _, perm := range modPerms {
			hasPerm = p.hasPermission(perm, check)
			return hasPerm, nil
		}
	}

	return hasPerm, nil
}

// checkProjectAccess owners of a project have all permissions on the project and the belonging apps
func (p *manager) checkProjectAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.AccessCheck,
) (hasPerm bool, err error) {
	if check.SubjectType != base.SubjectTypeUser || check.SubjectID == "" {
		return false, nil
	}

	var projectID string
	switch check.ResourceType { //nolint:exhaustive
	case base.ResourceTypeProject:
		projectID = check.ResourceID
	case base.ResourceTypeApp:
		projectID = check.ParentResourceID
	}
	if projectID == "" {
		return false, nil
	}

	project, err := p.projectRepo.GetByIDAndOwner(ctx, db, projectID, check.SubjectID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return false, apperrors.New(err)
	}

	if project == nil {
		return false, nil
	}
	return true, nil
}

func (p *manager) loadPermissions(
	ctx context.Context,
	db database.IDB,
	check *permission.AccessCheck,
	opts ...bunex.SelectQueryOption,
) (modPerms, parentPerms, objPerms []*entity.ACLPermission, err error) {
	var resources []*base.PermissionResource
	if check.ResourceModule != "" {
		resources = append(resources, &base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: base.ResourceTypeModule,
			ResourceID:   string(check.ResourceModule),
		})
	}
	if check.ResourceType != "" {
		resources = append(resources, &base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: check.ResourceType,
			ResourceID:   check.ResourceID,
		})
	}
	if check.ParentResourceType != "" && check.ParentResourceID != "" {
		resources = append(resources, &base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: check.ParentResourceType,
			ResourceID:   check.ParentResourceID,
		})
	}

	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, resources, opts...)
	if err != nil || len(perms) == 0 {
		return nil, nil, nil, apperrors.New(err)
	}

	for _, perm := range perms {
		if perm.ResourceType == base.ResourceTypeModule && perm.ResourceID == string(check.ResourceModule) {
			modPerms = append(modPerms, perm)
			continue
		}
		if perm.ResourceType == check.ParentResourceType && perm.ResourceID == check.ParentResourceID {
			parentPerms = append(parentPerms, perm)
			continue
		}
		objPerms = append(objPerms, perm)
	}
	return modPerms, parentPerms, objPerms, nil
}

func (p *manager) hasPermission(
	perm *entity.ACLPermission,
	check *permission.AccessCheck,
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

func (p *manager) LoadObjectAccesses(
	ctx context.Context,
	db database.IDB,
	check *permission.AccessCheck,
	extraLoadOpts ...bunex.SelectQueryOption,
) (objPerms, modPerms []*entity.ACLPermission, err error) {
	if check.ResourceID == "" {
		return nil, nil, nil
	}
	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectRelation("SubjectUser",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
		bunex.SelectJoin("JOIN users ON users.id = acl_permission.subj_id"),
		bunex.SelectWhere("users.deleted_at IS NULL"),
		bunex.SelectWhere("users.status = ?", base.UserStatusActive),
		bunex.SelectWhere("(users.access_expire_at IS NULL OR users.access_expire_at > NOW())"),
	}
	loadOpts = append(loadOpts, extraLoadOpts...)

	modPerms, parentPerms, objPerms, err := p.loadPermissions(ctx, db, check, loadOpts...)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	objPermMap := make(map[string]struct{})
	for _, perm := range objPerms {
		objPermMap[perm.SubjectID] = struct{}{}
	}
	for _, perm := range parentPerms {
		if _, exists := objPermMap[perm.SubjectID]; !exists {
			objPerms = append(objPerms, perm)
		}
	}

	return objPerms, modPerms, nil
}

func (p *manager) MergeObjectAccessesBySubjectID(
	objPerms, modPerms []*entity.ACLPermission,
) []*entity.ACLPermission {
	res := make([]*entity.ACLPermission, 0, len(objPerms)+len(modPerms))
	objPermMap := make(map[string]struct{}, len(objPerms))
	for _, perm := range objPerms {
		res = append(res, perm)
		objPermMap[perm.SubjectID] = struct{}{}
	}
	for _, perm := range modPerms {
		if _, exists := objPermMap[perm.SubjectID]; !exists {
			res = append(res, perm)
		}
	}
	return res
}
