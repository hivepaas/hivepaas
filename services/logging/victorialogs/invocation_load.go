package victorialogs

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// BuildInvocationLoadQuery turns a request into LogsQL by the rules BuildQuery
// keeps: every value from the request in a Go-quoted literal, the structure
// fixed here. One query for every function asked about, grouped by app: its
// cost does not grow with them.
//
// A call counts for the range at most: one that ran longer was in flight for
// the whole range, not more - its whole duration, ending in it, would read as
// many calls at once.
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
	span, err := loadSpan(req.Start, req.End)
	if err != nil {
		return "", err
	}
	short, err := shortPart(req.Start, req.ShortStart, req.End)
	if err != nil {
		return "", err
	}
	field := strconv.Quote(req.Field)
	duration := strconv.Quote(UnpackPrefix + "durationMs")
	capped := strconv.Quote(UnpackPrefix + "cappedMs")
	q := field + `:in(` + strings.Join(ids, ",") + `) AND ` + strconv.Quote(invocationPhrase) +
		` | unpack_json from _msg fields (hp, outcome, durationMs) result_prefix ` + strconv.Quote(UnpackPrefix) +
		` | filter ` + strconv.Quote(UnpackPrefix+"hp") + `:="invocation"` +
		` | math min(` + duration + `, ` + strconv.FormatInt(span.Milliseconds(), 10) + `) as ` + capped
	if short != nil {
		q += ` | math min(` + duration + `, ` + strconv.FormatInt(short.span.Milliseconds(), 10) + `) as ` +
			strconv.Quote(UnpackPrefix+"cappedShortMs")
	}
	// min() of a duration that is not a number is the range: only numbers
	// are summed.
	q += ` | stats by (` + field + `) sum(` + capped + `) if (` + duration + `:>=0) busyMs, count() calls,` +
		` count() if (` + strconv.Quote(UnpackPrefix+"outcome") + `:="throttled") throttled`
	if short != nil {
		q += `, sum(` + strconv.Quote(UnpackPrefix+"cappedShortMs") + `) if (` + short.filter + ` ` + duration +
			`:>=0) shortBusyMs, count() if (` + short.filter + `) shortCalls`
	}
	return q, nil
}

// shortRange is a load query's last part: its span, and the filter on a
// line's time that keeps it.
type shortRange struct {
	span   time.Duration
	filter string
}

// shortPart is the range's last part from shortStart, nil when there is none:
// it must lie inside the range. The time is written as LogsQL reads one, not
// quoted: it is no text of anyone's.
func shortPart(start, shortStart, end time.Time) (*shortRange, error) {
	if shortStart.IsZero() {
		return nil, nil //nolint:nilnil // no part asked for
	}
	if !shortStart.After(start) || !shortStart.Before(end) {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the range's last part must lie in it")
	}
	return &shortRange{span: end.Sub(shortStart),
		filter: `_time:>=` + shortStart.UTC().Format(time.RFC3339Nano)}, nil
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
			ShortBusyMs: value(row["shortBusyMs"]), ShortCalls: whole(row["shortCalls"]),
		}
	}
	return out, nil
}

// loadSpan is a load query's range, which must end after it starts: what a
// request or a call is counted for at most.
func loadSpan(start, end time.Time) (time.Duration, error) {
	span := end.Sub(start)
	if span <= 0 {
		return 0, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the range must end after it starts")
	}
	return span, nil
}
