package dockerproxy

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func eventsPath(filters string) string {
	return "/v1.51/events?filters=" + url.QueryEscape(filters)
}

// Appwrite's orchestrator watches a build's sidecar and worker, filtered the way
// the Go SDK writes filters.
func TestEventsPassForTheAppsChildren(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodGet,
		eventsPath(`{"container":{"child1":true},"type":{"container":true}}`), nil)
	stop(t, assert.Equal(t, http.StatusOK, status, string(raw)))
	assert.True(t, w.reached(http.MethodGet, "/events"))

	status, raw = w.do(t, http.MethodGet,
		eventsPath(`{"container":["child1"],"type":["container"],"event":["die"]}`), nil)
	assert.Equal(t, http.StatusOK, status, string(raw))
}

func TestEventsSayNothingOfWhatIsNotTheApps(t *testing.T) {
	w := newWorld(t, testPolicy())
	for filters, want := range map[string]string{
		``:                         "events must be filtered to containers of the app",
		`{"type":["container"]}`:   "events must be filtered to containers of the app",
		`{"container":["child1"]}`: "events must be filtered to type container",
		`{"container":["child1"],"type":["container","network"]}`:       "events must be filtered to type container",
		`{"container":["child1","other1"],"type":["container"]}`:        "container other1 is not one this app started",
		`{"container":["task1"],"type":["container"]}`:                  "container task1 is not one this app started",
		`{"container":["child1"],"type":["container"],"label":["a=b"]}`: "events filter label is not allowed",
		`not json`: "events filters are not readable",
	} {
		status, raw := w.do(t, http.MethodGet, eventsPath(filters), nil)
		assert.Equal(t, http.StatusForbidden, status, filters)
		assert.Equal(t, "hivepaas: "+want, refusalMessage(t, raw), filters)
	}
	assert.False(t, w.reached(http.MethodGet, "/events"))
}
