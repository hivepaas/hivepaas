package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// An app can belong to another in two ways, and a listing shows neither of them
// beside the app they belong to:
//
//   - a child app names its parent in app.parent_id, which is what a preview
//     deployment is: a copy of the app it was made from;
//   - a logical child app names it in app.logical_parent_id, which says it was
//     created to serve that app. A template's dependencies are these, and so are
//     the databases a preview clones for itself.
//
// Both are left out unless they are asked for, because a person looking at the
// apps of an environment is looking for the ones they made, not the six a
// template created underneath them.
func excludeChildApps() []bunex.SelectQueryOption {
	return []bunex.SelectQueryOption{
		bunex.SelectWhere("app.parent_id IS NULL"),
		bunex.SelectWhere("app.logical_parent_id IS NULL"),
	}
}

// attachChildApps fills in the children of every app on a page: those that name
// it as their parent, and those created to serve it.
//
// It reads them in two queries for the whole page rather than two per app.
func (uc *UC) attachChildApps(
	ctx context.Context,
	db database.IDB,
	projectID string,
	parents []*entity.App,
	resp []*appdto.AppResp,
) error {
	if len(parents) == 0 {
		return nil
	}
	parentIDs := make([]string, 0, len(parents))
	for _, app := range parents {
		parentIDs = append(parentIDs, app.ID)
	}

	children, err := uc.listAppsBy(ctx, db, projectID, "app.parent_id IN (?)", parentIDs)
	if err != nil {
		return err
	}
	logicalChildren, err := uc.listAppsBy(ctx, db, projectID, "app.logical_parent_id IN (?)", parentIDs)
	if err != nil {
		return err
	}

	byParent, err := transformByOwner(children, func(app *entity.App) string { return app.ParentID })
	if err != nil {
		return err
	}
	byLogicalParent, err := transformByOwner(logicalChildren,
		func(app *entity.App) string { return app.LogicalParentID })
	if err != nil {
		return err
	}

	for _, app := range resp {
		app.ChildApps = byParent[app.ID]
		app.LogicalChildApps = byLogicalParent[app.ID]
	}
	return nil
}

func transformByOwner(apps []*entity.App, ownerOf func(*entity.App) string) (map[string][]*appdto.AppResp, error) {
	out := map[string][]*appdto.AppResp{}
	for _, app := range apps {
		transformed, err := appdto.TransformApp(app, nil)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		owner := ownerOf(app)
		out[owner] = append(out[owner], transformed)
	}
	return out, nil
}

// listAppsBy reads the apps one column points at, with the same relations the
// apps beside them were read with: a nested app is not thinner than a top-level
// one.
func (uc *UC) listAppsBy(
	ctx context.Context,
	db database.IDB,
	projectID string,
	where string,
	ids []string,
) ([]*entity.App, error) {
	apps, _, err := uc.appRepo.List(ctx, db, projectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		),
		bunex.SelectWhereIn(where, ids...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return apps, nil
}
