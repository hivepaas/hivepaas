package appsettingsuc

import (
	"context"
	"strings"
	"time"

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

const (
	// autoscaleEventsSince and autoscaleEventsLimit bound the scalings the
	// settings show: the latest of the last week.
	autoscaleEventsSince = 7 * 24 * time.Hour
	autoscaleEventsLimit = 20
)

// GetAppAutoscale answers an app's autoscale: its settings, its replicas now,
// why it cannot act, what it can scale on, and its latest scalings.
func (uc *UC) GetAppAutoscale(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetAppAutoscaleReq,
) (*appsettingsdto.GetAppAutoscaleResp, error) {
	app, isFunction, err := uc.loadAutoscaledApp(ctx, uc.db, req.ProjectID, req.AppID, false)
	if err != nil {
		return nil, err
	}
	setting, err := uc.appAutoscaleSetting(ctx, uc.db, app.ID)
	if err != nil {
		return nil, err
	}
	autoscale := entity.NewAppAutoscale()
	if setting != nil {
		if autoscale, err = setting.AsAppAutoscale(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		service = nil
	}
	st := &appsettingsdto.AppAutoscaleState{IsFunction: isFunction}
	if service != nil {
		st.Replicas = replicasOf(service)
	}
	if st.Paused, st.Check, err = uc.autoscaleBlocker(ctx, uc.db, app, isFunction, autoscale, service); err != nil {
		return nil, err
	}
	st.Events, err = uc.appAutoscale.Events(ctx, uc.db, app.ID, timeutil.NowUTC().Add(-autoscaleEventsSince),
		autoscaleEventsLimit)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := appsettingsdto.TransformAppAutoscale(setting, st)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.GetAppAutoscaleResp{Data: resp}, nil
}

// UpdateAppAutoscale saves an app's autoscale, turns the job that scales apps
// on or off with it, and brings the app within its bounds.
func (uc *UC) UpdateAppAutoscale(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppAutoscaleReq,
) (*appsettingsdto.UpdateAppAutoscaleResp, error) {
	next := req.ToEntity()
	var app *entity.App
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var err error
		var isFunction bool
		if app, isFunction, err = uc.loadAutoscaledApp(ctx, db, req.ProjectID, req.AppID, true); err != nil {
			return err
		}
		row, err := uc.appAutoscaleSetting(ctx, db, app.ID)
		if err != nil {
			return err
		}
		if row != nil && row.UpdateVer != req.UpdateVer {
			return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
		}
		if err = uc.checkAutoscaleUpdate(ctx, db, app, isFunction, row, next); err != nil {
			return err
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
		if err = uc.appAutoscale.EnsureJob(ctx, db); err != nil {
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

// loadAutoscaledApp is the app, with its kind and routing settings, and
// whether it is a function: a function scales on its calls, any other app on
// its requests and its CPU.
func (uc *UC) loadAutoscaledApp(
	ctx context.Context, db database.IDB, projectID, appID string, forUpdate bool,
) (*entity.App, bool, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings", bunex.SelectWhereIn("setting.type IN (?)",
			base.SettingTypeAppKind, base.SettingTypeAppRouting)),
	}
	if forUpdate {
		opts = append(opts, bunex.SelectFor("UPDATE OF app"))
	}
	app, err := uc.appService.LoadApp(ctx, db, projectID, appID, forUpdate, forUpdate, opts...)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	return app, entity.IsFunctionKind(app.GetSettingByType(base.SettingTypeAppKind)), nil
}

// autoscaleBlocker is why an app's autoscale cannot act now, "" when it can,
// and, for an app other than a function, what it can scale on.
func (uc *UC) autoscaleBlocker(
	ctx context.Context, db database.IDB, app *entity.App, isFunction bool, autoscale *entity.AppAutoscale,
	service *swarm.Service,
) (string, *appautoscaleservice.Check, error) {
	if isFunction {
		paused, err := uc.autoscalePaused(ctx, app)
		if err != nil {
			return "", nil, err
		}
		if paused == "" && service != nil && service.Spec.Mode.Replicated == nil {
			paused = appautoscaleservice.RefusedNotReplicated
		}
		return paused, nil, nil
	}
	check, err := uc.appAutoscale.Check(ctx, db, app, service)
	if err != nil {
		return "", nil, hperrors.Wrap(err)
	}
	if check.Refused != "" {
		return check.Refused, check, nil
	}
	if unreadable := unreadableSignals(autoscale, check); len(unreadable) > 0 {
		return unreadable[0].Reason, check, nil
	}
	return "", check, nil
}

// signalReason is a signal an app scales on, and why it cannot be read.
type signalReason struct {
	Signal string
	Reason string
}

// unreadableSignals is the signals an app scales on when none of them can be
// read, each with why; nil when one can, or it scales on none.
func unreadableSignals(autoscale *entity.AppAutoscale, check *appautoscaleservice.Check) []signalReason {
	var out []signalReason
	if autoscale.RequestsTarget > 0 {
		if check.Requests == "" {
			return nil
		}
		out = append(out, signalReason{Signal: "requests", Reason: check.Requests})
	}
	if autoscale.CPUTarget > 0 {
		if check.CPU == "" {
			return nil
		}
		out = append(out, signalReason{Signal: "CPU", Reason: check.CPU})
	}
	return out
}

// unreadableText says why each signal an app scales on cannot be read.
func unreadableText(unreadable []signalReason) string {
	parts := make([]string, 0, len(unreadable))
	for _, u := range unreadable {
		parts = append(parts, "its "+u.Signal+" cannot be read, as "+appautoscaleservice.ReasonText(u.Reason))
	}
	return strings.Join(parts, "; and ")
}

// checkAutoscaleUpdate refuses an update that leaves autoscale on and unable
// to act, as checkAutoscale says; turned off, anything goes.
func (uc *UC) checkAutoscaleUpdate(
	ctx context.Context, db database.IDB, app *entity.App, isFunction bool, row *entity.Setting,
	next *entity.AppAutoscale,
) error {
	if !next.Enabled {
		return nil
	}
	wasOn := false
	if row != nil {
		current, err := row.AsAppAutoscale()
		if err != nil {
			return hperrors.Wrap(err)
		}
		wasOn = current.Enabled
	}
	return uc.checkAutoscale(ctx, db, app, isFunction, next, !wasOn)
}

// checkAutoscale refuses an autoscale that is on and could never act: an app
// scaling on nothing, or publishing a port on its node. Being turned on, it
// refuses one that cannot act now - a function's calls, or every signal an
// app scales on, unreadable; one already on is left to say it is paused, its
// settings still changed.
func (uc *UC) checkAutoscale(
	ctx context.Context, db database.IDB, app *entity.App, isFunction bool, next *entity.AppAutoscale,
	turningOn bool,
) error {
	if isFunction {
		if !turningOn {
			return nil
		}
		paused, err := uc.autoscalePaused(ctx, app)
		if err != nil {
			return err
		}
		if paused != "" {
			return hperrors.Wrap(hperrors.ErrAutoscaleUnreadable).WithParam("Name", app.Name).WithParam("Reason",
				"its calls are read from its stored logs, and "+appautoscaleservice.ReasonText(paused))
		}
		return nil
	}
	if next.RequestsTarget == 0 && next.CPUTarget == 0 {
		return hperrors.Wrap(hperrors.ErrAutoscaleNoSignal).WithParam("Name", app.Name)
	}
	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		service = nil
	}
	check, err := uc.appAutoscale.Check(ctx, db, app, service)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if check.Refused == appautoscaleservice.RefusedHostPorts {
		return hperrors.Wrap(hperrors.ErrAutoscaleHostPorts).WithParam("Name", app.Name)
	}
	if unreadable := unreadableSignals(next, check); turningOn && len(unreadable) > 0 {
		return hperrors.Wrap(hperrors.ErrAutoscaleUnreadable).WithParam("Name", app.Name).
			WithParam("Reason", unreadableText(unreadable))
	}
	return nil
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
