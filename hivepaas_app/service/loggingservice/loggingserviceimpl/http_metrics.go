package loggingserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

const (
	// httpMetricsTopPaths is how many paths an app's HTTP numbers count apart,
	// as a function's do.
	httpMetricsTopPaths = 20
	// httpMetricsTopReplicas is how many replicas: more than an app runs.
	httpMetricsTopReplicas = 20
)

// ProxyHistory reads the logging settings as AppHistory does, without an app:
// the proxy is one of HivePaaS's own containers, collected with the apps' or
// with HivePaaS's.
func (s *service) ProxyHistory(ctx context.Context, db database.IDB) (*loggingservice.AppHistory, error) {
	cfg, err := s.loadEnabledSettings(ctx, db)
	if errors.Is(err, hperrors.ErrLoggingNotEnabled) {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Sources.Apps && !cfg.Sources.HivePaaS {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonAppsNotCollected}, nil
	}
	if !hasQueryEndpoint(cfg) {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonNoQueryEndpoint}, nil
	}
	return &loggingservice.AppHistory{Available: true, Retention: retentionOf(cfg)}, nil
}

// HTTPMetrics implements loggingservice.Service. The lines are the proxy's by
// the identity the daemon wrote into them, never by what they say; the app's
// by its services' names, which are by its id.
func (s *service) HTTPMetrics(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
) (*logging.HTTPStatsResp, error) {
	if app == nil || app.ID == "" {
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.HTTPStats(ctx, &logging.HTTPStatsReq{
		Match: []logging.FieldMatch{{
			Field: vlagent.AttrField(base.LabelLogComponent), Value: base.LogComponentTraefik,
		}},
		ServicePattern: traefikservice.AppHTTPServicePattern(app.ID),
		ServicePhrase:  traefikservice.AppHTTPServicePhrase(app.ID),
		Start:          q.Start,
		End:            q.End,
		Step:           q.Step,
		TopPaths:       httpMetricsTopPaths,
		TopReplicas:    httpMetricsTopReplicas,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Buckets = everyHTTPStep(resp.Buckets, q.Start, q.End, q.Step)
	return resp, nil
}

// resourceMetricsTopContainers is how many of an app's containers are listed:
// its replicas, and those replaced within the range.
const resourceMetricsTopContainers = 50

// ResourceMetrics implements loggingservice.Service. The rows are the agent's
// by the identity the daemon wrote into them, never by what they say; the
// app's by the id the agent wrote, from the container's own label.
func (s *service) ResourceMetrics(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
) (*logging.ResourceStatsResp, error) {
	if app == nil || app.ID == "" {
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.ResourceStats(ctx, &logging.ResourceStatsReq{
		Match: []logging.FieldMatch{{
			Field: vlagent.AttrField(base.LabelLogComponent), Value: base.LogComponentAgent,
		}},
		AppID:         app.ID,
		Start:         q.Start,
		End:           q.End,
		Step:          q.Step,
		TopContainers: resourceMetricsTopContainers,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Buckets = everyResourceStep(resp.Buckets, q.Start, q.End, q.Step)
	return resp, nil
}

// FunctionLoad implements loggingservice.Service. The calls are each
// function's by the app identity the daemon wrote into its lines.
func (s *service) FunctionLoad(
	ctx context.Context,
	db database.IDB,
	appIDs []string,
	start, end, shortStart time.Time,
) (map[string]*logging.InvocationLoad, error) {
	if len(appIDs) == 0 {
		return map[string]*logging.InvocationLoad{}, nil
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.InvocationLoad(ctx, &logging.InvocationLoadReq{
		Field: vlagent.AttrField(appservice.LabelLogAppID), AppIDs: appIDs, Start: start, End: end,
		ShortStart: shortStart,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.ByApp, nil
}

// RequestLoad implements loggingservice.Service. The lines are the proxy's by
// the identity the daemon wrote into them; each app's by its services' names.
func (s *service) RequestLoad(
	ctx context.Context,
	db database.IDB,
	appIDs []string,
	start, end, shortStart time.Time,
) (map[string]*logging.RequestLoad, error) {
	if len(appIDs) == 0 {
		return map[string]*logging.RequestLoad{}, nil
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.RequestLoad(ctx, &logging.RequestLoadReq{
		Match: []logging.FieldMatch{{
			Field: vlagent.AttrField(base.LabelLogComponent), Value: base.LogComponentTraefik,
		}},
		AppIDs: appIDs, Start: start, End: end, ShortStart: shortStart,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.ByApp, nil
}

// CPULoad implements loggingservice.Service. The rows are the agent's by the
// identity the daemon wrote into them; each app's by the id the agent wrote.
func (s *service) CPULoad(
	ctx context.Context,
	db database.IDB,
	appIDs []string,
	start, end time.Time,
) (map[string][]*logging.ContainerCPU, error) {
	if len(appIDs) == 0 {
		return map[string][]*logging.ContainerCPU{}, nil
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.CPULoad(ctx, &logging.CPULoadReq{
		Match: []logging.FieldMatch{{
			Field: vlagent.AttrField(base.LabelLogComponent), Value: base.LogComponentAgent,
		}},
		AppIDs: appIDs, Start: start, End: end,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.ByApp, nil
}

// everyResourceStep is everyStep for usage.
func everyResourceStep(
	buckets []*logging.ResourceBucket, start, end time.Time, step time.Duration,
) []*logging.ResourceBucket {
	byTime := make(map[int64]*logging.ResourceBucket, len(buckets))
	for _, b := range buckets {
		byTime[b.Time.Unix()] = b
	}
	out := make([]*logging.ResourceBucket, 0, int(end.Sub(start)/step))
	for at := start; at.Before(end); at = at.Add(step) {
		if b, ok := byTime[at.Unix()]; ok {
			out = append(out, b)
			continue
		}
		out = append(out, &logging.ResourceBucket{Time: at})
	}
	return out
}

// everyHTTPStep is everyStep for requests.
func everyHTTPStep(buckets []*logging.HTTPBucket, start, end time.Time, step time.Duration) []*logging.HTTPBucket {
	byTime := make(map[int64]*logging.HTTPBucket, len(buckets))
	for _, b := range buckets {
		byTime[b.Time.Unix()] = b
	}
	out := make([]*logging.HTTPBucket, 0, int(end.Sub(start)/step))
	for at := start; at.Before(end); at = at.Add(step) {
		if b, ok := byTime[at.Unix()]; ok {
			out = append(out, b)
			continue
		}
		out = append(out, &logging.HTTPBucket{Time: at})
	}
	return out
}
