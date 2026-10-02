package approutingserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

// ReapplyRouteNames implements approutingservice.Service.
//
// It is the move from names by key to names by id: an app whose service
// carries no stale name is not touched, so the sweep is cheap once done and
// safe to run at every start. References are strict, as on any apply forward:
// an app whose basic auth is gone keeps its old names rather than lose it.
func (s *service) ReapplyRouteNames(
	ctx context.Context,
	db database.IDB,
) (*approutingservice.ReapplyRouteNamesResp, error) {
	apps, _, err := s.appRepo.List(ctx, db, "", nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &approutingservice.ReapplyRouteNamesResp{Failed: map[string]string{}}
	for _, app := range sweepTargets(apps, "", false) {
		applied, err := s.reapplyAppRouteNames(ctx, db, app)
		switch {
		case err != nil:
			resp.Failed[app.ID] = err.Error()
		case applied:
			resp.Applied++
		}
	}
	return resp, nil
}

func (s *service) reapplyAppRouteNames(ctx context.Context, db database.IDB, app *entity.App) (bool, error) {
	setting := app.GetSettingByType(base.SettingTypeAppRouting)
	if setting == nil || app.ServiceID == "" {
		return false, nil
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if errors.Is(err, hperrors.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if !traefikservice.HasStaleRouteNames(inspect.Service.Spec.Labels, app.ID) {
		return false, nil
	}
	routingSettings, err := setting.AsAppRoutingSettings()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	req := sweepApplyReq(app, routingSettings, false)
	req.Service = &inspect.Service
	if _, err = s.ApplyRoutingSettings(ctx, db, req); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}
