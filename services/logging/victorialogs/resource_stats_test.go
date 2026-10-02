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

const agentField = "attrs.hivepaas.component"

func resourceStatsReq(appID string) *loggingmodel.ResourceStatsReq {
	return &loggingmodel.ResourceStatsReq{
		Match: []loggingmodel.FieldMatch{{Field: agentField, Value: "agent"}},
		AppID: appID,
		Start: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
		Step: 15 * time.Minute, TopContainers: 20,
	}
}

// The agent's rows only, by the identity the daemon wrote; the app's by the
// app the agent wrote into them; each container averaged, then summed.
func TestResourceStatsQueriesScopeTheAgentsRowsToTheApp(t *testing.T) {
	q, err := BuildResourceStatsQueries(resourceStatsReq("01K6A"))
	assert.NoError(t, err)
	for _, query := range []string{q.Series, q.Containers} {
		assert.True(t, strings.HasPrefix(query, `"attrs.hivepaas.component":="agent" AND "\"hp\":\"resources\""`))
		assert.Contains(t, query, `| filter "res.hp":="resources" "res.app":="01K6A"`)
	}
	assert.Contains(t, q.Series, `| stats by (_time:900s, "res.container") avg("res.cpu") cpu`)
	assert.Contains(t, q.Series, `| stats by (_time) sum(cpu) cpu`)
	assert.True(t, strings.HasSuffix(q.Containers, `| sort by (lastSeen desc, "res.container") limit 20`))

	_, err = BuildResourceStatsQueries(resourceStatsReq(""))
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)
	q, err = BuildResourceStatsQueries(resourceStatsReq(`x" | delete | "`))
	assert.NoError(t, err)
	assert.Contains(t, q.Series, `"res.app":="x\" | delete | \""`)
}

// Run against a VictoriaLogs whose URL is in HP_TEST_VICTORIALOGS_URL.
func TestLiveResourceStatsSumTheAppsContainers(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	app := "APP" + time.Now().UTC().Format("150405000000000")
	start := time.Now().UTC().Truncate(time.Minute).Add(-10 * time.Minute)
	row := func(component, appID, container string, at time.Duration, cpu float64, memory, oom int) string {
		msg := fmt.Sprintf(`{\"hp\":\"resources\",\"app\":\"%s\",\"container\":\"%s\",\"cpu\":%g,\"cpuLimit\":1,`+
			`\"memory\":%d,\"memoryLimit\":1000,\"oomKills\":%d,\"netRx\":10,\"netTx\":5,\"ioRead\":0,\"ioWrite\":0}`,
			appID, container, cpu, memory, oom)
		return fmt.Sprintf(`{"_time":%q,"_msg":"%s","%s":%q,"stream":"stdout"}`,
			start.Add(at).Format(time.RFC3339Nano), msg, agentField, component)
	}
	lines := strings.Join([]string{
		row("agent", app, "aaa", 15*time.Second, 0.2, 100, 0),
		row("agent", app, "aaa", 30*time.Second, 0.4, 300, 0),
		row("agent", app, "bbb", 15*time.Second, 0.5, 200, 1),
		row("agent", app, "aaa", 3*time.Minute, 0.1, 50, 0),
		// Another app's, and a row an app printed.
		row("agent", app+"X", "ccc", 15*time.Second, 9, 9, 0),
		row("app", app, "evil", 15*time.Second, 9, 9, 0),
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
	req := resourceStatsReq(app)
	req.Start, req.End, req.Step = start, start.Add(10*time.Minute), time.Minute
	var got *loggingmodel.ResourceStatsResp
	for range 20 { // ingestion becomes visible within a second or two
		got, err = c.ResourceStats(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Containers) >= 2 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if assert.Len(t, got.Buckets, 2) {
		first := got.Buckets[0]
		assert.Equal(t, start, first.Time)
		if assert.NotNil(t, first.CPU) && assert.NotNil(t, first.Memory) {
			assert.InDelta(t, 0.8, *first.CPU, 1e-9, "aaa's average 0.3 and bbb's 0.5")
			assert.InDelta(t, 500, *first.Memory, 1e-9, "aaa's peak 300 and bbb's 200")
		}
		assert.Equal(t, int64(1), first.OOMKills)
		assert.InDelta(t, 2, first.CPULimit, 1e-9)
	}
	if assert.Len(t, got.Containers, 2) {
		assert.Equal(t, "aaa", got.Containers[0].Container, "last seen first")
		assert.InDelta(t, 0.4, got.Containers[0].CPUPeak, 1e-9)
		assert.Equal(t, int64(1), got.Containers[1].OOMKills)
	}
}
