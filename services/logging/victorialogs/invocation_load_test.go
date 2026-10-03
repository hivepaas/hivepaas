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

// loadStart and loadEnd are a minute's range for the load queries.
var (
	loadStart = time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	loadEnd   = loadStart.Add(time.Minute)
)

func TestInvocationLoadQueryIsOneForAllTheFunctions(t *testing.T) {
	q, err := BuildInvocationLoadQuery(&loggingmodel.InvocationLoadReq{
		Field: appField, AppIDs: []string{"FN1", `x") | delete | ("`}, Start: loadStart, End: loadEnd,
	})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(q,
		`"attrs.hivepaas.app.id":in("FN1","x\") | delete | (\"") AND "\"hp\":\"invocation\""`), q)
	// A call counts for the range at most.
	assert.Contains(t, q, `| math min("app.durationMs", 60000) as "app.cappedMs"`)
	assert.Contains(t, q,
		`| stats by ("attrs.hivepaas.app.id") sum("app.cappedMs") if ("app.durationMs":>=0) busyMs, count() calls,`)
	assert.Contains(t, q, `count() if ("app.outcome":="throttled") throttled`)

	_, err = BuildInvocationLoadQuery(&loggingmodel.InvocationLoadReq{Field: appField})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	_, err = BuildInvocationLoadQuery(&loggingmodel.InvocationLoadReq{Field: appField, AppIDs: []string{""}})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	_, err = BuildInvocationLoadQuery(&loggingmodel.InvocationLoadReq{Field: appField, AppIDs: []string{"FN1"}})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "no range")
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveInvocationLoadSumsEachFunctionsCalls(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := time.Now().UTC().Format("150405.000000000")
	busy, throttled, other, long := "BUSY-"+run, "THR-"+run, "OTHER-"+run, "LONG-"+run
	start := time.Now().UTC().Truncate(time.Minute).Add(-2 * time.Minute)
	call := func(app string, at time.Duration, ms float64, outcome string) string {
		msg := fmt.Sprintf(`{\"hp\":\"invocation\",\"durationMs\":%g,\"outcome\":\"%s\"}`, ms, outcome)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, appField, app)
	}
	lines := strings.Join([]string{
		call(busy, 10*time.Second, 400, "ok"), call(busy, 20*time.Second, 600, "error"),
		call(throttled, 30*time.Second, 50, "ok"), call(throttled, 31*time.Second, 0, "throttled"),
		call(other, 10*time.Second, 99999, "ok"),
		// Two minutes, ending in the minute: in flight for the minute, not two.
		call(long, 40*time.Second, 120_000, "ok"),
		// A call whose duration is not a number: counted, none of its time.
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"invocation\",\"durationMs\":\"x\",\"outcome\":\"ok\"}",`+
			`"%s":%q}`, start.Add(45*time.Second).Format(time.RFC3339Nano), appField, long),
		// The function's own line mentioning an invocation is no call.
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"log\",\"msg\":\"hp invocation\"}","%s":%q}`,
			start.Add(time.Second).Format(time.RFC3339Nano), appField, busy),
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
	req := &loggingmodel.InvocationLoadReq{Field: appField, AppIDs: []string{busy, throttled, long},
		Start: start, End: start.Add(time.Minute)}
	var got *loggingmodel.InvocationLoadResp
	for range 20 { // ingestion becomes visible within a second or two
		got, err = c.InvocationLoad(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.ByApp) >= 3 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if assert.Contains(t, got.ByApp, busy) && assert.Contains(t, got.ByApp, throttled) {
		assert.Equal(t, &loggingmodel.InvocationLoad{BusyMs: 1000, Calls: 2}, got.ByApp[busy])
		assert.Equal(t, &loggingmodel.InvocationLoad{BusyMs: 50, Calls: 2, Throttled: 1}, got.ByApp[throttled])
	}
	assert.Equal(t, &loggingmodel.InvocationLoad{BusyMs: 60_000, Calls: 2}, got.ByApp[long])
	assert.NotContains(t, got.ByApp, other)
}
