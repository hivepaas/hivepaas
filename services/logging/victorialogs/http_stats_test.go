package victorialogs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
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
		msg := fmt.Sprintf(`{\"ServiceName\":\"%s\",\"ServiceURL\":\"%s\",\"RequestMethod\":\"GET\",`+
			`\"RequestPath\":\"%s\",\"DownstreamStatus\":%d,\"OriginStatus\":%d,\"Duration\":%d}`,
			service, replica, path, status, origin, int64(ms*1e6))
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, traefikField, component)
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
