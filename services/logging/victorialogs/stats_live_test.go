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

// Run as TestLiveQueryCannotLeaveItsScope is run: against a VictoriaLogs whose
// URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveInvocationStatsCountOnlyTheAppsInvocationLines(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := time.Now().UTC().Format("150405.000000000")
	self, other := "FN-"+run, "OTHER-"+run
	field := "attrs.hivepaas.app.id"
	start := time.Now().UTC().Truncate(time.Minute).Add(-10 * time.Minute)
	invocationAt := func(app, path string, at time.Duration, outcome string, status int, ms float64) string {
		msg := fmt.Sprintf(`{\"hp\":\"invocation\",\"requestId\":\"r\",\"method\":\"GET\",\"path\":\"%s\",`+
			`\"status\":%d,\"durationMs\":%g,\"outcome\":\"%s\"}`, path, status, ms, outcome)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, field, app)
	}
	invocation := func(app string, at time.Duration, outcome string, status int, ms float64) string {
		return invocationAt(app, "/", at, outcome, status, ms)
	}
	lines := strings.Join([]string{
		invocation(self, 30*time.Second, "ok", 200, 10),
		invocation(self, 40*time.Second, "error", 500, 30),
		invocation(self, 3*time.Minute, "ok", 200, 20),
		invocation(self, 3*time.Minute+5*time.Second, "timeout", 504, 1000),
		invocationAt(self, "/.env", 3*time.Minute+10*time.Second, "ok", 404, 1),
		// The handler's own lines: one mentions an invocation, and is no call.
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"log\",\"msg\":\"hp invocation\"}","%s":%q,"stream":"stdout"}`,
			start.Add(time.Minute).Format(time.RFC3339Nano), field, self),
		// Another app's call, and another app's line naming self.
		invocation(other, time.Minute, "ok", 200, 5),
		fmt.Sprintf(`{"_time":%q,"_msg":"{\"hp\":\"invocation\",\"%s\":\"%s\",\"outcome\":\"ok\"}","%s":%q,`+
			`"stream":"stdout"}`, start.Add(time.Minute).Format(time.RFC3339Nano), field, self, field, other),
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
	req := &loggingmodel.InvocationStatsReq{
		Match: []loggingmodel.FieldMatch{{Field: field, Value: self}},
		Start: start, End: start.Add(10 * time.Minute), Step: time.Minute, TopPaths: 20,
	}
	var got *loggingmodel.InvocationStatsResp
	for range 20 { // ingestion becomes visible within a second or two
		got, err = c.InvocationStats(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if got.Totals.Calls >= 5 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	assert.Equal(t, int64(5), got.Totals.Calls)
	assert.Equal(t, int64(2), got.Totals.Failed)
	assert.Equal(t, int64(1), got.Totals.Errors4xx)
	assert.Equal(t, int64(2), got.Totals.Errors5xx)
	assert.Equal(t, map[string]int64{"ok": 3, "error": 1, "timeout": 1}, got.ByOutcome)
	// The most called path first; the other app's call is not one of them.
	if assert.Len(t, got.ByPath, 2) {
		assert.Equal(t, "/", got.ByPath[0].Path)
		assert.Equal(t, "GET", got.ByPath[0].Method)
		assert.Equal(t, int64(4), got.ByPath[0].Calls)
		assert.Equal(t, "/.env", got.ByPath[1].Path)
		assert.Equal(t, int64(1), got.ByPath[1].Errors4xx)
	}
	if assert.Len(t, got.Buckets, 2) {
		assert.Equal(t, start, got.Buckets[0].Time)
		assert.Equal(t, int64(2), got.Buckets[0].Calls)
		assert.Equal(t, start.Add(3*time.Minute), got.Buckets[1].Time)
	}
	if assert.NotNil(t, got.Totals.P99) {
		assert.Greater(t, *got.Totals.P99, 30.0)
	}
}
