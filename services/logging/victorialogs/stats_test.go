package victorialogs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

func statsReq(step time.Duration) *loggingmodel.InvocationStatsReq {
	return &loggingmodel.InvocationStatsReq{
		Match: []loggingmodel.FieldMatch{{Field: appField, Value: "FN1"}},
		Start: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
		Step:  step,
	}
}

const statsHead = `"attrs.hivepaas.app.id":="FN1" AND "\"hp\":\"invocation\""` +
	` | unpack_json from _msg fields (hp, outcome, status, durationMs) result_prefix "app."` +
	` | filter "app.hp":="invocation"`

const statsCounts = ` count() calls, count() if ("app.outcome":!="ok") failed,` +
	` count() if ("app.status":>=500) errors5xx, quantile(0.5, "app.durationMs") p50,` +
	` quantile(0.95, "app.durationMs") p95, quantile(0.99, "app.durationMs") p99`

func TestInvocationStatsQueriesCountTheAppsInvocationLines(t *testing.T) {
	q, err := BuildInvocationStatsQueries(statsReq(15 * time.Minute))

	assert.NoError(t, err)
	assert.Equal(t, statsHead+" | stats by (_time:900s)"+statsCounts+" | sort by (_time)", q.Series)
	assert.Equal(t, statsHead+" | stats"+statsCounts, q.Totals)
	assert.Equal(t, statsHead+` | stats by ("app.outcome") count() calls`, q.Outcomes)
}

func TestInvocationStatsQueriesRefuseNoScopeAndNoStep(t *testing.T) {
	req := statsReq(time.Minute)
	req.Match = nil
	_, err := BuildInvocationStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)

	req = statsReq(0)
	_, err = BuildInvocationStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid)

	req = statsReq(1500 * time.Millisecond)
	_, err = BuildInvocationStatsQueries(req)
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "a step is whole seconds")
}

func TestInvocationStatsQueriesQuoteTheScope(t *testing.T) {
	req := statsReq(time.Minute)
	req.Match[0].Value = `FN1" OR *`
	q, err := BuildInvocationStatsQueries(req)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(q.Totals, `"attrs.hivepaas.app.id":="FN1\" OR *" AND `), q.Totals)
}

// A backend answering the three queries as VictoriaLogs does: one JSON object
// per line, every value a string, buckets in no order.
func TestInvocationStatsReadsTheThreeAnswers(t *testing.T) {
	var forms []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		forms = append(forms, form)
		q := form.Get("query")
		switch {
		case strings.Contains(q, "by (_time:"):
			_, _ = io.WriteString(w, `{"_time":"2026-10-02T00:15:00Z","calls":"3","failed":"1","errors5xx":"1",`+
				`"p50":"20.5","p95":"40","p99":"41"}`+"\n"+
				`{"_time":"2026-10-02T00:00:00Z","calls":"2","failed":"0","errors5xx":"0",`+
				`"p50":"10","p95":"12","p99":"12.5"}`+"\n")
		case strings.Contains(q, `by ("app.outcome")`):
			_, _ = io.WriteString(w, `{"app.outcome":"ok","calls":"4"}`+"\n"+`{"app.outcome":"error","calls":"1"}`+"\n")
		default:
			_, _ = io.WriteString(w, `{"calls":"5","failed":"1","errors5xx":"1","p50":"15","p95":"40","p99":"41"}`+"\n")
		}
	}))
	defer srv.Close()
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})

	got, err := c.InvocationStats(context.Background(), statsReq(15*time.Minute))

	assert.NoError(t, err)
	f := func(v float64) *float64 { return &v }
	assert.Equal(t, &loggingmodel.InvocationStatsResp{
		Buckets: []*loggingmodel.InvocationBucket{
			{Time: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), InvocationCounts: loggingmodel.InvocationCounts{
				Calls: 2, P50: f(10), P95: f(12), P99: f(12.5)}},
			{Time: time.Date(2026, 10, 2, 0, 15, 0, 0, time.UTC), InvocationCounts: loggingmodel.InvocationCounts{
				Calls: 3, Failed: 1, Errors5xx: 1, P50: f(20.5), P95: f(40), P99: f(41)}},
		},
		Totals:    loggingmodel.InvocationCounts{Calls: 5, Failed: 1, Errors5xx: 1, P50: f(15), P95: f(40), P99: f(41)},
		ByOutcome: map[string]int64{"ok": 4, "error": 1},
	}, got)
	if assert.Len(t, forms, 3) {
		for _, form := range forms {
			assert.Equal(t, "2026-10-02T00:00:00Z", form.Get("start"))
			assert.Equal(t, "2026-10-03T00:00:00Z", form.Get("end"))
		}
	}
}

// No call in the range: VictoriaLogs answers the totals with zero counts and
// quantiles that are no number.
func TestInvocationStatsWithoutACallHasNoDurations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "by+%28_time") || strings.Contains(string(body), "outcome") {
			return
		}
		_, _ = io.WriteString(w, `{"calls":"0","failed":"0","errors5xx":"0","p50":"NaN","p95":"NaN","p99":""}`+"\n")
	}))
	defer srv.Close()
	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: srv.URL}})

	got, err := c.InvocationStats(context.Background(), statsReq(time.Hour))

	assert.NoError(t, err)
	assert.Empty(t, got.Buckets)
	assert.Equal(t, loggingmodel.InvocationCounts{}, got.Totals)
	assert.Empty(t, got.ByOutcome)
}
