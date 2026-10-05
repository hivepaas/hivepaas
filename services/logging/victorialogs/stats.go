package victorialogs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// invocationPhrase narrows the lines to unpack to those that may be invocation
// lines; the filter after the unpacking is what decides.
const invocationPhrase = `"hp":"invocation"`

var (
	methodField = strconv.Quote(UnpackPrefix + "method")
	pathField   = strconv.Quote(UnpackPrefix + "path")
)

// invocationCounts are the counts every stats query asks for.
const invocationCounts = ` count() calls, count() if ("app.outcome":!="ok") failed,` +
	` count() if ("app.status":>=400 "app.status":<500) errors4xx,` +
	` count() if ("app.status":>=500) errors5xx, quantile(0.5, "app.durationMs") p50,` +
	` quantile(0.95, "app.durationMs") p95, quantile(0.99, "app.durationMs") p99`

// InvocationStatsQueries are the four queries a function's metrics take: by
// step, in all, by outcome, and by method and path.
type InvocationStatsQueries struct {
	Series   string
	Totals   string
	Outcomes string
	Paths    string
}

// BuildInvocationStatsQueries turns a request into LogsQL by the rules
// BuildQuery keeps: every value from the request in a Go-quoted literal, the
// structure fixed here, the unpacked fields under UnpackPrefix.
func BuildInvocationStatsQueries(req *loggingmodel.InvocationStatsReq) (*InvocationStatsQueries, error) {
	if len(req.Match) == 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Step < time.Second || req.Step%time.Second != 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the step must be whole seconds")
	}
	if req.TopPaths <= 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the paths must be a number of them")
	}
	filters := make([]string, 0, len(req.Match)+1)
	for _, m := range req.Match {
		if m.Field == "" {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	filters = append(filters, strconv.Quote(invocationPhrase))
	head := strings.Join(filters, " AND ") +
		` | unpack_json from _msg fields (hp, method, path, outcome, status, durationMs) result_prefix ` +
		strconv.Quote(UnpackPrefix) +
		` | filter ` + strconv.Quote(UnpackPrefix+"hp") + `:="invocation"`
	return &InvocationStatsQueries{
		Series:   fmt.Sprintf("%s | stats by (_time:%ds)%s | sort by (_time)", head, req.Step/time.Second, invocationCounts),
		Totals:   head + " | stats" + invocationCounts,
		Outcomes: head + ` | stats by (` + strconv.Quote(UnpackPrefix+"outcome") + `) count() calls`,
		// Ties keep one order, so that the same calls list the same paths.
		Paths: fmt.Sprintf("%s | stats by (%s, %s)%s | sort by (calls desc, %s, %s) limit %d", head,
			methodField, pathField, invocationCounts, pathField, methodField, req.TopPaths),
	}, nil
}

// InvocationStats counts a function's invocation lines: four queries, built
// by BuildInvocationStatsQueries, over the request's range.
func (c *Client) InvocationStats(
	ctx context.Context, req *loggingmodel.InvocationStatsReq,
) (*loggingmodel.InvocationStatsResp, error) {
	q, err := BuildInvocationStatsQueries(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingmodel.InvocationStatsResp{ByOutcome: map[string]int64{}}

	rows, err := c.rows(ctx, q.Series, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		at, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("a bucket's time: %s", row["_time"])
		}
		out.Buckets = append(out.Buckets, &loggingmodel.InvocationBucket{Time: at, InvocationCounts: countsOf(row)})
	}
	sort.Slice(out.Buckets, func(i, j int) bool { return out.Buckets[i].Time.Before(out.Buckets[j].Time) })

	rows, err = c.rows(ctx, q.Totals, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(rows) > 0 {
		out.Totals = countsOf(rows[0])
	}

	rows, err = c.rows(ctx, q.Outcomes, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		if calls := whole(row["calls"]); calls > 0 {
			out.ByOutcome[row[UnpackPrefix+"outcome"]] = calls
		}
	}

	rows, err = c.rows(ctx, q.Paths, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		out.ByPath = append(out.ByPath, &loggingmodel.InvocationPath{
			Method: row[UnpackPrefix+"method"], Path: row[UnpackPrefix+"path"], InvocationCounts: countsOf(row),
		})
	}
	return out, nil
}

// rows runs a stats query; VictoriaLogs answers one JSON object per line, every
// value a string.
func (c *Client) rows(ctx context.Context, q string, start, end time.Time) ([]map[string]string, error) {
	resp, err := c.post(ctx, q, start, end)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out []map[string]string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		row := map[string]string{}
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("a row is not JSON: %v", err)
		}
		out = append(out, row)
	}
	return out, hperrors.Wrap(scanner.Err())
}

// rowsAtOnce runs stats queries over one range at once, their rows in the
// queries' order. The first to fail stops the others, and is the error.
func (c *Client) rowsAtOnce(
	ctx context.Context, start, end time.Time, queries ...string,
) ([][]map[string]string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make([][]map[string]string, len(queries))
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	for i, q := range queries {
		wg.Go(func() {
			// A panic here is this query's error, as a failure is.
			var err error
			defer func() {
				if err != nil {
					once.Do(func() { first = err; cancel() })
				}
			}()
			defer safego.RecoverTo(&err)
			var rows []map[string]string
			if rows, err = c.rows(ctx, q, start, end); err == nil {
				results[i] = rows
			}
		})
	}
	wg.Wait()
	if first != nil {
		return nil, hperrors.Wrap(first)
	}
	return results, nil
}

func countsOf(row map[string]string) loggingmodel.InvocationCounts {
	counts := loggingmodel.InvocationCounts{
		Calls: whole(row["calls"]), Failed: whole(row["failed"]),
		Errors4xx: whole(row["errors4xx"]), Errors5xx: whole(row["errors5xx"]),
	}
	if counts.Calls > 0 {
		counts.P50, counts.P95, counts.P99 = number(row["p50"]), number(row["p95"]), number(row["p99"])
	}
	return counts
}

func whole(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// number is a quantile; nil for what is no number - NaN, or nothing.
func number(s string) *float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	return &f
}
