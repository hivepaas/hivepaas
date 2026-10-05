package victorialogs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const traefikField = "attrs.hivepaas.component"

func httpStatsReq(pattern string) *loggingmodel.HTTPStatsReq {
	return &loggingmodel.HTTPStatsReq{
		Match:          []loggingmodel.FieldMatch{{Field: traefikField, Value: "traefik"}},
		ServicePattern: pattern,
		Start:          time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
		End:            time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
		Step:           15 * time.Minute,
		TopPaths:       20,
		TopReplicas:    10,
	}
}

// The proxy's lines only, by the identity the daemon wrote; the app's by its
// service, matched as a quoted regular expression.
func TestHTTPStatsQueriesScopeTheProxysLinesToTheApp(t *testing.T) {
	q, err := BuildHTTPStatsQueries(httpStatsReq(`^svc-01k6a-[0-9]+@swarm$`))
	assert.NoError(t, err)
	for _, query := range []string{q.Series, q.Totals, q.Paths, q.Replicas} {
		assert.True(t, strings.HasPrefix(query, `"attrs.hivepaas.component":="traefik" AND "ServiceName"`), query)
		assert.Contains(t, query, `| filter "http.ServiceName":~"^svc-01k6a-[0-9]+@swarm$"`)
		assert.Contains(t, query, `| math "http.Duration" / 1000000 as "http.ms"`)
	}
	assert.Contains(t, q.Series, "| stats by (_time:900s) count() requests")
	assert.Contains(t, q.Paths, `replace_regexp ("/[0-9]+(/|$)", "/:n$1") at "http.RequestPath"`)
	assert.True(t, strings.HasSuffix(q.Paths,
		`| sort by (requests desc, "http.RequestPath", "http.RequestMethod") limit 20`))
	assert.True(t, strings.HasSuffix(q.Replicas, `| sort by (requests desc, "http.ServiceURL") limit 10`))
}

func TestHTTPStatsQueriesRefuseNoScope(t *testing.T) {
	req := httpStatsReq("")
	_, err := BuildHTTPStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired, "no service pattern")

	req = httpStatsReq(`^svc-x-[0-9]+@swarm$`)
	req.Match = nil
	_, err = BuildHTTPStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)

	req = httpStatsReq(`^svc-x-[0-9]+@swarm$`)
	req.Step = 1500 * time.Millisecond
	_, err = BuildHTTPStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)
}

// A pattern is a quoted literal: it cannot end the filter and start a pipe.
func TestHTTPStatsQueriesQuoteThePattern(t *testing.T) {
	q, err := BuildHTTPStatsQueries(httpStatsReq(`x" | delete | "`))
	assert.NoError(t, err)
	assert.Contains(t, q.Totals, `:~"x\" | delete | \""`)
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveHTTPStatsCountOnlyTheAppsRequests(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := liveRun()
	self := "svc-" + run + "-0@swarm"
	start := time.Now().UTC().Truncate(time.Minute).Add(-10 * time.Minute)
	line := func(component, service, replica, path string, at time.Duration, status, origin int, ms float64) string {
		return traefikAccessLine(t, component, start.Add(at), map[string]any{
			"ServiceName": service, "ServiceURL": replica, "RequestMethod": "GET", "RequestPath": path,
			"DownstreamStatus": status, "OriginStatus": origin, "Duration": int64(ms * 1e6),
		})
	}
	lines := strings.Join([]string{
		line("traefik", self, "http://10.0.0.5:8080", "/users/12", 30*time.Second, 200, 200, 10),
		line("traefik", self, "http://10.0.0.5:8080", "/users/34", 40*time.Second, 500, 500, 30),
		line("traefik", self, "http://10.0.0.6:8080", "/", 3*time.Minute, 404, 404, 2),
		line("traefik", self, "", "/", 3*time.Minute+5*time.Second, 502, 0, 1),
		// Another app whose id starts the same, and a line an app printed.
		line("traefik", "svc-"+run+"1-0@swarm", "http://10.0.0.9:80", "/", time.Minute, 200, 200, 5),
		line("app", self, "http://10.0.0.5:8080", "/", time.Minute, 500, 500, 5),
	}, "\n") + "\n"
	ingest, err := http.NewRequest(http.MethodPost, base+"/insert/jsonline", strings.NewReader(lines))
	if err != nil {
		t.Fatal(err)
	}
	ingest.Header.Set("Content-Type", "application/stream+json")
	resp, err := http.DefaultClient.Do(ingest)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := httpStatsReq("^svc-" + run + "-[0-9]+@swarm$")
	req.Start, req.End, req.Step = start, start.Add(10*time.Minute), time.Minute
	var got *loggingmodel.HTTPStatsResp
	for range 20 { // ingestion becomes visible within a second or two
		got, err = c.HTTPStats(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		// The series' rows too, read first: the lines may become visible
		// between the queries.
		if len(got.Buckets) >= 2 && got.Totals.Requests >= 4 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	assert.Equal(t, int64(4), got.Totals.Requests)
	assert.Equal(t, int64(1), got.Totals.Errors4xx)
	assert.Equal(t, int64(2), got.Totals.Errors5xx)
	assert.Equal(t, int64(1), got.Totals.Unreachable)
	if assert.Len(t, got.ByPath, 2) {
		assert.Equal(t, "/", got.ByPath[0].Path)
		assert.Equal(t, "/users/:n", got.ByPath[1].Path)
		assert.Equal(t, int64(2), got.ByPath[1].Requests)
	}
	if assert.Len(t, got.ByReplica, 3) {
		assert.Equal(t, "http://10.0.0.5:8080", got.ByReplica[0].Address)
		assert.Equal(t, int64(2), got.ByReplica[0].Requests)
	}
	if assert.Len(t, got.Buckets, 2) {
		assert.Equal(t, start, got.Buckets[0].Time)
		assert.Equal(t, int64(2), got.Buckets[0].Requests)
	}
	if assert.NotNil(t, got.Totals.P99) {
		assert.Greater(t, *got.Totals.P99, 10.0)
	}
}

// Given the phrase every one of the app's lines holds, the lines are narrowed
// by it before they are unpacked; without one, as before.
func TestHTTPStatsQueriesNarrowByTheServicePhrase(t *testing.T) {
	req := httpStatsReq("^svc-01k6a-[0-9]+@swarm$")
	q, err := BuildHTTPStatsQueries(req)
	assert.NoError(t, err)
	assert.Contains(t, q.Totals, `AND "ServiceName" | extract`)

	req.ServicePhrase = "svc-01k6a-"
	q, err = BuildHTTPStatsQueries(req)
	assert.NoError(t, err)
	for _, query := range []string{q.Series, q.Totals, q.Paths, q.Replicas} {
		assert.Contains(t, query, `AND "ServiceName" AND "svc-01k6a-" | extract`)
	}
}

// The four queries run at once: each waits here until all four have come.
// The first to fail is the error, and the others are not waited for.
func TestHTTPStatsRunsItsQueriesAtOnce(t *testing.T) {
	var arrived sync.WaitGroup
	arrived.Add(4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arrived.Done()
		done := make(chan struct{})
		go func() { arrived.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		_, _ = io.WriteString(w, `{"_time":"2026-10-02T00:00:00Z","requests":"1"}`+"\n")
	}))
	defer srv.Close()
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})

	got, err := c.HTTPStats(context.Background(), httpStatsReq("^svc-01k6a-[0-9]+@swarm$"))

	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, int64(1), got.Totals.Requests)
		assert.Len(t, got.Buckets, 1)
	}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "limit+10") || strings.Contains(string(body), "limit%2010") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer failing.Close()
	c = New(&Config{Endpoint: loggingmodel.Endpoint{URL: failing.URL}})
	started := time.Now()

	_, err = c.HTTPStats(context.Background(), httpStatsReq("^svc-01k6a-[0-9]+@swarm$"))

	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "the replicas' query's own error")
	assert.Less(t, time.Since(started), time.Second, "the others were not waited for")
}

// An access log line's fields are taken each by the text before its value in
// Traefik's JSON - its own key, quoted, a colon - up to the comma after it.
func TestAccessLogFieldsAreExtractedByTheirKeys(t *testing.T) {
	assert.Equal(t, ` | extract "\"ServiceName\":<http.ServiceName>," from _msg`+
		` | extract "\"Duration\":<http.Duration>," from _msg`, extractAccessLog("ServiceName", "Duration"))

	q, err := BuildHTTPStatsQueries(httpStatsReq("^svc-01k6a-[0-9]+@swarm$"))
	assert.NoError(t, err)
	assert.NotContains(t, q.Totals, "unpack_json")
	// Each query takes the fields it counts by, and only after the service
	// filter.
	assert.Contains(t, q.Totals, `@swarm$" | extract "\"DownstreamStatus\":`)
	assert.NotContains(t, q.Totals, "RequestPath")
	assert.Contains(t, q.Paths, `extract "\"RequestPath\":<http.RequestPath>," from _msg | replace_regexp`)
	assert.Contains(t, q.Replicas, `extract "\"ServiceURL\":<http.ServiceURL>," from _msg | stats by ("http.ServiceURL")`)
}

// traefikOwnLine is an access log line Traefik 3.7.13 wrote, verbatim but its
// service, which %s is: the fields are read from the layout it writes, not
// from one a test made up.
const traefikOwnLine = `{"ClientAddr":"172.19.0.1:62090","ClientHost":"172.19.0.1","ClientPort":"62090",` +
	`"ClientUsername":"-","DownstreamContentSize":365,"DownstreamStatus":200,"Duration":2196000,` +
	`"OriginContentSize":365,"OriginDuration":2079000,"OriginStatus":200,"Overhead":117000,` +
	`"RequestAddr":"127.0.0.1:18081","RequestContentSize":0,"RequestCount":1,"RequestHost":"127.0.0.1",` +
	`"RequestMethod":"GET","RequestPath":"/ok/a,b","RequestPort":"18081","RequestProtocol":"HTTP/1.1",` +
	`"RequestScheme":"http","RetryAttempts":0,"RouterName":"ok@file","ServiceAddr":"hp-tr-who:80",` +
	`"ServiceName":"%s","ServiceURL":"http://hp-tr-who:80","StartLocal":"2026-10-05T07:33:48.823112007Z",` +
	`"StartUTC":"2026-10-05T07:33:48.823112007Z","entryPointName":"web","level":"info","msg":"",` +
	`"time":"2026-10-05T07:33:48Z"}`

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveHTTPStatsReadTraefiksOwnLine(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := liveRun()
	at := time.Now().UTC().Add(-time.Minute)
	msg := strings.Replace(traefikOwnLine, "%s", "svc-"+run+"-0@swarm", 1)
	line, err := json.Marshal(map[string]string{
		"_time": at.Format(time.RFC3339Nano), "_msg": msg, traefikField: "traefik", "stream": "stdout",
	})
	if err != nil {
		t.Fatal(err)
	}
	ingestLines(t, base, []string{string(line)})

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	req := httpStatsReq("^svc-" + run + "-[0-9]+@swarm$")
	req.Start, req.End, req.Step = at.Add(-time.Minute), at.Add(time.Minute), time.Minute
	var got *loggingmodel.HTTPStatsResp
	for range 20 { // ingestion becomes visible within a second or two
		if got, err = c.HTTPStats(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if len(got.Buckets) >= 1 && got.Totals.Requests >= 1 && len(got.ByPath) >= 1 && len(got.ByReplica) >= 1 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	assert.Equal(t, int64(1), got.Totals.Requests)
	assert.Zero(t, got.Totals.Errors5xx)
	if assert.NotNil(t, got.Totals.P50) {
		assert.InDelta(t, 2.196, *got.Totals.P50, 0.01, "Duration, in ms")
	}
	if assert.Len(t, got.ByPath, 1) {
		assert.Equal(t, "GET", got.ByPath[0].Method)
		assert.Equal(t, "/ok/a,b", got.ByPath[0].Path, "a quoted value with a comma, whole")
	}
	if assert.Len(t, got.ByReplica, 1) {
		assert.Equal(t, "http://hp-tr-who:80", got.ByReplica[0].Address)
	}
}
