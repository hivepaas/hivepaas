package appuc

import (
	"context"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// metricsRanges are the ranges metrics are given for, each with its step.
var metricsRanges = map[string]struct{ length, step time.Duration }{
	"1h":  {time.Hour, time.Minute},
	"6h":  {6 * time.Hour, 5 * time.Minute},
	"24h": {24 * time.Hour, 15 * time.Minute},
	"7d":  {7 * 24 * time.Hour, time.Hour},
}

// GetFunctionMetrics counts a function's calls over a range ending now, from
// the invocation lines its runtime writes into its logs.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized, as for the logs it reads.
func (uc *UC) GetFunctionMetrics(
	ctx context.Context,
	_ *basedto.Auth,
	req *appdto.GetFunctionMetricsReq,
) (*appdto.GetFunctionMetricsResp, error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings", bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !entity.IsFunctionKind(app.GetSettingByType(base.SettingTypeAppKind)) {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFunction).WithParam("Name", app.Name)
	}
	if featureSettings.LoggingSettings != nil && !featureSettings.LoggingSettings.Enabled {
		return nil, hperrors.NewUnavailable("App logs")
	}

	history, err := uc.loggingService.AppHistory(ctx, uc.db, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !history.Available {
		return &appdto.GetFunctionMetricsResp{Data: &appdto.FunctionMetricsDataResp{
			Range: req.Range, Reason: string(history.Reason),
		}}, nil
	}

	window := metricsWindow(req.Range, timeutil.NowUTC(), history.Retention)
	stats, err := uc.loggingService.FunctionMetrics(ctx, uc.db, app, &loggingservice.FunctionMetricsQuery{
		Start: window.start, End: window.end, Step: window.step,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := toFunctionMetricsData(req.Range, window, stats)
	if err = uc.addReplicas(ctx, app, window, data); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appdto.GetFunctionMetricsResp{Data: data}, nil
}

// addReplicas puts the function's replicas on its points, when it has
// autoscale on or was scaled within the range: from its scalings, and its
// count now for the time after the last.
func (uc *UC) addReplicas(ctx context.Context, app *entity.App, w window, data *appdto.FunctionMetricsDataResp) error {
	events, err := uc.functionAutoscale.Events(ctx, uc.db, app.ID, w.start, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	enabled, err := uc.autoscaleEnabled(ctx, app.ID)
	if err != nil {
		return err
	}
	if !enabled && len(events) == 0 {
		return nil
	}
	inspect, err := uc.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return nil //nolint:nilerr // no service, no count now: the calls are answered without it
	}
	current := 0
	if mode := inspect.Service.Spec.Mode.Replicated; mode != nil && mode.Replicas != nil {
		current = int(*mode.Replicas) //nolint:gosec // a service's replicas
	}
	// The latest first, as Events answers them: oldest first for the walk.
	slices.Reverse(events)
	for i, point := range data.Series {
		replicas := replicasAt(point.Time.Add(w.step), events, current, i == len(data.Series)-1)
		data.Series[i].Replicas = &replicas
	}
	return nil
}

// replicasAt is the replicas at a step's end, from the scalings oldest first:
// after the last before it, or before the first after it; now's count for the
// last step past the last scaling, and with no scaling at all.
func replicasAt(at time.Time, events []*functionautoscaleservice.Event, current int, last bool) int {
	var before *functionautoscaleservice.Event
	for _, e := range events {
		if e.Time.After(at) {
			if before == nil {
				return e.From
			}
			break
		}
		before = e
	}
	if before == nil || (last && before == events[len(events)-1]) {
		return current
	}
	return before.To
}

// autoscaleEnabled is whether the app has autoscale on.
func (uc *UC) autoscaleEnabled(ctx context.Context, appID string) (bool, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppAutoscale),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectLimit(1),
	)
	if err != nil || len(settings) == 0 {
		return false, hperrors.Wrap(err)
	}
	autoscale, err := settings[0].AsAppAutoscale()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return autoscale.Enabled, nil
}

// window is the range of a metrics request: whole steps, ending with the step
// now is in.
type window struct {
	start, end time.Time
	step       time.Duration
	// clamped: the logs are kept for less than the range, which starts where
	// they do.
	clamped bool
}

func metricsWindow(rangeName string, now time.Time, retention timeutil.Duration) window {
	r := metricsRanges[rangeName]
	w := window{end: now.Truncate(r.step).Add(r.step), step: r.step}
	w.start = w.end.Add(-r.length)
	if kept := time.Duration(retention); kept > 0 && kept < r.length {
		w.start = now.Add(-kept).Truncate(r.step).Add(r.step)
		w.clamped = true
	}
	return w
}

func toFunctionMetricsData(rangeName string, w window, stats *logging.InvocationStatsResp,
) *appdto.FunctionMetricsDataResp {
	data := &appdto.FunctionMetricsDataResp{
		Available:   true,
		Range:       rangeName,
		Start:       w.start,
		End:         w.end,
		StepSeconds: int(w.step / time.Second),
		Clamped:     w.clamped,
		Totals:      countsResp(stats.Totals),
		ByOutcome:   stats.ByOutcome,
		Series:      make([]*appdto.FunctionMetricsPointResp, 0, len(stats.Buckets)),
		ByPath:      make([]*appdto.FunctionMetricsPathResp, 0, len(stats.ByPath)),
	}
	for _, p := range stats.ByPath {
		data.ByPath = append(data.ByPath, &appdto.FunctionMetricsPathResp{
			Method: p.Method, Path: p.Path, FunctionMetricsCountsResp: *countsResp(p.InvocationCounts),
		})
	}
	for _, b := range stats.Buckets {
		data.Series = append(data.Series, &appdto.FunctionMetricsPointResp{
			Time: b.Time, FunctionMetricsCountsResp: *countsResp(b.InvocationCounts),
		})
	}
	return data
}

func countsResp(c logging.InvocationCounts) *appdto.FunctionMetricsCountsResp {
	return &appdto.FunctionMetricsCountsResp{
		Calls: c.Calls, Failed: c.Failed, Errors4xx: c.Errors4xx, Errors5xx: c.Errors5xx,
		P50: c.P50, P95: c.P95, P99: c.P99,
	}
}
