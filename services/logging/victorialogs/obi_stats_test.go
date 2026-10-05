package victorialogs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

var obiBuckets = []string{"le5", "le10", "le25", "leInf"}

func obiStatsReq() *loggingmodel.OBIStatsReq {
	return &loggingmodel.OBIStatsReq{Match: agentMatch, AppID: `A1") | delete | ("`, Row: "routes",
		GroupBy: []string{"method", "route"}, BucketFields: obiBuckets, Start: loadStart, End: loadEnd,
		Step: time.Minute, TopGroups: 50}
}

func TestOBIStatsQueriesAreTheAppsRowsOnly(t *testing.T) {
	q, err := BuildOBIStatsQueries(obiStatsReq())
	assert.NoError(t, err)
	for _, s := range []string{q.Series, q.Groups} {
		assert.True(t, strings.HasPrefix(s, `"`+agentField+`":="agent" AND "\"hp\":\"routes\""`), s)
		assert.Contains(t, s, `| unpack_json from _msg fields (hp, app, count, errors, sumMs, method, route, le5, le10, `+
			`le25, leInf) result_prefix "obi." | filter "obi.hp":="routes" "obi.app":="A1\") | delete | (\""`)
		assert.Contains(t, s, `sum("obi.count") count, sum("obi.errors") errors, sum("obi.sumMs") sumMs, `+
			`sum("obi.le5") "le5", sum("obi.le10") "le10", sum("obi.le25") "le25", sum("obi.leInf") "leInf"`)
	}
	assert.Contains(t, q.Series, `| stats by (_time:60s) `)
	assert.True(t, strings.HasSuffix(q.Series, `| sort by (_time)`), q.Series)
	assert.Contains(t, q.Groups, `| stats by ("obi.method", "obi.route") `)
	assert.True(t, strings.HasSuffix(q.Groups, `| sort by (count desc, "obi.method", "obi.route") limit 50`), q.Groups)
}

func TestOBIStatsQueriesSplitTheSeries(t *testing.T) {
	req := obiStatsReq()
	req.Row, req.GroupBy, req.SeriesBy = "calls", []string{"kind", "peer"}, []string{"kind"}
	q, err := BuildOBIStatsQueries(req)
	assert.NoError(t, err)
	assert.Contains(t, q.Series, `fields (hp, app, count, errors, sumMs, kind, peer, le5, `, "kind unpacked once")
	assert.Contains(t, q.Series, `| stats by (_time:60s, "obi.kind") `)
	assert.Contains(t, q.Groups, `| stats by ("obi.kind", "obi.peer") `)

	req.SeriesBy = []string{"kind) | delete"}
	_, err = BuildOBIStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
}

func TestOBIStatsQueriesRefuseWhatIsNoRowsField(t *testing.T) {
	for name, edit := range map[string]func(r *loggingmodel.OBIStatsReq){
		"a group":    func(r *loggingmodel.OBIStatsReq) { r.GroupBy = []string{`route") | delete`} },
		"a bucket":   func(r *loggingmodel.OBIStatsReq) { r.BucketFields = []string{"le 5"} },
		"no groups":  func(r *loggingmodel.OBIStatsReq) { r.GroupBy = nil },
		"no buckets": func(r *loggingmodel.OBIStatsReq) { r.BucketFields = nil },
		"no top":     func(r *loggingmodel.OBIStatsReq) { r.TopGroups = 0 },
		"a step":     func(r *loggingmodel.OBIStatsReq) { r.Step = 1500 * time.Millisecond },
	} {
		req := obiStatsReq()
		edit(req)
		_, err := BuildOBIStatsQueries(req)
		assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, name)
	}
	for name, edit := range map[string]func(r *loggingmodel.OBIStatsReq){
		"no match": func(r *loggingmodel.OBIStatsReq) { r.Match = nil },
		"no app":   func(r *loggingmodel.OBIStatsReq) { r.AppID = "" },
		"no row":   func(r *loggingmodel.OBIStatsReq) { r.Row = "" },
	} {
		req := obiStatsReq()
		edit(req)
		_, err := BuildOBIStatsQueries(req)
		assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired, name)
	}
}

func TestOBIStatusQueryIsTheLatestFirst(t *testing.T) {
	q, err := BuildOBIStatusQuery(&loggingmodel.OBIStatusReq{Match: agentMatch, Limit: 20})
	assert.NoError(t, err)
	assert.Equal(t, `"`+agentField+`":="agent" AND "\"hp\":\"obi\"" | unpack_json from _msg fields (hp) `+
		`result_prefix "obi." | filter "obi.hp":="obi" | fields _time, _msg | sort by (_time desc) limit 20`, q)

	_, err = BuildOBIStatusQuery(&loggingmodel.OBIStatusReq{Match: agentMatch})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
	_, err = BuildOBIStatusQuery(&loggingmodel.OBIStatusReq{Limit: 1})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveOBIStatsSumTheAppsRows(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	app := "OBI" + liveRun()
	start := time.Now().UTC().Truncate(time.Minute).Add(-3 * time.Minute)
	row := func(component, hp, appID, route string, at time.Duration, count, errors int, sumMs float64,
		buckets ...int) string {
		msg := fmt.Sprintf(`{\"hp\":\"%s\",\"app\":\"%s\",\"kind\":\"http\",\"method\":\"GET\",\"route\":\"%s\",`+
			`\"count\":%d,\"errors\":%d,\"sumMs\":%g,\"le5\":%d,\"le10\":%d,\"le25\":%d,\"leInf\":%d}`,
			hp, appID, route, count, errors, sumMs, buckets[0], buckets[1], buckets[2], count)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, agentField, component)
	}
	ingestLines(t, base, []string{
		// Two replicas' rows in the first minute, one in the second.
		row("agent", "routes", app, "/users", 15*time.Second, 10, 1, 80, 4, 8, 10),
		row("agent", "routes", app, "/users", 30*time.Second, 6, 0, 30, 3, 6, 6),
		row("agent", "routes", app, "/orders", 45*time.Second, 2, 2, 400, 0, 0, 1),
		row("agent", "routes", app, "/users", time.Minute+15*time.Second, 4, 0, 12, 4, 4, 4),
		// Its calls, another app's rows, and a line an app printed.
		row("agent", "calls", app, "/users", 15*time.Second, 99, 0, 1, 1, 1, 1),
		row("agent", "routes", app+"X", "/users", 15*time.Second, 99, 0, 1, 1, 1, 1),
		row("app", "routes", app, "/users", 15*time.Second, 99, 0, 1, 1, 1, 1),
	})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := &loggingmodel.OBIStatsReq{Match: agentMatch, AppID: app, Row: "routes", GroupBy: []string{"method", "route"},
		BucketFields: obiBuckets, Start: start, End: start.Add(2 * time.Minute), Step: time.Minute, TopGroups: 10}
	var got *loggingmodel.OBIStatsResp
	for range 20 { // ingestion becomes visible within a second or two
		var err error
		if got, err = c.OBIStats(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if len(got.Buckets) >= 2 && len(got.Groups) >= 2 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !assert.Len(t, got.Buckets, 2) || !assert.Len(t, got.Groups, 2) {
		return
	}
	assert.Equal(t, start, got.Buckets[0].Time.UTC())
	assert.Equal(t, loggingmodel.OBIHistogram{Count: 18, Errors: 3, SumMs: 510,
		Buckets: map[string]int64{"le5": 7, "le10": 14, "le25": 17, "leInf": 18}}, got.Buckets[0].OBIHistogram)
	assert.Equal(t, int64(4), got.Buckets[1].Count)

	assert.Equal(t, map[string]string{"method": "GET", "route": "/users"}, got.Groups[0].Keys)
	assert.Equal(t, loggingmodel.OBIHistogram{Count: 20, Errors: 1, SumMs: 122,
		Buckets: map[string]int64{"le5": 11, "le10": 18, "le25": 20, "leInf": 20}}, got.Groups[0].OBIHistogram)
	assert.Equal(t, map[string]string{"method": "GET", "route": "/orders"}, got.Groups[1].Keys)
	assert.Equal(t, int64(2), got.Groups[1].Errors)
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveOBIStatsSplitTheCallsByKind(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	app := "OBIC" + liveRun()
	start := time.Now().UTC().Truncate(time.Minute).Add(-3 * time.Minute)
	call := func(kind, peer string, at time.Duration, count int) string {
		msg := fmt.Sprintf(`{\"hp\":\"calls\",\"app\":\"%s\",\"kind\":\"%s\",\"peer\":\"%s\",\"count\":%d,`+
			`\"errors\":0,\"sumMs\":1,\"le5\":%d,\"le10\":%d,\"le25\":%d,\"leInf\":%d}`,
			app, kind, peer, count, count, count, count, count)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":"agent"}`, start.Add(at).Format(time.RFC3339Nano), msg,
			agentField)
	}
	ingestLines(t, base, []string{
		call("http", "api:8080", 10*time.Second, 3),
		call("db", "postgresql/shop", 20*time.Second, 5),
		call("db", "postgresql/shop", 30*time.Second, 1),
	})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := &loggingmodel.OBIStatsReq{Match: agentMatch, AppID: app, Row: "calls", GroupBy: []string{"kind", "peer"},
		SeriesBy: []string{"kind"}, BucketFields: obiBuckets, Start: start, End: start.Add(time.Minute),
		Step: time.Minute, TopGroups: 10}
	var got *loggingmodel.OBIStatsResp
	for range 20 {
		var err error
		if got, err = c.OBIStats(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if len(got.Buckets) >= 2 && len(got.Groups) >= 2 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	byKind := map[string]int64{}
	for _, b := range got.Buckets {
		byKind[b.Keys["kind"]] += b.Count
	}
	assert.Equal(t, map[string]int64{"http": 3, "db": 6}, byKind)
	if assert.Len(t, got.Groups, 2) {
		assert.Equal(t, map[string]string{"kind": "db", "peer": "postgresql/shop"}, got.Groups[0].Keys)
	}
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveOBIStatusIsTheAgentsStatusRows(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	node := "NODE" + liveRun()
	start := time.Now().UTC().Truncate(time.Minute).Add(-3 * time.Minute)
	status := func(component string, at time.Duration, running bool) string {
		return fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"obi\",\"node\":\"%s\",\"running\":%t}","%s":%q}`,
			start.Add(at).Format(time.RFC3339Nano), node, running, agentField, component)
	}
	ingestLines(t, base, []string{
		status("agent", 10*time.Second, false),
		status("agent", 70*time.Second, true),
		status("app", 90*time.Second, false),
	})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := &loggingmodel.OBIStatusReq{Match: agentMatch, Start: start, End: start.Add(2 * time.Minute), Limit: 10}
	var mine []*loggingmodel.OBIStatusRow
	for range 20 {
		got, err := c.OBIStatus(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		mine = mine[:0]
		for _, r := range got {
			if strings.Contains(r.Msg, node) {
				mine = append(mine, r)
			}
		}
		if len(mine) >= 2 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if assert.Len(t, mine, 2) {
		assert.Equal(t, start.Add(70*time.Second), mine[0].Time.UTC(), "the latest first")
		assert.Contains(t, mine[0].Msg, `"running":true`)
	}
}
