package permissionimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

func (p *manager) checkProjectAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.ProjectAccessCheck,
) (hasPerm bool, allowedResources map[base.ResourceType][]string, err error) {
	// Project owner has all permissions on the project
	if check.ProjectID != "" {
		project, err := p.projectRepo.GetByIDAndOwner(ctx, db, check.ProjectID, check.SubjectID,
			bunex.SelectColumns("id"),
		)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return false, nil, hperrors.Wrap(err)
		}
		if project != nil {
			return true, nil, nil
		}
	}

	resources := make([]*base.PermissionResource, 0, 3) //nolint:mnd
	if check.ProjectID != "" && check.ProjectEnv != nil {
		resID := ""
		if *check.ProjectEnv != "" {
			resID = projecthelper.CalcProjectEnvID(check.ProjectID, *check.ProjectEnv)
		}
		resources = append(resources, &base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: base.ResourceTypeProjectEnv,
			ResourceID:   resID,
		})
	}
	resources = append(resources,
		&base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: base.ResourceTypeProject,
			ResourceID:   check.ProjectID,
		},
		&base.PermissionResource{
			SubjectType:  check.SubjectType,
			SubjectID:    check.SubjectID,
			ResourceType: base.ResourceTypeModule,
			ResourceID:   string(base.ResourceModuleProject),
		})

	return p.checkAccess(ctx, db, &check.BaseAccessCheck, resources)
}

func (p *manager) LoadProjectAccesses(
	ctx context.Context,
	db database.IDB,
	projectID string,
	projectEnvIDs []string,
	makeAdjustment bool,
) (
	modPerms []*entity.ACLPermission,
	projectPerms []*entity.ACLPermission,
	envPerms map[string][]*entity.ACLPermission,
	err error,
) {
	perms, err := p.LoadProjectRawAccesses(ctx, db, projectID, projectEnvIDs,
		bunex.SelectRelation("SubjectUser",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
		bunex.SelectJoin("JOIN users ON users.id = acl_permission.subj_id"),
		bunex.SelectWhere("users.deleted_at IS NULL"),
		// bunex.SelectWhere("(users.access_expire_at IS NULL OR users.access_expire_at > NOW())"),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}

	mapModPerms := make(map[string]*entity.ACLPermission, 10)           //nolint:mnd
	mapProjectPerms := make(map[string]*entity.ACLPermission, 10)       //nolint:mnd
	mapEnvPerms := make(map[string]map[string]*entity.ACLPermission, 3) //nolint:mnd
	for _, perm := range perms {
		switch {
		case perm.ResourceType == base.ResourceTypeModule:
			mapModPerms[perm.SubjectID] = perm
		case perm.ResourceID == projectID:
			mapProjectPerms[perm.SubjectID] = perm
		default:
			envMap := mapEnvPerms[perm.ResourceID]
			if envMap == nil {
				envMap = make(map[string]*entity.ACLPermission, 10) //nolint:mnd
				mapEnvPerms[perm.ResourceID] = envMap
			}
			envMap[perm.SubjectID] = perm
		}
	}

	if makeAdjustment {
		// Any perm in projectPerms but envPerms -> copy it from projectPerms to envPerms
		for subjectID, perm := range mapProjectPerms {
			for envID, envPermMap := range mapEnvPerms {
				if _, ok := envPermMap[subjectID]; ok {
					continue
				}
				copiedPerm := *perm
				copiedPerm.ResourceType = base.ResourceTypeProjectEnv
				copiedPerm.ResourceID = envID
				envPermMap[subjectID] = &copiedPerm
			}
		}
	}

	modPerms = gofn.MapValues(mapModPerms)
	projectPerms = gofn.MapValues(mapProjectPerms)
	envPerms = make(map[string][]*entity.ACLPermission, len(mapEnvPerms))
	for envID, envPermMap := range mapEnvPerms {
		envPerms[envID] = gofn.MapValues(envPermMap)
	}

	return modPerms, projectPerms, envPerms, nil
}

func (p *manager) LoadProjectAccessUsers(
	ctx context.Context,
	db database.IDB,
	projectID string,
	projectEnvIDs []string,
) (userPerms []*entity.ACLPermission, err error) {
	perms, err := p.LoadProjectRawAccesses(ctx, db, projectID, projectEnvIDs,
		bunex.SelectRelation("SubjectUser",
			bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
		),
		bunex.SelectJoin("JOIN users ON users.id = acl_permission.subj_id"),
		bunex.SelectWhere("users.deleted_at IS NULL"),
		// bunex.SelectWhere("(users.access_expire_at IS NULL OR users.access_expire_at > NOW())"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	mapPerms := make(map[string]*entity.ACLPermission, len(perms))
	for _, perm := range perms {
		if perm.Actions.IsNoAccess() {
			continue
		}
		mapPerms[perm.SubjectID] = perm
	}
	return gofn.MapValues(mapPerms), nil
}

func (p *manager) LoadProjectRawAccesses(
	ctx context.Context,
	db database.IDB,
	projectID string,
	projectEnvIDs []string,
	extraLoadOpts ...bunex.SelectQueryOption,
) ([]*entity.ACLPermission, error) {
	var whereEnvFilter bunex.SelectQueryOption
	if len(projectEnvIDs) > 0 {
		// Filter only satisfied records
		whereEnvFilter = bunex.SelectWhereOr("(acl_permission.res_type = ? AND acl_permission.res_id IN (?))",
			string(base.ResourceTypeProjectEnv), bunex.List(projectEnvIDs))
	} else {
		// Filter all records of the project
		whereEnvFilter = bunex.SelectWhereOr("(acl_permission.res_type = ? AND acl_permission.res_id LIKE ?)",
			string(base.ResourceTypeProjectEnv), projectID+":%")
	}

	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("acl_permission.subj_type = ?", string(base.SubjectTypeUser)),
		bunex.SelectWhereGroup(
			// all ACLs of project
			bunex.SelectWhere("(acl_permission.res_type = ? AND acl_permission.res_id = ?)",
				string(base.ResourceTypeProject), projectID),
			// all ACLs of all project's envs
			whereEnvFilter,
			// all ACLs of project module
			bunex.SelectWhereOr("(acl_permission.res_type = ? AND acl_permission.res_id = ?)",
				string(base.ResourceTypeModule), string(base.ResourceModuleProject)),
		),
	}
	loadOpts = append(loadOpts, extraLoadOpts...)

	perms, _, err := p.aclPermissionRepo.List(ctx, db, nil, loadOpts...)
	if err != nil || len(perms) == 0 {
		return nil, hperrors.Wrap(err)
	}
	return perms, nil
}
