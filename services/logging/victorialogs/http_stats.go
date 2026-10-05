package victorialogs

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// HTTPUnpackPrefix is where an access log line's fields are unpacked to, apart
// from the line's own and from an app's.
const HTTPUnpackPrefix = "http."

// accessLogPhrase narrows the proxy's lines to its access log lines; the
// service filter after the unpacking is what decides.
const accessLogPhrase = "ServiceName"

var (
	httpService  = strconv.Quote(HTTPUnpackPrefix + "ServiceName")
	httpReplica  = strconv.Quote(HTTPUnpackPrefix + "ServiceURL")
	httpMethod   = strconv.Quote(HTTPUnpackPrefix + "RequestMethod")
	httpPath     = strconv.Quote(HTTPUnpackPrefix + "RequestPath")
	httpStatus   = strconv.Quote(HTTPUnpackPrefix + "DownstreamStatus")
	httpOrigin   = strconv.Quote(HTTPUnpackPrefix + "OriginStatus")
	httpDuration = strconv.Quote(HTTPUnpackPrefix + "Duration")
	httpMs       = strconv.Quote(HTTPUnpackPrefix + "ms")
)

// httpCounts are the counts every HTTP stats query asks for. Unreachable is a
// gateway error with no status from the app: the proxy found no replica, or
// none answered.
var httpCounts = fmt.Sprintf(` count() requests,`+
	` count() if (%[1]s:>=400 %[1]s:<500) errors4xx, count() if (%[1]s:>=500) errors5xx,`+
	` count() if (%[2]s:=0 %[1]s:>=502 %[1]s:<=504) unreachable,`+
	` quantile(0.5, %[3]s) p50, quantile(0.95, %[3]s) p95, quantile(0.99, %[3]s) p99`,
	httpStatus, httpOrigin, httpMs)

// normalizePaths replaces a path's ids and numbers - a UUID, a segment of
// digits - so that /users/1 and /users/2 count as one. Numbers run twice: a
// replaced segment's slash is not there for the next to match.
var normalizePaths = fmt.Sprintf(` | replace_regexp (%[2]s, "/:id$1") at %[1]s`+
	` | replace_regexp ("/[0-9]+(/|$)", "/:n$1") at %[1]s | replace_regexp ("/[0-9]+(/|$)", "/:n$1") at %[1]s`,
	httpPath, strconv.Quote(uuidSegment))

// uuidSegment is a path segment that is a UUID.
const uuidSegment = `/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}(/|$)`

// HTTPStatsQueries are the four queries an app's HTTP numbers take: by step,
// in all, by method and path, and by replica.
type HTTPStatsQueries struct {
	Series   string
	Totals   string
	Paths    string
	Replicas string
}

// BuildHTTPStatsQueries turns a request into LogsQL by the rules BuildQuery
// keeps: every value from the request in a Go-quoted literal, the structure
// fixed here, the unpacked fields under HTTPUnpackPrefix.
func BuildHTTPStatsQueries(req *loggingmodel.HTTPStatsReq) (*HTTPStatsQueries, error) {
	if len(req.Match) == 0 || req.ServicePattern == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Step < time.Second || req.Step%time.Second != 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the step must be whole seconds")
	}
	if req.TopPaths <= 0 || req.TopReplicas <= 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).
			WithExtraDetail("the paths and replicas must be a number of them")
	}
	filters := make([]string, 0, len(req.Match)+1)
	for _, m := range req.Match {
		if m.Field == "" {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	filters = append(filters, strconv.Quote(accessLogPhrase))
	if req.ServicePhrase != "" {
		filters = append(filters, strconv.Quote(req.ServicePhrase))
	}
	head := strings.Join(filters, " AND ") +
		` | unpack_json from _msg fields (ServiceName, ServiceURL, RequestMethod, RequestPath,` +
		` DownstreamStatus, OriginStatus, Duration) result_prefix ` + strconv.Quote(HTTPUnpackPrefix) +
		` | filter ` + httpService + `:~` + strconv.Quote(req.ServicePattern) +
		` | math ` + httpDuration + ` / 1000000 as ` + httpMs
	return &HTTPStatsQueries{
		Series: fmt.Sprintf("%s | stats by (_time:%ds)%s | sort by (_time)", head, req.Step/time.Second, httpCounts),
		Totals: head + " | stats" + httpCounts,
		// Ties keep one order, so that the same requests list the same rows.
		Paths: fmt.Sprintf("%s%s | stats by (%s, %s)%s | sort by (requests desc, %s, %s) limit %d", head,
			normalizePaths, httpMethod, httpPath, httpCounts, httpPath, httpMethod, req.TopPaths),
		Replicas: fmt.Sprintf("%s | stats by (%s)%s | sort by (requests desc, %s) limit %d", head,
			httpReplica, httpCounts, httpReplica, req.TopReplicas),
	}, nil
}

// HTTPStats counts an app's access log lines: four queries, built by
// BuildHTTPStatsQueries, over the request's range. They run at once: each
// reads the range's lines anew, and one after another an app's busy hour took
// the four's time added up.
func (c *Client) HTTPStats(ctx context.Context, req *loggingmodel.HTTPStatsReq) (*loggingmodel.HTTPStatsResp, error) {
	q, err := BuildHTTPStatsQueries(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	results, err := c.rowsAtOnce(ctx, req.Start, req.End, q.Series, q.Totals, q.Paths, q.Replicas)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	series, totals, paths, replicas := results[0], results[1], results[2], results[3]
	out := &loggingmodel.HTTPStatsResp{}

	for _, row := range series {
		at, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("a bucket's time: %s", row["_time"])
		}
		out.Buckets = append(out.Buckets, &loggingmodel.HTTPBucket{Time: at, HTTPCounts: httpCountsOf(row)})
	}
	sort.Slice(out.Buckets, func(i, j int) bool { return out.Buckets[i].Time.Before(out.Buckets[j].Time) })

	if len(totals) > 0 {
		out.Totals = httpCountsOf(totals[0])
	}

	for _, row := range paths {
		out.ByPath = append(out.ByPath, &loggingmodel.HTTPPath{
			Method: row[HTTPUnpackPrefix+"RequestMethod"], Path: row[HTTPUnpackPrefix+"RequestPath"],
			HTTPCounts: httpCountsOf(row),
		})
	}

	for _, row := range replicas {
		out.ByReplica = append(out.ByReplica, &loggingmodel.HTTPReplica{
			Address: row[HTTPUnpackPrefix+"ServiceURL"], HTTPCounts: httpCountsOf(row),
		})
	}
	return out, nil
}

func httpCountsOf(row map[string]string) loggingmodel.HTTPCounts {
	counts := loggingmodel.HTTPCounts{
		Requests: whole(row["requests"]), Errors4xx: whole(row["errors4xx"]),
		Errors5xx: whole(row["errors5xx"]), Unreachable: whole(row["unreachable"]),
	}
	if counts.Requests > 0 {
		counts.P50, counts.P95, counts.P99 = number(row["p50"]), number(row["p95"]), number(row["p99"])
	}
	return counts
}
