package victorialogs

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// Run with:
//
//	docker run -d --rm --name vl-test -p 19428:9428 victoriametrics/victoria-logs:v1.52.0
//	HP_TEST_VICTORIALOGS_URL=http://localhost:19428 go test ./services/logging/victorialogs/ -run Live
//	docker rm -f vl-test
func TestLiveQueryCannotLeaveItsScope(t *testing.T) {
	base := os.Getenv("HP_TEST_VICTORIALOGS_URL")
	if base == "" {
		t.Skip("HP_TEST_VICTORIALOGS_URL not set")
	}
	run := time.Now().UTC().Format("150405.000000000")
	self, other := "APP-"+run, "OTHER-"+run
	field := "attrs.hivepaas.app.id"

	lines := strings.Join([]string{
		// A line from self whose JSON forges another app's identity, in both
		// spellings the attack could use.
		`{"_msg":"{\"level\":\"ERROR\",\"attrs.hivepaas.app.id\":\"` + other + `\",\"hivepaas.app.id\":\"` + other +
			`\"}","` + field + `":"` + self + `","stream":"stderr"}`,
		`{"_msg":"plain line from self","` + field + `":"` + self + `","stream":"stdout"}`,
		`{"_msg":"secret of other","` + field + `":"` + other + `","stream":"stdout"}`,
	}, "\n") + "\n"
	ingest, err := http.NewRequest(http.MethodPost, base+"/insert/jsonline", strings.NewReader(lines))
	if err != nil {
		t.Fatal(err)
	}
	// Without this content type the server answers 200 and stores nothing.
	ingest.Header.Set("Content-Type", "application/stream+json")
	resp, err := http.DefaultClient.Do(ingest)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	c := New(&Config{Endpoint: loggingmodel.Endpoint{URL: base}})
	scope := []loggingmodel.FieldMatch{{Field: field, Value: self}}
	query := func(mod func(r *loggingmodel.QueryReq)) []loggingmodel.LogEntry {
		r := &loggingmodel.QueryReq{Match: scope, Limit: 100}
		if mod != nil {
			mod(r)
		}
		var got *loggingmodel.QueryResp
		for range 20 { // ingestion becomes visible within a second or two
			got, err = c.Query(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Entries) > 0 || mod != nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		return got.Entries
	}

	all := query(nil)
	assert.Len(t, all, 2, "both of self's lines, the forged one included")
	for _, e := range all {
		assert.NotContains(t, e.Message, "secret of other")
	}

	for _, hostile := range []string{
		`x" OR "` + field + `":="` + other,
		`") OR ("` + field + `":="` + other,
		"secret of other",
	} {
		got := query(func(r *loggingmodel.QueryReq) { r.Contains = hostile })
		for _, e := range got {
			assert.NotContains(t, e.Message, "secret of other", hostile)
		}
	}

	errs := query(func(r *loggingmodel.QueryReq) { r.Levels = []string{"error"} })
	if assert.Len(t, errs, 1) {
		assert.Equal(t, "ERROR", errs[0].Level, "level comes from the message, matched case-insensitively")
	}

	// The forged field inside self's message must not re-scope the query to
	// other: scoping to other returns only other's own line.
	c2 := &loggingmodel.QueryReq{Match: []loggingmodel.FieldMatch{{Field: field, Value: other}}, Limit: 100}
	otherLines, err := c.Query(context.Background(), c2)
	assert.NoError(t, err)
	if assert.Len(t, otherLines.Entries, 1) {
		assert.Equal(t, "secret of other", otherLines.Entries[0].Message)
	}
}
