package appserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadChildApps(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	loadChildApps, loadLogicalChildApps bool,
) error {
	if app.ParentID != "" {
		return nil
	}
	childApps, _, err := s.appRepo.List(ctx, db, app.ProjectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectWhereGroup(
			bunex.SelectWhere("(app.parent_id = ? AND TRUE = ?)", app.ID, loadChildApps),
			bunex.SelectWhereOr("(app.logical_parent_id = ? AND TRUE = ?)", app.ID, loadLogicalChildApps),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, childApp := range childApps {
		if childApp.ParentID == app.ID {
			app.ChildApps = append(app.ChildApps, childApp)
			continue
		}
		if childApp.LogicalParentID == app.ID {
			app.LogicalChildApps = append(app.LogicalChildApps, childApp)
			continue
		}
	}
	return nil
}
