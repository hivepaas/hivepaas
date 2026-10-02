package victorialogs

import (
	"context"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// BuildInvocationLoadQuery turns a request into LogsQL by the rules BuildQuery
// keeps: every value from the request in a Go-quoted literal, the structure
// fixed here. One query for every function asked about, grouped by app: its
// cost does not grow with them.
func BuildInvocationLoadQuery(req *loggingmodel.InvocationLoadReq) (string, error) {
	if req.Field == "" || len(req.AppIDs) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	ids := make([]string, 0, len(req.AppIDs))
	for _, id := range req.AppIDs {
		if id == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		ids = append(ids, strconv.Quote(id))
	}
	field := strconv.Quote(req.Field)
	return field + `:in(` + strings.Join(ids, ",") + `) AND ` + strconv.Quote(invocationPhrase) +
		` | unpack_json from _msg fields (hp, outcome, durationMs) result_prefix ` + strconv.Quote(UnpackPrefix) +
		` | filter ` + strconv.Quote(UnpackPrefix+"hp") + `:="invocation"` +
		` | stats by (` + field + `) sum(` + strconv.Quote(UnpackPrefix+"durationMs") + `) busyMs, count() calls,` +
		` count() if (` + strconv.Quote(UnpackPrefix+"outcome") + `:="throttled") throttled`, nil
}

// InvocationLoad says how busy functions were over the request's range.
func (c *Client) InvocationLoad(
	ctx context.Context, req *loggingmodel.InvocationLoadReq,
) (*loggingmodel.InvocationLoadResp, error) {
	q, err := BuildInvocationLoadQuery(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	rows, err := c.rows(ctx, q, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingmodel.InvocationLoadResp{ByApp: make(map[string]*loggingmodel.InvocationLoad, len(rows))}
	for _, row := range rows {
		out.ByApp[row[req.Field]] = &loggingmodel.InvocationLoad{
			BusyMs: value(row["busyMs"]), Calls: whole(row["calls"]), Throttled: whole(row["throttled"]),
		}
	}
	return out, nil
}
