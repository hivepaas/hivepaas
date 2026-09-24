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
	if check.ProjectID == "" && (check.ProjectEnv != nil && *check.ProjectEnv != "") {
		check.ProjectID, _ = projecthelper.ParseProjectEnvID(*check.ProjectEnv)
	}
	if check.ProjectID == "" && check.ProjectEnv == nil {
		return p.checkProjectsAccess(ctx, db, check)
	}

	// Project owner has all permissions on the project
	if check.ProjectID != "" {
		owned, err := p.ownsProject(ctx, db, check.SubjectID, check.ProjectID)
		if err != nil || owned {
			return owned, nil, err
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
		p.projectModuleResource(check))

	hasPerm, allowedResources, err = p.checkAccess(ctx, db, &check.BaseAccessCheck, resources)
	if err != nil || hasPerm || check.ProjectEnv != nil || check.Action != base.ActionTypeRead {
		return hasPerm, allowedResources, err
	}

	// Reading the project as a whole. Grants are written per env, so one on any
	// env of the project is a way into the project, or a user given a single env
	// could never open the project it is in. Only reading: changing or deleting
	// the project reaches every env, and a grant on one says nothing of the rest.
	envIDs, err := p.grantedEnvs(ctx, db, &check.BaseAccessCheck)
	if err != nil {
		return false, nil, err
	}
	for _, envID := range envIDs {
		if projectID, _ := projecthelper.ParseProjectEnvID(envID); projectID == check.ProjectID {
			return true, nil, nil
		}
	}
	return false, nil, nil
}

// checkProjectsAccess answers a question about projects in general: which ones
// may be read, for a list, or whether a project may be made at all.
//
// The project module grant answers for every project. Short of it, reading is
// allowed on the projects the user was given - on the project, on any of its
// envs - or owns, and those are what the list is narrowed to. Anything other
// than reading, such as making a project, is the module's alone: a grant inside
// one project says nothing about the others, or about new ones.
func (p *manager) checkProjectsAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.ProjectAccessCheck,
) (bool, map[base.ResourceType][]string, error) {
	module := p.projectModuleResource(check)
	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, []*base.PermissionResource{module})
	if err != nil {
		return false, nil, hperrors.Wrap(err)
	}
	for _, perm := range perms {
		if perm.ResourceType == module.ResourceType && perm.ResourceID == module.ResourceID &&
			p.hasPermission(perm, &check.BaseAccessCheck) {
			return true, nil, nil
		}
	}
	if check.Action != base.ActionTypeRead {
		return false, nil, nil
	}

	projectIDs, err := p.grantedProjects(ctx, db, &check.BaseAccessCheck)
	if err != nil {
		return false, nil, err
	}
	owned, _, err := p.projectRepo.List(ctx, db, nil,
		bunex.SelectColumns("id"),
		bunex.SelectWhere("project.owner_id = ?", check.SubjectID),
	)
	if err != nil {
		return false, nil, hperrors.Wrap(err)
	}
	for _, project := range owned {
		projectIDs = append(projectIDs, project.ID)
	}
	if len(projectIDs) == 0 {
		return false, nil, nil
	}
	return true, map[base.ResourceType][]string{base.ResourceTypeProject: gofn.ToSet(projectIDs)}, nil
}

// grantedProjects is the projects the check's action is granted on, on the
// project itself or on any of its envs.
func (p *manager) grantedProjects(
	ctx context.Context,
	db database.IDB,
	check *permission.BaseAccessCheck,
) ([]string, error) {
	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, []*base.PermissionResource{
		{SubjectType: check.SubjectType, SubjectID: check.SubjectID, ResourceType: base.ResourceTypeProject},
		{SubjectType: check.SubjectType, SubjectID: check.SubjectID, ResourceType: base.ResourceTypeProjectEnv},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	projectIDs := make([]string, 0, len(perms))
	for _, perm := range perms {
		if !p.hasPermission(perm, check) {
			continue
		}
		if perm.ResourceType == base.ResourceTypeProject {
			projectIDs = append(projectIDs, perm.ResourceID)
			continue
		}
		if projectID, _ := projecthelper.ParseProjectEnvID(perm.ResourceID); projectID != "" {
			projectIDs = append(projectIDs, projectID)
		}
	}
	return projectIDs, nil
}

// grantedEnvs is the envs, of any project, the check's action is granted on.
func (p *manager) grantedEnvs(
	ctx context.Context,
	db database.IDB,
	check *permission.BaseAccessCheck,
) ([]string, error) {
	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, []*base.PermissionResource{
		{SubjectType: check.SubjectType, SubjectID: check.SubjectID, ResourceType: base.ResourceTypeProjectEnv},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	envIDs := make([]string, 0, len(perms))
	for _, perm := range perms {
		if perm.ResourceType == base.ResourceTypeProjectEnv && p.hasPermission(perm, check) {
			envIDs = append(envIDs, perm.ResourceID)
		}
	}
	return envIDs, nil
}

func (p *manager) ownsProject(ctx context.Context, db database.IDB, userID, projectID string) (bool, error) {
	project, err := p.projectRepo.GetByIDAndOwner(ctx, db, projectID, userID, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return false, hperrors.Wrap(err)
	}
	return project != nil, nil
}

func (p *manager) projectModuleResource(check *permission.ProjectAccessCheck) *base.PermissionResource {
	return &base.PermissionResource{
		SubjectType:  check.SubjectType,
		SubjectID:    check.SubjectID,
		ResourceType: base.ResourceTypeModule,
		ResourceID:   string(base.ResourceModuleProject),
	}
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
