package victorialogs

import (
	"context"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// loadIDsPerQuery is how many apps one load query asks about at most:
// VictoriaLogs refuses a query longer than its -search.maxQueryLen, 16 KB
// unless told otherwise, and one HivePaaS did not deploy may not be. 200
// app ids, ULIDs, twice each in the request query, come to 13.5 KB. Each
// query reads the whole range again: no fewer to a query than fit.
const loadIDsPerQuery = 200

// inChunks reads the apps a query's worth at a time, the results of each into
// the same answer. None is read once, for the query to refuse.
func inChunks(ids []string, read func(ids []string) error) error {
	if len(ids) == 0 {
		return read(ids)
	}
	for chunk := range slices.Chunk(ids, loadIDsPerQuery) {
		if err := read(chunk); err != nil {
			return err
		}
	}
	return nil
}

// anyPhrase is a filter of the lines holding one of the phrases, quoted: the
// block of lines that holds none is passed over by its bloom filter, before
// a line of it is unpacked.
func anyPhrase(phrases []string) string {
	quoted := make([]string, 0, len(phrases))
	for _, phrase := range phrases {
		quoted = append(quoted, strconv.Quote(phrase))
	}
	return "(" + strings.Join(quoted, " OR ") + ")"
}

// loadMatch is the filters every load query starts with: the identity the
// daemon wrote into the component's lines, then the phrase its lines carry.
func loadMatch(match []loggingmodel.FieldMatch, phrase string) (string, error) {
	if len(match) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	filters := make([]string, 0, len(match)+1)
	for _, m := range match {
		if m.Field == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	filters = append(filters, strconv.Quote(phrase))
	return strings.Join(filters, " AND "), nil
}

var (
	// httpApp is the app id cut out of a service's name.
	httpApp = strconv.Quote(HTTPUnpackPrefix + "app")
	// httpOriginDuration is how long the proxy waited on the app for a
	// request, in ns: 0 for one it found no replica for.
	httpOriginDuration = strconv.Quote(HTTPUnpackPrefix + "OriginDuration")
	// httpCapped is that time, the range's at most; httpCappedShort the range's
	// last part's at most.
	httpCapped      = strconv.Quote(HTTPUnpackPrefix + "capped")
	httpCappedShort = strconv.Quote(HTTPUnpackPrefix + "cappedShort")
)

// BuildRequestLoadQuery turns a request into LogsQL by the rules BuildQuery
// keeps: every value from the request in a Go-quoted literal, the structure
// fixed here. One query for every app asked about, by the app id its services'
// names hold - 200 at most, see loadIDsPerQuery. The lines are narrowed
// to the apps' by their services' names as phrases before they are unpacked:
// what the query reads grows with the apps' requests, not the cluster's.
//
// A request's time is OriginDuration, how long the proxy waited on the app.
// Not filtered by OriginStatus: Traefik 3.7 writes it 0 for nearly every
// request the app answered - 22 of 61,460 carried one, on a local install -
// so a filter on it counted next to nothing. And not Duration, which adds the
// proxy's own time and a slow client's. A request no replica answered took
// none of the app's time.
//
// A request counts for the range at most: a WebSocket or a long poll, logged
// when it ends with its whole duration, would otherwise read as many requests
// at once - and is not counted at all while it is open.
func BuildRequestLoadQuery(req *loggingmodel.RequestLoadReq) (string, error) {
	if len(req.AppIDs) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	span, err := loadSpan(req.Start, req.End)
	if err != nil {
		return "", err
	}
	ids := make([]string, 0, len(req.AppIDs))
	names := make([]string, 0, len(req.AppIDs))
	for _, id := range req.AppIDs {
		if id == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		ids = append(ids, regexp.QuoteMeta(strings.ToLower(id)))
		names = append(names, "svc-"+strings.ToLower(id)+"-")
	}
	short, err := shortPart(req.Start, req.ShortStart, req.End)
	if err != nil {
		return "", err
	}
	head, err := loadMatch(req.Match, accessLogPhrase)
	if err != nil {
		return "", err
	}
	// Anchored, so that an app's id never matches as the start of another's.
	services := "^svc-(" + strings.Join(ids, "|") + ")-[0-9]+@swarm$"
	q := head + ` AND ` + anyPhrase(names) +
		` | unpack_json from _msg fields (ServiceName, OriginDuration) result_prefix ` +
		strconv.Quote(HTTPUnpackPrefix) +
		` | filter ` + httpService + `:~` + strconv.Quote(services) +
		` | copy ` + httpService + ` as ` + httpApp +
		` | replace_regexp ("^svc-(.+)-[0-9]+@swarm$", "$1") at ` + httpApp +
		` | math min(` + httpOriginDuration + `, ` + strconv.FormatInt(span.Nanoseconds(), 10) + `) as ` +
		httpCapped
	if short != nil {
		q += ` | math min(` + httpOriginDuration + `, ` + strconv.FormatInt(short.span.Nanoseconds(), 10) +
			`) as ` + httpCappedShort
	}
	// min() of a duration that is not a number is the range: only numbers
	// are summed.
	q += ` | stats by (` + httpApp + `) sum(` + httpCapped + `) if (` + httpOriginDuration + `:>=0) busyNs,` +
		` count() requests`
	if short != nil {
		q += `, sum(` + httpCappedShort + `) if (` + short.filter + ` ` + httpOriginDuration +
			`:>=0) shortBusyNs, count() if (` + short.filter + `) shortRequests`
	}
	return q, nil
}

// RequestLoad says how busy apps were over the request's range, by the ids
// asked with.
func (c *Client) RequestLoad(
	ctx context.Context, req *loggingmodel.RequestLoadReq,
) (*loggingmodel.RequestLoadResp, error) {
	// The names hold the ids lower-cased: back to the ids asked with.
	byLower := make(map[string]string, len(req.AppIDs))
	for _, id := range req.AppIDs {
		byLower[strings.ToLower(id)] = id
	}
	// The access log's durations are in ns.
	ms := float64(time.Millisecond)
	out := &loggingmodel.RequestLoadResp{ByApp: make(map[string]*loggingmodel.RequestLoad, len(req.AppIDs))}
	err := inChunks(req.AppIDs, func(ids []string) error {
		part := *req
		part.AppIDs = ids
		q, err := BuildRequestLoadQuery(&part)
		if err != nil {
			return hperrors.Wrap(err)
		}
		rows, err := c.rows(ctx, q, req.Start, req.End)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, row := range rows {
			id, ok := byLower[row[HTTPUnpackPrefix+"app"]]
			if !ok {
				continue
			}
			out.ByApp[id] = &loggingmodel.RequestLoad{
				BusyMs: value(row["busyNs"]) / ms, Requests: whole(row["requests"]),
				ShortBusyMs: value(row["shortBusyNs"]) / ms, ShortRequests: whole(row["shortRequests"]),
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuildCPULoadQuery turns a request into LogsQL by the rules BuildQuery keeps.
// One query for every app asked about - 200 at most, see
// loadIDsPerQuery: each container's average and limit. The rows are narrowed
// to the apps' by their ids as phrases before they are unpacked.
func BuildCPULoadQuery(req *loggingmodel.CPULoadReq) (string, error) {
	if len(req.AppIDs) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	ids := make([]string, 0, len(req.AppIDs))
	for _, id := range req.AppIDs {
		if id == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		ids = append(ids, strconv.Quote(id))
	}
	head, err := loadMatch(req.Match, resourcePhrase)
	if err != nil {
		return "", err
	}
	return head + ` AND ` + anyPhrase(req.AppIDs) +
		` | unpack_json from _msg fields (hp, app, container, cpu, cpuLimit) result_prefix ` +
		strconv.Quote(ResourceUnpackPrefix) +
		` | filter ` + resField("hp") + `:="resources" ` + resField("app") + `:in(` + strings.Join(ids, ",") + `)` +
		` | stats by (` + resField("app") + `, ` + resField("container") + `) avg(` + resField("cpu") + `) cpu,` +
		` max(` + resField("cpuLimit") + `) cpuLimit`, nil
}

// CPULoad reads apps' containers' CPU over the request's range.
func (c *Client) CPULoad(ctx context.Context, req *loggingmodel.CPULoadReq) (*loggingmodel.CPULoadResp, error) {
	out := &loggingmodel.CPULoadResp{ByApp: make(map[string][]*loggingmodel.ContainerCPU, len(req.AppIDs))}
	err := inChunks(req.AppIDs, func(ids []string) error {
		part := *req
		part.AppIDs = ids
		q, err := BuildCPULoadQuery(&part)
		if err != nil {
			return hperrors.Wrap(err)
		}
		rows, err := c.rows(ctx, q, req.Start, req.End)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, row := range rows {
			// A container with no measure in the range - its first row has
			// none - is left out, not read as idle.
			cpu := number(row["cpu"])
			if cpu == nil {
				continue
			}
			app := row[ResourceUnpackPrefix+"app"]
			out.ByApp[app] = append(out.ByApp[app], &loggingmodel.ContainerCPU{
				Container: row[ResourceUnpackPrefix+"container"], CPU: *cpu, Limit: value(row["cpuLimit"]),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
