package loggingserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// functionMetricsTopPaths is how many paths a function's metrics count apart:
// the most called, enough to tell a function's own paths from a scanner's.
const functionMetricsTopPaths = 20

// FunctionMetrics counts a function's calls in its stored logs.
//
// The scope is app's, as for QueryAppLogs. The backend leaves out the steps
// without a call; they are put back here, so that every step of the range is a
// point.
func (s *service) FunctionMetrics(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
) (*logging.InvocationStatsResp, error) {
	if app == nil || app.ID == "" {
		// An empty value would count every line that carries no app at all.
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := backend.InvocationStats(ctx, &logging.InvocationStatsReq{
		Match:    []logging.FieldMatch{{Field: vlagent.AttrField(appservice.LabelLogAppID), Value: app.ID}},
		Start:    q.Start,
		End:      q.End,
		Step:     q.Step,
		TopPaths: functionMetricsTopPaths,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Buckets = everyStep(resp.Buckets, q.Start, q.End, q.Step)
	return resp, nil
}

// everyStep is one bucket per step of [start, end), the backend's where it has
// one, an empty one where it has none.
func everyStep(
	buckets []*logging.InvocationBucket, start, end time.Time, step time.Duration,
) []*logging.InvocationBucket {
	byTime := make(map[int64]*logging.InvocationBucket, len(buckets))
	for _, b := range buckets {
		byTime[b.Time.Unix()] = b
	}
	out := make([]*logging.InvocationBucket, 0, int(end.Sub(start)/step))
	for at := start; at.Before(end); at = at.Add(step) {
		if b, ok := byTime[at.Unix()]; ok {
			out = append(out, b)
			continue
		}
		out = append(out, &logging.InvocationBucket{Time: at})
	}
	return out
}
