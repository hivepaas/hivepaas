package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (p *manager) checkAppAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.AppAccessCheck,
) (hasPerm bool, allowedResources map[base.ResourceType][]string, err error) {
	// NOTE: for simplicity, we only allow putting permissions on from project env up to higher scope.
	// So if a user has a permission on a project env, they will have that permission on all the apps
	// within the env.
	if check.AppID != "" && (check.ProjectID == "" || check.ProjectEnv == "") {
		// loads app to get missing fields
		app, err := p.appRepo.GetByID(ctx, db, check.ProjectID, check.AppID,
			bunex.SelectColumns("id", "project_id", "project_env_id", "parent_id"),
			bunex.SelectRelation("ProjectEnv", bunex.SelectColumns("id", "key")),
		)
		if err != nil {
			return false, nil, hperrors.Wrap(err)
		}
		check.ProjectID = app.ProjectID
		check.ParentID = app.ParentID
		check.ProjectEnv = app.ProjectEnv.Key
	}

	hasPerm, allowedResources, err = p.checkProjectAccess(ctx, db, &permission.ProjectAccessCheck{
		BaseAccessCheck: check.BaseAccessCheck,
		ProjectID:       check.ProjectID,
		ProjectEnv:      &check.ProjectEnv,
	})
	if err != nil {
		return false, nil, hperrors.Wrap(err)
	}
	return hasPerm, allowedResources, nil
}
