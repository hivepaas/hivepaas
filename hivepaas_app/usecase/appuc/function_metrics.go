package appuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	return &appdto.GetFunctionMetricsResp{Data: toFunctionMetricsData(req.Range, window, stats)}, nil
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
