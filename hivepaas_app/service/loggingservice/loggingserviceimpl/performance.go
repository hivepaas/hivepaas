package loggingserviceimpl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

const (
	// routeMetricsTopRoutes is how many of an app's routes are listed: more
	// than a page shows.
	routeMetricsTopRoutes = 100
	// dependencyMetricsTopCalls is how many of its peers' operations: an app
	// calls a few peers, each a few ways.
	dependencyMetricsTopCalls = 200
	// performanceStatusSince is how far back the nodes' status rows are read:
	// the agent writes one a minute, so a node without one in three is one
	// whose agent is not writing them.
	performanceStatusSince = 3 * time.Minute
	// performanceStatusRows is at most how many: three a node, for over a
	// hundred nodes.
	performanceStatusRows = 500
)

// agentMatch is the agent's lines, by the identity the daemon wrote into
// them: never by what they say.
func agentMatch() []logging.FieldMatch {
	return []logging.FieldMatch{{Field: vlagent.AttrField(base.LabelLogComponent), Value: base.LogComponentAgent}}
}

// RouteMetrics implements loggingservice.Service.
func (s *service) RouteMetrics(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
) (*logging.OBIStatsResp, error) {
	return s.obiStats(ctx, db, app, q, &logging.OBIStatsReq{
		Row: obi.RowRoutes, GroupBy: []string{obi.FieldKind, obi.FieldMethod, obi.FieldRoute},
		TopGroups: routeMetricsTopRoutes,
	})
}

// DependencyMetrics implements loggingservice.Service.
func (s *service) DependencyMetrics(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
) (*logging.OBIStatsResp, error) {
	return s.obiStats(ctx, db, app, q, &logging.OBIStatsReq{
		Row:       obi.RowCalls,
		GroupBy:   []string{obi.FieldKind, obi.FieldPeer, obi.FieldMethod, obi.FieldOperation},
		SeriesBy:  []string{obi.FieldKind},
		TopGroups: dependencyMetricsTopCalls,
	})
}

// obiStats sums an app's rows of one kind: the agent's identity and the app's
// id are put into the query here, whatever req says of them.
func (s *service) obiStats(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.FunctionMetricsQuery,
	req *logging.OBIStatsReq,
) (*logging.OBIStatsResp, error) {
	if app == nil || app.ID == "" {
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	req.Match, req.AppID, req.BucketFields = agentMatch(), app.ID, obi.BucketFields()
	req.Start, req.End, req.Step = q.Start, q.End, q.Step
	resp, err := backend.OBIStats(ctx, req)
	return resp, hperrors.Wrap(err)
}

// PerformanceStatus implements loggingservice.Service.
func (s *service) PerformanceStatus(
	ctx context.Context,
	db database.IDB,
) (map[string]*loggingservice.PerformanceNodeStatus, error) {
	backend, err := s.queryBackend(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	now := timeutil.NowUTC()
	rows, err := backend.OBIStatus(ctx, &logging.OBIStatusReq{
		Match: agentMatch(), Start: now.Add(-performanceStatusSince), End: now.Add(time.Second),
		Limit: performanceStatusRows,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return latestStatuses(rows), nil
}

// latestStatuses is each node's latest status, from rows the latest first; a
// row that is no status is skipped.
func latestStatuses(rows []*logging.OBIStatusRow) map[string]*loggingservice.PerformanceNodeStatus {
	out := map[string]*loggingservice.PerformanceNodeStatus{}
	for _, row := range rows {
		var status obi.Status
		if json.Unmarshal([]byte(row.Msg), &status) != nil || status.HP != obi.RowStatus || status.Node == "" {
			continue
		}
		if _, seen := out[status.Node]; !seen {
			out[status.Node] = &loggingservice.PerformanceNodeStatus{Time: row.Time, Status: status}
		}
	}
	return out
}
