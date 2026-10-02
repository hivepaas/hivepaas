package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// GetAppAutoscale answers a function's autoscale: its settings, its replicas
// now, and why it is paused when it cannot act.
func (uc *UC) GetAppAutoscale(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetAppAutoscaleReq,
) (*appsettingsdto.GetAppAutoscaleResp, error) {
	app, err := uc.loadFunction(ctx, uc.db, req.ProjectID, req.AppID, false)
	if err != nil {
		return nil, err
	}
	setting, err := uc.appAutoscaleSetting(ctx, uc.db, app.ID)
	if err != nil {
		return nil, err
	}
	replicas := 0
	if service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false); err == nil && service != nil {
		replicas = replicasOf(service)
	}
	paused, err := uc.autoscalePaused(ctx, app)
	if err != nil {
		return nil, err
	}
	resp, err := appsettingsdto.TransformAppAutoscale(setting, replicas, paused)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.GetAppAutoscaleResp{Data: resp}, nil
}

// UpdateAppAutoscale saves a function's autoscale, turns the job that scales
// functions on or off with it, and brings the function within its bounds.
func (uc *UC) UpdateAppAutoscale(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppAutoscaleReq,
) (*appsettingsdto.UpdateAppAutoscaleResp, error) {
	next := req.ToEntity()
	var app *entity.App
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var err error
		if app, err = uc.loadFunction(ctx, db, req.ProjectID, req.AppID, true); err != nil {
			return err
		}
		row, err := uc.appAutoscaleSetting(ctx, db, app.ID)
		if err != nil {
			return err
		}
		if row != nil && row.UpdateVer != req.UpdateVer {
			return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
		}
		if next.Enabled {
			paused, err := uc.autoscalePaused(ctx, app)
			if err != nil {
				return err
			}
			if paused != "" {
				return hperrors.Wrap(hperrors.ErrValueInvalid).WithExtraDetail(
					"Autoscale reads the function's calls from its stored logs, which cannot be read (%s): "+
						"turn stored logs on in System → Logging.", paused)
			}
		}

		now := timeutil.NowUTC()
		if row == nil {
			row = &entity.Setting{
				ID: gofn.Must(ulid.NewStringULID()), Scope: base.ObjectScopeApp, ObjectID: app.ID,
				Type: base.SettingTypeAppAutoscale, Status: base.SettingStatusActive, Name: "Autoscale",
				Version: entity.CurrentAppAutoscaleVersion, CreatedAt: now,
			}
		} else {
			row.UpdateVer++
		}
		row.UpdatedAt = now
		if err = row.SetData(next); err != nil {
			return hperrors.Wrap(err)
		}
		persisting := &persistingAppData{}
		persisting.UpsertingSettings = append(persisting.UpsertingSettings, row)
		if err = uc.persistData(ctx, db, persisting); err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.functionAutoscale.EnsureJob(ctx, db); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordAppUpdate(ctx, db, auth, app, base.AuditLogSourceAPIUpdate, "autoscale",
			auditdetail.New().Set("enabled", next.Enabled).Set("minReplicas", next.MinReplicas).
				Set("maxReplicas", next.MaxReplicas).Set("target", next.Target))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &appsettingsdto.UpdateAppAutoscaleResp{Meta: &basedto.Meta{}}
	if next.Enabled {
		if err = uc.scaleIntoBounds(ctx, app, next); err != nil {
			resp.Meta.Warning = "Autoscale saved, but the function could not be brought within its bounds:\n" +
				err.Error()
		}
	}
	return resp, nil
}

// loadFunction is the app, which must be a function: autoscale reads the
// line its runtime writes for every call.
func (uc *UC) loadFunction(
	ctx context.Context, db database.IDB, projectID, appID string, forUpdate bool,
) (*entity.App, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings", bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind)),
	}
	if forUpdate {
		opts = append(opts, bunex.SelectFor("UPDATE OF app"))
	}
	app, err := uc.appService.LoadApp(ctx, db, projectID, appID, forUpdate, forUpdate, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !entity.IsFunctionKind(app.GetSettingByType(base.SettingTypeAppKind)) {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFunction).WithParam("Name", app.Name)
	}
	return app, nil
}

// appAutoscaleSetting is an app's autoscale setting, nil when it has none.
func (uc *UC) appAutoscaleSetting(ctx context.Context, db database.IDB, appID string) (*entity.Setting, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppAutoscale),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}
	return settings[0], nil
}

// autoscalePaused is why a function's calls cannot be read, empty when they
// can: autoscale decides from them.
func (uc *UC) autoscalePaused(ctx context.Context, app *entity.App) (string, error) {
	history, err := uc.loggingService.AppHistory(ctx, uc.db, app)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if history.Available {
		return "", nil
	}
	return string(history.Reason), nil
}

// scaleIntoBounds brings a running function within [Min, Max] at once, rather
// than at the job's next run. A stopped one stays stopped.
func (uc *UC) scaleIntoBounds(ctx context.Context, app *entity.App, autoscale *entity.AppAutoscale) error {
	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil || service == nil {
		return hperrors.Wrap(err)
	}
	current := replicasOf(service)
	if current == 0 {
		return nil
	}
	desired := min(max(current, autoscale.MinReplicas), autoscale.MaxReplicas)
	if desired == current {
		return nil
	}
	replicas := uint64(desired) //nolint:gosec // within 1 and the limit
	err = uc.dockerManager.ServiceUpdateFunc(ctx, service.ID, service, func(_ int, svc *swarm.Service) (bool, error) {
		if svc.Spec.Mode.Replicated == nil || replicasOf(svc) == 0 {
			return false, nil
		}
		svc.Spec.Mode.Replicated.Replicas = &replicas
		return true, nil
	}, 2, 0) //nolint:mnd // retries
	return hperrors.Wrap(err)
}

// replicasOf is a replicated service's replicas, 0 for any other.
func replicasOf(service *swarm.Service) int {
	if service.Spec.Mode.Replicated == nil || service.Spec.Mode.Replicated.Replicas == nil {
		return 0
	}
	return int(*service.Spec.Mode.Replicated.Replicas) //nolint:gosec // a service's replicas
}
