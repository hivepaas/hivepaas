package appmetricsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmetricsuc/appmetricsdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// httpMetricsReasonNotExposed: the app is reached by no domain, so no request
// of its goes through the proxy.
const httpMetricsReasonNotExposed = "not-exposed"

// GetAppHTTPMetrics counts an app's requests over a range ending now, from the
// proxy's access log.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized, as for its logs.
func (uc *UC) GetAppHTTPMetrics(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmetricsdto.GetAppHTTPMetricsReq,
) (*appmetricsdto.GetAppHTTPMetricsResp, error) {
	app, _, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings", bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	unavailable := func(reason string) *appmetricsdto.GetAppHTTPMetricsResp {
		return &appmetricsdto.GetAppHTTPMetricsResp{
			Data: &appmetricsdto.AppHTTPMetricsDataResp{Range: req.Range, Reason: reason},
		}
	}

	if !isExposed(app) {
		return unavailable(httpMetricsReasonNotExposed), nil
	}
	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if reason := traefikservice.AccessLogReadiness(&traefikSvc.Spec); reason != "" {
		return unavailable(string(reason)), nil
	}
	history, err := uc.loggingService.ProxyHistory(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !history.Available {
		return unavailable(string(history.Reason)), nil
	}

	window := metricsWindow(req.Range, timeutil.NowUTC(), history.Retention)
	stats, err := uc.loggingService.HTTPMetrics(ctx, uc.db, app, &loggingservice.FunctionMetricsQuery{
		Start: window.start, End: window.end, Step: window.step,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := toAppHTTPMetricsData(req.Range, window, stats)
	err = addReplicas(ctx, uc, app, window, data.Series,
		func(p *appmetricsdto.AppHTTPPointResp) time.Time { return p.Time },
		func(p *appmetricsdto.AppHTTPPointResp, n *int) { p.Replicas = n })
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmetricsdto.GetAppHTTPMetricsResp{Data: data}, nil
}

// isExposed is whether the app is reached by a domain through the proxy.
func isExposed(app *entity.App) bool {
	setting := app.GetSettingByType(base.SettingTypeAppRouting)
	if setting == nil {
		return false
	}
	routing, err := setting.AsAppRoutingSettings()
	if err != nil || !routing.ExposePublicly {
		return false
	}
	for _, domain := range routing.Domains {
		if domain.Enabled {
			return true
		}
	}
	return false
}

func toAppHTTPMetricsData(
	rangeName string, w window, stats *logging.HTTPStatsResp,
) *appmetricsdto.AppHTTPMetricsDataResp {
	data := &appmetricsdto.AppHTTPMetricsDataResp{
		Available:   true,
		Range:       rangeName,
		Start:       w.start,
		End:         w.end,
		StepSeconds: int(w.step.Seconds()),
		Clamped:     w.clamped,
		Totals:      httpCountsResp(stats.Totals),
		Series:      make([]*appmetricsdto.AppHTTPPointResp, 0, len(stats.Buckets)),
		ByPath:      make([]*appmetricsdto.AppHTTPPathResp, 0, len(stats.ByPath)),
		ByReplica:   make([]*appmetricsdto.AppHTTPReplicaResp, 0, len(stats.ByReplica)),
	}
	for _, b := range stats.Buckets {
		data.Series = append(data.Series, &appmetricsdto.AppHTTPPointResp{
			Time: b.Time, AppHTTPCountsResp: *httpCountsResp(b.HTTPCounts)})
	}
	for _, p := range stats.ByPath {
		data.ByPath = append(data.ByPath, &appmetricsdto.AppHTTPPathResp{
			Method: p.Method, Path: p.Path, AppHTTPCountsResp: *httpCountsResp(p.HTTPCounts)})
	}
	for _, r := range stats.ByReplica {
		data.ByReplica = append(data.ByReplica, &appmetricsdto.AppHTTPReplicaResp{
			Address: r.Address, AppHTTPCountsResp: *httpCountsResp(r.HTTPCounts)})
	}
	return data
}

func httpCountsResp(c logging.HTTPCounts) *appmetricsdto.AppHTTPCountsResp {
	return &appmetricsdto.AppHTTPCountsResp{
		Requests: c.Requests, Errors4xx: c.Errors4xx, Errors5xx: c.Errors5xx, Unreachable: c.Unreachable,
		P50: c.P50, P95: c.P95, P99: c.P99,
	}
}
