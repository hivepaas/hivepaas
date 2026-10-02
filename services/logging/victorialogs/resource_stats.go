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

// ResourceUnpackPrefix is where a resource row's fields are unpacked to.
const ResourceUnpackPrefix = "res."

// resourcePhrase narrows the agent's lines to its resource rows; the filter
// after the unpacking is what decides.
const resourcePhrase = `"hp":"resources"`

func resField(name string) string { return strconv.Quote(ResourceUnpackPrefix + name) }

// ResourceStatsQueries are the two queries an app's usage takes: by step, and
// by container.
type ResourceStatsQueries struct {
	Series     string
	Containers string
}

// BuildResourceStatsQueries turns a request into LogsQL by the rules
// BuildQuery keeps: every value from the request in a Go-quoted literal, the
// structure fixed here, the unpacked fields under ResourceUnpackPrefix.
//
// A step's usage is each container's average over it - its memory's peak -
// then summed over the containers: two replicas at half a core are a core.
func BuildResourceStatsQueries(req *loggingmodel.ResourceStatsReq) (*ResourceStatsQueries, error) {
	if len(req.Match) == 0 || req.AppID == "" {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Step < time.Second || req.Step%time.Second != 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the step must be whole seconds")
	}
	if req.TopContainers <= 0 {
		return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("the containers must be a number of them")
	}
	filters := make([]string, 0, len(req.Match)+1)
	for _, m := range req.Match {
		if m.Field == "" {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	filters = append(filters, strconv.Quote(resourcePhrase))
	head := strings.Join(filters, " AND ") +
		` | unpack_json from _msg fields (hp, app, container, cpu, cpuLimit, memory, memoryLimit, oomKills,` +
		` netRx, netTx, ioRead, ioWrite) result_prefix ` + strconv.Quote(ResourceUnpackPrefix) +
		` | filter ` + resField("hp") + `:="resources" ` + resField("app") + `:=` + strconv.Quote(req.AppID)

	perContainer := fmt.Sprintf(` | stats by (_time:%ds, %s) avg(%s) cpu, max(%s) cpuLimit, max(%s) memory,`+
		` max(%s) memoryLimit, sum(%s) oomKills, avg(%s) netRx, avg(%s) netTx, avg(%s) ioRead, avg(%s) ioWrite`,
		req.Step/time.Second, resField("container"), resField("cpu"), resField("cpuLimit"), resField("memory"),
		resField("memoryLimit"), resField("oomKills"), resField("netRx"), resField("netTx"), resField("ioRead"),
		resField("ioWrite"))
	summed := ` | stats by (_time) sum(cpu) cpu, sum(cpuLimit) cpuLimit, sum(memory) memory,` +
		` sum(memoryLimit) memoryLimit, sum(oomKills) oomKills, sum(netRx) netRx, sum(netTx) netTx,` +
		` sum(ioRead) ioRead, sum(ioWrite) ioWrite | sort by (_time)`

	return &ResourceStatsQueries{
		Series: head + perContainer + summed,
		// The containers last seen first: the running ones, then those replaced.
		Containers: fmt.Sprintf(`%s | stats by (%s) avg(%s) cpu, max(%s) cpuPeak, max(%s) memory,`+
			` max(%s) memoryLimit, sum(%s) oomKills, max(_time) lastSeen | sort by (lastSeen desc, %s) limit %d`,
			head, resField("container"), resField("cpu"), resField("cpu"), resField("memory"),
			resField("memoryLimit"), resField("oomKills"), resField("container"), req.TopContainers),
	}, nil
}

// ResourceStats reads an app's usage from the agent's rows: two queries, built
// by BuildResourceStatsQueries, over the request's range.
func (c *Client) ResourceStats(
	ctx context.Context, req *loggingmodel.ResourceStatsReq,
) (*loggingmodel.ResourceStatsResp, error) {
	q, err := BuildResourceStatsQueries(req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingmodel.ResourceStatsResp{}

	rows, err := c.rows(ctx, q.Series, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		at, err := time.Parse(time.RFC3339Nano, row["_time"])
		if err != nil {
			return nil, hperrors.Wrap(loggingmodel.ErrQueryInvalid).WithExtraDetail("a bucket's time: %s", row["_time"])
		}
		out.Buckets = append(out.Buckets, &loggingmodel.ResourceBucket{Time: at, ResourceUsage: usageOf(row)})
	}
	sort.Slice(out.Buckets, func(i, j int) bool { return out.Buckets[i].Time.Before(out.Buckets[j].Time) })

	rows, err = c.rows(ctx, q.Containers, req.Start, req.End)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, row := range rows {
		lastSeen, _ := time.Parse(time.RFC3339Nano, row["lastSeen"])
		out.Containers = append(out.Containers, &loggingmodel.ResourceContainer{
			Container: row[ResourceUnpackPrefix+"container"],
			CPU:       value(row["cpu"]), CPUPeak: value(row["cpuPeak"]),
			Memory: value(row["memory"]), MemoryLimit: value(row["memoryLimit"]),
			OOMKills: whole(row["oomKills"]), LastSeen: lastSeen,
		})
	}
	return out, nil
}

func usageOf(row map[string]string) loggingmodel.ResourceUsage {
	return loggingmodel.ResourceUsage{
		CPU: number(row["cpu"]), CPULimit: value(row["cpuLimit"]),
		Memory: number(row["memory"]), MemoryLimit: value(row["memoryLimit"]),
		OOMKills: int64(value(row["oomKills"])),
		NetRx:    value(row["netRx"]), NetTx: value(row["netTx"]),
		IORead: value(row["ioRead"]), IOWrite: value(row["ioWrite"]),
	}
}

// value is a number, 0 for what is none.
func value(s string) float64 {
	if n := number(s); n != nil {
		return *n
	}
	return 0
}
