package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logidentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// resourceMetricsReasonAgentUnlabelled: the agent's lines do not carry its
// identity yet - an install from before, until the agent's next update - so
// its rows cannot be told from a line an app printed.
const resourceMetricsReasonAgentUnlabelled = "agent-unlabelled"

// GetAppResourceMetrics reads an app's containers' CPU and memory over a range
// ending now, from the rows the agent on each node writes.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized, as for its logs.
func (uc *UC) GetAppResourceMetrics(
	ctx context.Context,
	_ *basedto.Auth,
	req *appdto.GetAppResourceMetricsReq,
) (*appdto.GetAppResourceMetricsResp, error) {
	app, _, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
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
	unavailable := func(reason string) *appdto.GetAppResourceMetricsResp {
		return &appdto.GetAppResourceMetricsResp{
			Data: &appdto.AppResourceMetricsDataResp{Range: req.Range, Reason: reason}}
	}

	agentSvc, err := uc.hpAppService.GetHpAgentSwarmService(ctx)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if agentSvc == nil || !logidentity.HasComponent(&agentSvc.Spec, base.LogComponentAgent) {
		return unavailable(resourceMetricsReasonAgentUnlabelled), nil
	}
	history, err := uc.loggingService.ProxyHistory(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !history.Available {
		return unavailable(string(history.Reason)), nil
	}

	window := metricsWindow(req.Range, timeutil.NowUTC(), history.Retention)
	stats, err := uc.loggingService.ResourceMetrics(ctx, uc.db, app, &loggingservice.FunctionMetricsQuery{
		Start: window.start, End: window.end, Step: window.step,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := toAppResourceMetricsData(req.Range, window, stats)
	err = addReplicas(ctx, uc, app, window, data.Series,
		func(p *appdto.AppResourcePointResp) time.Time { return p.Time },
		func(p *appdto.AppResourcePointResp, n *int) { p.Replicas = n })
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appdto.GetAppResourceMetricsResp{Data: data}, nil
}

func toAppResourceMetricsData(
	rangeName string, w window, stats *logging.ResourceStatsResp,
) *appdto.AppResourceMetricsDataResp {
	data := &appdto.AppResourceMetricsDataResp{
		Available:   true,
		Range:       rangeName,
		Start:       w.start,
		End:         w.end,
		StepSeconds: int(w.step.Seconds()),
		Clamped:     w.clamped,
		Totals:      &appdto.AppResourceTotalsResp{},
		Series:      make([]*appdto.AppResourcePointResp, 0, len(stats.Buckets)),
		Containers:  make([]*appdto.AppResourceContainerResp, 0, len(stats.Containers)),
	}
	var cpuSum float64
	var cpuSteps int
	for _, b := range stats.Buckets {
		data.Series = append(data.Series, &appdto.AppResourcePointResp{
			Time: b.Time, CPU: b.CPU, CPULimit: b.CPULimit, Memory: b.Memory, MemoryLimit: b.MemoryLimit,
			OOMKills: b.OOMKills, NetRx: b.NetRx, NetTx: b.NetTx, IORead: b.IORead, IOWrite: b.IOWrite,
		})
		data.Totals.OOMKills += b.OOMKills
		if b.CPU != nil {
			cpuSum += *b.CPU
			cpuSteps++
			data.Totals.CPUPeak = maxOf(data.Totals.CPUPeak, *b.CPU)
			data.Totals.CPULimit = b.CPULimit
		}
		if b.Memory != nil {
			data.Totals.MemoryPeak = maxOf(data.Totals.MemoryPeak, *b.Memory)
			data.Totals.MemoryLimit = b.MemoryLimit
		}
	}
	if cpuSteps > 0 {
		avg := cpuSum / float64(cpuSteps)
		data.Totals.CPU = &avg
	}
	for _, c := range stats.Containers {
		data.Containers = append(data.Containers, &appdto.AppResourceContainerResp{
			Container: c.Container, CPU: c.CPU, CPUPeak: c.CPUPeak, Memory: c.Memory,
			MemoryLimit: c.MemoryLimit, OOMKills: c.OOMKills, LastSeen: c.LastSeen,
		})
	}
	return data
}

func maxOf(current *float64, v float64) *float64 {
	if current == nil || v > *current {
		return &v
	}
	return current
}
