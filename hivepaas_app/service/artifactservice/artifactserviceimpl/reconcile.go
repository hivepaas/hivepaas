package artifactserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
)

func (s *service) ReconcileAll(ctx context.Context, db database.IDB) (*artifactservice.ReconcileAllResp, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppArtifacts),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return &artifactservice.ReconcileAllResp{}, nil
	}
	appIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		appIDs = append(appIDs, setting.ObjectID)
	}
	apps, err := s.appRepo.ListByIDs(ctx, db, "", appIDs,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appByID := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appByID[app.ID] = app
	}

	resp := &artifactservice.ReconcileAllResp{}
	for _, setting := range settings {
		app := appByID[setting.ObjectID]
		if app == nil || app.ServiceID == "" {
			continue
		}
		reconciled, err := s.reconcileApp(ctx, app, setting)
		if err != nil {
			resp.Failed = append(resp.Failed, &artifactservice.ReconcileFailure{App: app, Err: err})
			continue
		}
		if reconciled {
			resp.Apps++
		}
	}
	return resp, nil
}

// reconcileApp has an app's artifacts where its service may start a task now;
// false for an app whose service is gone, as one being deleted.
func (s *service) reconcileApp(ctx context.Context, app *entity.App, setting *entity.Setting) (bool, error) {
	artifacts, err := setting.AsAppArtifactsSettings()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return false, nil
		}
		return false, hperrors.Wrap(err)
	}
	if _, err = s.MountsFor(ctx, artifacts, &inspect.Service.Spec); err != nil {
		return false, err
	}
	return true, nil
}
