package victorialogs

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// OBIUnpackPrefix is where the agent's OBI rows' fields are unpacked: apart
// from any app's own.
const OBIUnpackPrefix = "obi."

// obiField names a row's field as written, a name of the agent's own: letters,
// digits, dots - never a request's text.
var obiField = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9.]*$`)

func obiQuoted(name string) string { return strconv.Quote(OBIUnpackPrefix + name) }

// OBIStatsQueries are the two queries an app's OBI rows take: by step, and by
// group.
type OBIStatsQueries struct {
	Series string
	Groups string
}

// BuildOBIStatsQueries turns a request into LogsQL by the rules BuildQuery
// keeps: every value from the request in a Go-quoted literal, the structure
// fixed here; field names only the agent's own.
func BuildOBIStatsQueries(req *loggingmodel.OBIStatsReq) (*OBIStatsQueries, error) {
	if len(req.Match) == 0 || req.AppID == "" || req.Row == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Step < time.Second || req.Step%time.Second != 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the step must be whole seconds")
	}
	if req.TopGroups <= 0 || len(req.GroupBy) == 0 || len(req.BucketFields) == 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).
			WithExtraDetail("the groups and the buckets must be given")
	}
	fields := append([]string{"hp", "app", "count", "errors", "sumMs"}, req.GroupBy...)
	for _, f := range req.SeriesBy {
		if !slices.Contains(fields, f) {
			fields = append(fields, f)
		}
	}
	fields = append(fields, req.BucketFields...)
	for _, f := range fields {
		if !obiField.MatchString(f) {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("not a row's field: %q", f)
		}
	}
	head, err := loadMatch(req.Match, `"hp":"`+req.Row+`"`)
	if err != nil {
		return nil, err
	}
	head += ` | unpack_json from _msg fields (` + strings.Join(fields, ", ") + `) result_prefix ` +
		strconv.Quote(OBIUnpackPrefix) +
		` | filter ` + obiQuoted("hp") + `:=` + strconv.Quote(req.Row) + ` ` + obiQuoted("app") + `:=` +
		strconv.Quote(req.AppID)

	sums := make([]string, 0, 3+len(req.BucketFields)) //nolint:mnd // count, errors, sumMs
	sums = append(sums, `sum(`+obiQuoted("count")+`) count`, `sum(`+obiQuoted("errors")+`) errors`,
		`sum(`+obiQuoted("sumMs")+`) sumMs`)
	for _, b := range req.BucketFields {
		sums = append(sums, `sum(`+obiQuoted(b)+`) `+strconv.Quote(b))
	}
	stats := strings.Join(sums, ", ")
	quoted := func(names []string) string {
		out := make([]string, 0, len(names))
		for _, name := range names {
			out = append(out, obiQuoted(name))
		}
		return strings.Join(out, ", ")
	}
	by := fmt.Sprintf("_time:%ds", req.Step/time.Second)
	if len(req.SeriesBy) > 0 {
		by += ", " + quoted(req.SeriesBy)
	}
	groups := quoted(req.GroupBy)
	return &OBIStatsQueries{
		Series: fmt.Sprintf("%s | stats by (%s) %s | sort by (_time)", head, by, stats),
		// Ties keep one order, so that the same rows list the same groups.
		Groups: fmt.Sprintf("%s | stats by (%s) %s | sort by (count desc, %s) limit %d", head, groups, stats,
			groups, req.TopGroups),
	}, nil
}

// OBIStats sums an app's OBI rows over the request's range: two queries,
// built by BuildOBIStatsQueries.
func (c *Client) OBIStats(ctx context.Context, req *loggingmodel.OBIStatsReq) (*loggingmodel.OBIStatsResp, error) {
	q, err := BuildOBIStatsQueries(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingmodel.OBIStatsResp{}
	rows, err := c.rows(ctx, q.Series, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		at, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("a bucket's time: %s", row["_time"])
		}
		out.Buckets = append(out.Buckets, &loggingmodel.OBIBucket{Time: at, Keys: keysOf(row, req.SeriesBy),
			OBIHistogram: histogramOf(row, req)})
	}
	sort.SliceStable(out.Buckets, func(i, j int) bool { return out.Buckets[i].Time.Before(out.Buckets[j].Time) })

	if rows, err = c.rows(ctx, q.Groups, req.Start, req.End); err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		out.Groups = append(out.Groups, &loggingmodel.OBIGroup{Keys: keysOf(row, req.GroupBy),
			OBIHistogram: histogramOf(row, req)})
	}
	return out, nil
}

// keysOf are a row's values of fields it was summed by; nil for none.
func keysOf(row map[string]string, fields []string) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	keys := make(map[string]string, len(fields))
	for _, f := range fields {
		keys[f] = row[OBIUnpackPrefix+f]
	}
	return keys
}

func histogramOf(row map[string]string, req *loggingmodel.OBIStatsReq) loggingmodel.OBIHistogram {
	h := loggingmodel.OBIHistogram{Count: wholeOf(row["count"]), Errors: wholeOf(row["errors"]),
		SumMs: value(row["sumMs"]), Buckets: make(map[string]int64, len(req.BucketFields))}
	for _, b := range req.BucketFields {
		h.Buckets[b] = wholeOf(row[b])
	}
	return h
}

// wholeOf is a sum's whole number: VictoriaLogs may write it as a float.
func wholeOf(s string) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	return int64(value(s))
}

// BuildOBIStatusQuery is the nodes' status rows, the latest first.
func BuildOBIStatusQuery(req *loggingmodel.OBIStatusReq) (string, error) {
	if req.Limit <= 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the rows must be a number of them")
	}
	head, err := loadMatch(req.Match, `"hp":"obi"`)
	if err != nil {
		return "", err
	}
	return head + ` | unpack_json from _msg fields (hp) result_prefix ` + strconv.Quote(OBIUnpackPrefix) +
		` | filter ` + obiQuoted("hp") + `:="obi" | fields _time, _msg | sort by (_time desc) limit ` +
		strconv.Itoa(req.Limit), nil
}

// OBIStatus reads the nodes' status rows over the request's range.
func (c *Client) OBIStatus(ctx context.Context, req *loggingmodel.OBIStatusReq) ([]*loggingmodel.OBIStatusRow, error) {
	q, err := BuildOBIStatusQuery(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	rows, err := c.rows(ctx, q, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := make([]*loggingmodel.OBIStatusRow, 0, len(rows))
	for _, row := range rows {
		at, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			continue
		}
		out = append(out, &loggingmodel.OBIStatusRow{Time: at, Msg: row["_msg"]})
	}
	return out, nil
}
