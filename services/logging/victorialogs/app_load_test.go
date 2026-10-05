package victorialogs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

var (
	traefikMatch = []loggingmodel.FieldMatch{{Field: traefikField, Value: "traefik"}}
	agentMatch   = []loggingmodel.FieldMatch{{Field: agentField, Value: "agent"}}
)

func TestRequestLoadQueryIsOneForAllTheApps(t *testing.T) {
	q, err := BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{
		Match: traefikMatch, AppIDs: []string{"APP1", `x.y") | delete | ("`}, Start: loadStart, End: loadEnd,
	})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(q, `"`+traefikField+`":="traefik" AND "ServiceName"`), q)
	// The lines narrowed to the apps' by their services' names, quoted, before
	// they are unpacked.
	assert.Contains(t, q, `AND "ServiceName" AND ("svc-app1-" OR "svc-x.y\") | delete | (\"-") | unpack_json`)
	// Ids lower-cased, as the services are named, and their regexp's
	// metacharacters escaped; the pattern a quoted literal.
	assert.Contains(t, q, `:~"^svc-(app1|x\\.y\"\\) \\| delete \\| \\(\")-[0-9]+@swarm$"`)
	// The app's time, OriginDuration, counts for the range at most.
	assert.Contains(t, q, `| math min("http.OriginDuration", 60000000000) as "http.capped"`)
	assert.Contains(t, q, `sum("http.capped") if ("http.OriginDuration":>=0) busyNs, count() requests`)
	assert.NotContains(t, q, "OriginStatus", "Traefik 3 writes it 0 for what the app answered")

	_, err = BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{Match: traefikMatch})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	_, err = BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{AppIDs: []string{"A"}, Start: loadStart, End: loadEnd})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	_, err = BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{Match: traefikMatch, AppIDs: []string{"A"}})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "no range")

	// The range's last part, summed apart in the same query.
	q, err = BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{Match: traefikMatch, AppIDs: []string{"A"},
		Start: loadStart, End: loadEnd, ShortStart: loadEnd.Add(-15 * time.Second)})
	assert.NoError(t, err)
	assert.Contains(t, q, `| math min("http.OriginDuration", 15000000000) as "http.cappedShort"`)
	assert.True(t, strings.HasSuffix(q, `, sum("http.cappedShort") if (_time:>=2026-10-03T10:00:45Z `+
		`"http.OriginDuration":>=0) shortBusyNs, count() if (_time:>=2026-10-03T10:00:45Z) shortRequests`), q)
	_, err = BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{Match: traefikMatch, AppIDs: []string{"A"},
		Start: loadStart, End: loadEnd, ShortStart: loadEnd})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
}

func TestCPULoadQueryIsOneForAllTheApps(t *testing.T) {
	q, err := BuildCPULoadQuery(&loggingmodel.CPULoadReq{Match: agentMatch, AppIDs: []string{"A1", `x") | delete`}})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(q, `"`+agentField+`":="agent" AND`), q)
	assert.Contains(t, q, `AND "\"hp\":\"resources\"" AND ("A1" OR "x\") | delete") | unpack_json`)
	assert.Contains(t, q, `"res.app":in("A1","x\") | delete")`)
	assert.Contains(t, q, `| stats by ("res.app", "res.container") avg("res.cpu") cpu, max("res.cpuLimit") cpuLimit`)

	_, err = BuildCPULoadQuery(&loggingmodel.CPULoadReq{Match: agentMatch})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
}

// Apps are asked about 200 a query, every query's rows in the one answer: a
// query of them all could pass VictoriaLogs' 16 KB limit, which a full one of
// ULIDs does not.
func TestLoadsAskAbout200AppsAQuery(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		q := form.Get("query")
		queries = append(queries, q)
		// A row for the first app each query asks about.
		first := regexp.MustCompile(`"svc-(app\d+)-"|"(APP\d+)"|:in\("(APP\d+)"`).FindStringSubmatch(q)
		id := first[1] + first[2] + first[3]
		_, _ = fmt.Fprintf(w, `{"http.app":%q,"requests":"3","res.app":%q,"res.container":"c","cpu":"0.5",`+
			`"attrs.hivepaas.app.id":%q,"calls":"2"}`+"\n", id, id, id)
	}))
	defer srv.Close()
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})
	ids := make([]string, 0, 450)
	for i := range 450 {
		ids = append(ids, fmt.Sprintf("APP%023d", i))
	}
	ctx := context.Background()

	requests, err := c.RequestLoad(ctx, &loggingmodel.RequestLoadReq{Match: traefikMatch, AppIDs: ids,
		Start: loadStart, End: loadEnd, ShortStart: loadEnd.Add(-15 * time.Second)})
	assert.NoError(t, err)
	cpu, err := c.CPULoad(ctx, &loggingmodel.CPULoadReq{Match: agentMatch, AppIDs: ids, Start: loadStart, End: loadEnd})
	assert.NoError(t, err)
	calls, err := c.InvocationLoad(ctx, &loggingmodel.InvocationLoadReq{Field: "attrs.hivepaas.app.id", AppIDs: ids,
		Start: loadStart, End: loadEnd, ShortStart: loadEnd.Add(-15 * time.Second)})
	assert.NoError(t, err)

	assert.Len(t, queries, 9, "3 a load")
	for _, q := range queries {
		assert.Less(t, len(q), 16384)
	}
	for _, id := range []string{ids[0], ids[200], ids[400]} {
		if assert.Contains(t, requests.ByApp, id) {
			assert.Equal(t, int64(3), requests.ByApp[id].Requests)
		}
		assert.Len(t, cpu.ByApp[id], 1)
		if assert.Contains(t, calls.ByApp, id) {
			assert.Equal(t, int64(2), calls.ByApp[id].Calls)
		}
	}
	assert.Len(t, requests.ByApp, 3)

	// None is refused, as ever.
	_, err = c.RequestLoad(ctx, &loggingmodel.RequestLoadReq{Match: traefikMatch, Start: loadStart, End: loadEnd})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
}

// ingestLines posts JSON lines to a live VictoriaLogs.
func ingestLines(t *testing.T, base string, lines []string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, base+"/insert/jsonline",
		strings.NewReader(strings.Join(lines, "\n")+"\n"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/stream+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveRequestLoadSumsEachAppsRequests(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := time.Now().UTC().Format("150405000000000")
	busy, other, late := "BUSY"+run, "OTHER"+run, "LATE"+run
	start := time.Now().UTC().Truncate(time.Minute).Add(-2 * time.Minute)
	// Traefik 3 writes OriginStatus 0 for what the app answered: the app's
	// time is OriginDuration; Duration adds the proxy's.
	line := func(component, service string, at time.Duration, originMs float64) string {
		msg := fmt.Sprintf(`{\"ServiceName\":\"%s\",\"DownstreamStatus\":200,\"OriginStatus\":0,`+
			`\"OriginDuration\":%d,\"Duration\":%d}`, service, int64(originMs*1e6), int64(originMs*1e6)+5_000_000)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, traefikField, component)
	}
	svc := func(id string, n int) string { return fmt.Sprintf("svc-%s-%d@swarm", strings.ToLower(id), n) }
	ingestLines(t, base, []string{
		line("traefik", svc(busy, 0), 10*time.Second, 400),
		line("traefik", svc(busy, 1), 20*time.Second, 600),
		// No replica answered: counted, no time of the app's.
		line("traefik", svc(busy, 0), 30*time.Second, 0),
		// A WebSocket open for an hour, closed in the minute: in flight for
		// the minute, not an hour.
		line("traefik", svc(busy, 1), 40*time.Second, 3_600_000),
		// A line whose time is not a number: counted, none of its time.
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"ServiceName\":\"%s\",\"OriginDuration\":\"x\"}",`+
			`"%s":"traefik"}`, start.Add(45*time.Second).Format(time.RFC3339Nano), svc(busy, 0), traefikField),
		// Another app whose id starts the same, one not asked about, a line an
		// app printed.
		line("traefik", svc(busy+"X", 0), 10*time.Second, 99999),
		line("traefik", svc(other, 0), 10*time.Second, 99999),
		line("app", svc(busy, 0), 10*time.Second, 99999),
		// One request in the range's last part, long - counted for that part at
		// most there - and one before it.
		line("traefik", svc(late, 0), 15*time.Second, 1000),
		line("traefik", svc(late, 1), 50*time.Second, 20_000),
	})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := &loggingmodel.RequestLoadReq{Match: traefikMatch, AppIDs: []string{busy, late}, Start: start,
		End: start.Add(time.Minute), ShortStart: start.Add(44*time.Second + 500*time.Millisecond)}
	var got *loggingmodel.RequestLoadResp
	for range 20 { // ingestion becomes visible within a second or two
		var err error
		if got, err = c.RequestLoad(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if len(got.ByApp) > 1 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	assert.Equal(t, map[string]*loggingmodel.RequestLoad{
		// The line at 45 s, its time not a number, is in the last part.
		busy: {BusyMs: 61_000, Requests: 5, ShortRequests: 1},
		late: {BusyMs: 21_000, Requests: 2, ShortBusyMs: 15_500, ShortRequests: 1},
	}, got.ByApp)
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveCPULoadAveragesEachContainer(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	app := "CPU" + time.Now().UTC().Format("150405000000000")
	start := time.Now().UTC().Truncate(time.Minute).Add(-2 * time.Minute)
	row := func(component, appID, container string, at time.Duration, cpu, limit float64) string {
		msg := fmt.Sprintf(`{\"hp\":\"resources\",\"app\":\"%s\",\"container\":\"%s\",\"cpu\":%g,\"cpuLimit\":%g}`,
			appID, container, cpu, limit)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, agentField, component)
	}
	ingestLines(t, base, []string{
		row("agent", app, "aaa", 15*time.Second, 0.2, 0.5),
		row("agent", app, "aaa", 30*time.Second, 0.4, 0.5),
		row("agent", app, "bbb", 15*time.Second, 0.1, 0),
		row("agent", app+"X", "ccc", 15*time.Second, 9, 1),
		// A container's first row has no measure yet: it is left out, not
		// read as idle.
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"resources\",\"app\":\"%s\",\"container\":\"new\",`+
			`\"cpuLimit\":1}","%s":"agent"}`, start.Add(20*time.Second).Format(time.RFC3339Nano), app, agentField),
		row("app", app, "evil", 15*time.Second, 9, 1),
	})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := &loggingmodel.CPULoadReq{Match: agentMatch, AppIDs: []string{app}, Start: start, End: start.Add(time.Minute)}
	var got *loggingmodel.CPULoadResp
	for range 20 {
		var err error
		if got, err = c.CPULoad(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if len(got.ByApp[app]) >= 2 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	containers := map[string]loggingmodel.ContainerCPU{}
	for _, c := range got.ByApp[app] {
		containers[c.Container] = *c
	}
	assert.InDelta(t, 0.3, containers["aaa"].CPU, 1e-9)
	assert.InDelta(t, 0.5, containers["aaa"].Limit, 1e-9)
	assert.InDelta(t, 0.1, containers["bbb"].CPU, 1e-9)
	assert.Zero(t, containers["bbb"].Limit)
	assert.Len(t, got.ByApp, 1)
	assert.Len(t, containers, 2)
}
