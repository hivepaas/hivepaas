package permissionimpl

import (
	"context"
	"errors"

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
	//
	// The app's own env decides, never the one an address names: the app is looked up by its
	// project and ID alone, so a production app reached at an address of development would
	// otherwise be acted on with development's grant. An app not made yet - one an import is to
	// make - is in the env it is to be made in.
	if check.AppID != "" {
		app, err := p.appRepo.GetByID(ctx, db, check.ProjectID, check.AppID,
			bunex.SelectColumns("id", "project_id", "project_env_id", "parent_id"),
			bunex.SelectRelation("ProjectEnv", bunex.SelectColumns("id", "key")),
		)
		switch {
		case err == nil:
			check.ProjectID = app.ProjectID
			check.ParentID = app.ParentID
			check.ProjectEnv = app.ProjectEnv.Key
		case errors.Is(err, hperrors.ErrNotFound) && check.ProjectID != "" && check.ProjectEnv != "":
		default:
			return false, nil, hperrors.Wrap(err)
		}
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
