package appmetricsuc

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logidentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmetricsuc/appmetricsdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// The reasons an app's routes and calls are not shown, besides the logs' and
// the agent's.
const (
	// performanceReasonDisabled: their collection is off, in System > Logging.
	performanceReasonDisabled = "performance-disabled"
	// performanceReasonAppDisabled: it is off for the app, in its Feature
	// Settings.
	performanceReasonAppDisabled = "app-disabled"
	// performanceReasonNodeDisabled: no node the app runs on runs OBI.
	performanceReasonNodeDisabled = "node-disabled"
	// performanceReasonNodeUnsupported: those that should cannot, by what
	// their agents last said: preflightReasons say why.
	performanceReasonNodeUnsupported = "node-unsupported"
)

// performanceView is an app's routes or calls as they can be read: the app,
// the range, what of the app they cover - or why they cannot be.
type performanceView struct {
	app    *entity.App
	head   appmetricsdto.AppPerformanceMetricsHeadResp
	window window
}

// GetAppRouteMetrics reads what an app served over a range ending now, from
// the rows the agent on each node writes from OBI.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized, as for its logs.
func (uc *UC) GetAppRouteMetrics(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmetricsdto.GetAppPerformanceMetricsReq,
) (*appmetricsdto.GetAppRouteMetricsResp, error) {
	view, err := uc.loadPerformanceView(ctx, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := &appmetricsdto.AppRouteMetricsDataResp{AppPerformanceMetricsHeadResp: view.head}
	if !view.head.Available {
		return &appmetricsdto.GetAppRouteMetricsResp{Data: data}, nil
	}
	stats, err := uc.loggingService.RouteMetrics(ctx, uc.db, view.app, &loggingservice.FunctionMetricsQuery{
		Start: view.window.start, End: view.window.end, Step: view.window.step,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data.Totals = performanceCounts(sumHistograms(stats.Buckets))
	data.Series = everyPerformanceStep(stats.Buckets, view.window)
	data.Routes = make([]*appmetricsdto.AppRouteResp, 0, len(stats.Groups))
	for _, g := range stats.Groups {
		data.Routes = append(data.Routes, &appmetricsdto.AppRouteResp{Kind: g.Keys[obi.FieldKind],
			Method: g.Keys[obi.FieldMethod], Route: g.Keys[obi.FieldRoute],
			AppPerformanceCountsResp: *performanceCounts(g.OBIHistogram)})
	}
	err = addReplicas(ctx, uc, view.app, view.window, data.Series,
		func(p *appmetricsdto.AppPerformancePointResp) time.Time { return p.Time },
		func(p *appmetricsdto.AppPerformancePointResp, n *int) { p.Replicas = n })
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmetricsdto.GetAppRouteMetricsResp{Data: data}, nil
}

// GetAppDependencyMetrics reads what an app called over a range ending now -
// other apps, databases, outside hosts - from the rows the agent on each node
// writes from OBI.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized, as for its logs.
func (uc *UC) GetAppDependencyMetrics(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmetricsdto.GetAppPerformanceMetricsReq,
) (*appmetricsdto.GetAppDependencyMetricsResp, error) {
	view, err := uc.loadPerformanceView(ctx, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := &appmetricsdto.AppDependencyMetricsDataResp{AppPerformanceMetricsHeadResp: view.head}
	if !view.head.Available {
		return &appmetricsdto.GetAppDependencyMetricsResp{Data: data}, nil
	}
	stats, err := uc.loggingService.DependencyMetrics(ctx, uc.db, view.app, &loggingservice.FunctionMetricsQuery{
		Start: view.window.start, End: view.window.end, Step: view.window.step,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	peers, err := uc.newPeerResolver(ctx, view.app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data.Kinds = dependencyKinds(stats.Buckets, view.window)
	data.Peers = dependencyPeers(stats.Groups, peers.resolve)
	return &appmetricsdto.GetAppDependencyMetricsResp{Data: data}, nil
}

// loadPerformanceView loads the app and says whether its routes and calls can
// be read, over which range. The switches come first - the system's, then the
// app's - so that while they are off, nothing more is asked of Docker or the
// logs; then the agent, the logs and the app's nodes.
func (uc *UC) loadPerformanceView(
	ctx context.Context,
	req *appmetricsdto.GetAppPerformanceMetricsReq,
) (*performanceView, error) {
	app, features, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	view := &performanceView{app: app, head: appmetricsdto.AppPerformanceMetricsHeadResp{Range: req.Range}}

	perf, err := uc.loggingPerformance(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	switch {
	case !perf.Enabled:
		view.head.Reason = performanceReasonDisabled
		return view, nil
	case features.PerformanceSettings == nil || !features.PerformanceSettings.Enabled:
		view.head.Reason = performanceReasonAppDisabled
		return view, nil
	}

	agentSvc, err := uc.hpAppService.GetHpAgentSwarmService(ctx)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if agentSvc == nil || !logidentity.HasComponent(&agentSvc.Spec, base.LogComponentAgent) {
		view.head.Reason = resourceMetricsReasonAgentUnlabelled
		return view, nil
	}
	history, err := uc.loggingService.ProxyHistory(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !history.Available {
		view.head.Reason = string(history.Reason)
		return view, nil
	}
	if err = uc.performanceCoverage(ctx, app, perf, &view.head); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if view.head.Reason != "" {
		return view, nil
	}

	view.window = metricsWindow(req.Range, timeutil.NowUTC(), history.Retention)
	view.head.Available = true
	view.head.Start, view.head.End = view.window.start, view.window.end
	view.head.StepSeconds, view.head.Clamped = int(view.window.step.Seconds()), view.window.clamped
	return view, nil
}

// loggingPerformance is which nodes run OBI: none when the logging settings
// were never saved.
func (uc *UC) loggingPerformance(ctx context.Context) (*entity.LoggingPerformance, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true)
	if errors.Is(err, hperrors.ErrNotFound) || (err == nil && setting == nil) {
		return &entity.LoggingPerformance{}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := setting.AsLoggingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg.Performance == nil {
		return &entity.LoggingPerformance{}, nil
	}
	return cfg.Performance, nil
}

// performanceCoverage counts the nodes the app runs on now, and those of them
// that run OBI - by the settings, and by what their agents last said - and
// sets the reason when none does. An app running nowhere has its past shown.
func (uc *UC) performanceCoverage(
	ctx context.Context,
	app *entity.App,
	perf *entity.LoggingPerformance,
	head *appmetricsdto.AppPerformanceMetricsHeadResp,
) error {
	tasks, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return hperrors.Wrap(err)
	}
	nodes := map[string]bool{}
	for i := range tasks.Items {
		if t := &tasks.Items[i]; t.NodeID != "" && t.Status.State == swarm.TaskStateRunning {
			nodes[t.NodeID] = true
		}
	}
	head.Nodes = len(nodes)
	on := make([]string, 0, len(nodes))
	for id := range nodes {
		if perf.Node(id) != nil {
			on = append(on, id)
		}
	}
	if len(nodes) == 0 {
		return nil
	}
	if len(on) == 0 {
		head.Reason = performanceReasonNodeDisabled
		return nil
	}
	statuses, err := uc.loggingService.PerformanceStatus(ctx, uc.db, 3*obi.StatusEvery) //nolint:mnd // three rows
	if err != nil {
		return hperrors.Wrap(err)
	}
	var reasons []string
	for _, id := range on {
		// A node whose agent has said nothing yet is counted: it may.
		if status := statuses[id]; status != nil && !status.Preflight.OK {
			reasons = append(reasons, status.Preflight.Reasons...)
			continue
		}
		head.NodesCovered++
	}
	if head.NodesCovered == 0 {
		head.Reason = performanceReasonNodeUnsupported
		slices.Sort(reasons)
		head.PreflightReasons = slices.Compact(reasons)
	}
	return nil
}

// sumHistograms sums buckets' rows.
func sumHistograms(buckets []*logging.OBIBucket) logging.OBIHistogram {
	sum := logging.OBIHistogram{Buckets: map[string]int64{}}
	for _, b := range buckets {
		addHistogram(&sum, b.OBIHistogram)
	}
	return sum
}

func addHistogram(to *logging.OBIHistogram, h logging.OBIHistogram) {
	to.Count += h.Count
	to.Errors += h.Errors
	to.SumMs += h.SumMs
	if to.Buckets == nil {
		to.Buckets = map[string]int64{}
	}
	obi.AddBuckets(to.Buckets, h.Buckets)
}

// performanceCounts are summed rows as the API answers them: the quantiles
// read from their buckets.
func performanceCounts(h logging.OBIHistogram) *appmetricsdto.AppPerformanceCountsResp {
	return &appmetricsdto.AppPerformanceCountsResp{Requests: h.Count, Errors: h.Errors,
		P50: obi.Quantile(h.Buckets, 0.5),  //nolint:mnd // the median
		P95: obi.Quantile(h.Buckets, 0.95), //nolint:mnd // a percentile
		P99: obi.Quantile(h.Buckets, 0.99), //nolint:mnd // a percentile
	}
}

// everyPerformanceStep is a point per step of the window, oldest first, from
// one series' buckets: a step without one is a point with none.
func everyPerformanceStep(buckets []*logging.OBIBucket, w window) []*appmetricsdto.AppPerformancePointResp {
	byTime := make(map[int64]*logging.OBIBucket, len(buckets))
	for _, b := range buckets {
		byTime[b.Time.Unix()] = b
	}
	out := make([]*appmetricsdto.AppPerformancePointResp, 0, int(w.end.Sub(w.start)/w.step))
	for at := w.start; at.Before(w.end); at = at.Add(w.step) {
		point := &appmetricsdto.AppPerformancePointResp{Time: at}
		if b, ok := byTime[at.Unix()]; ok {
			point.AppPerformanceCountsResp = *performanceCounts(b.OBIHistogram)
		}
		out = append(out, point)
	}
	return out
}

// dependencyKinds are the calls by kind, the busiest first: each its totals
// and a point per step.
func dependencyKinds(buckets []*logging.OBIBucket, w window) []*appmetricsdto.AppDependencyKindResp {
	byKind := map[string][]*logging.OBIBucket{}
	for _, b := range buckets {
		byKind[b.Keys[obi.FieldKind]] = append(byKind[b.Keys[obi.FieldKind]], b)
	}
	out := make([]*appmetricsdto.AppDependencyKindResp, 0, len(byKind))
	for kind, of := range byKind {
		out = append(out, &appmetricsdto.AppDependencyKindResp{Kind: kind, Totals: performanceCounts(sumHistograms(of)),
			Series: everyPerformanceStep(of, w)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Totals.Requests != out[j].Totals.Requests {
			return out[i].Totals.Requests > out[j].Totals.Requests
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

// dependencyPeers merges the calls by peer and operation into the peers, the
// busiest first, each its operations, the busiest first as they come.
func dependencyPeers(
	groups []*logging.OBIGroup,
	resolve func(kind, peer string) *entity.App,
) []*appmetricsdto.AppDependencyPeerResp {
	type peerKey struct{ kind, peer string }
	sums := map[peerKey]*logging.OBIHistogram{}
	byKey := map[peerKey]*appmetricsdto.AppDependencyPeerResp{}
	out := make([]*appmetricsdto.AppDependencyPeerResp, 0, len(groups))
	for _, g := range groups {
		key := peerKey{kind: g.Keys[obi.FieldKind], peer: g.Keys[obi.FieldPeer]}
		peer := byKey[key]
		if peer == nil {
			peer = &appmetricsdto.AppDependencyPeerResp{Kind: key.kind, Peer: key.peer}
			if app := resolve(key.kind, key.peer); app != nil {
				peer.App = &appmetricsdto.AppDependencyAppResp{ID: app.ID, Key: app.Key, Name: app.Name}
			}
			byKey[key], sums[key] = peer, &logging.OBIHistogram{Buckets: map[string]int64{}}
			out = append(out, peer)
		}
		addHistogram(sums[key], g.OBIHistogram)
		peer.Operations = append(peer.Operations, &appmetricsdto.AppDependencyOperationResp{
			Method: g.Keys[obi.FieldMethod], Operation: g.Keys[obi.FieldOperation],
			AppPerformanceCountsResp: *performanceCounts(g.OBIHistogram)})
	}
	for key, peer := range byKey {
		peer.AppPerformanceCountsResp = *performanceCounts(*sums[key])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Requests > out[j].Requests })
	return out
}
