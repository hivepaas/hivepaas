package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

// No id, no services: Docker would read an empty id as a prefix of every
// service, so it is not asked at all.
func TestServicesOfNoIDsAreNone(t *testing.T) {
	m := &manager{} // no client: reaching Docker would panic
	for _, ids := range [][]string{nil, {}, {"", "  "}} {
		resp, err := m.ServiceListByIDs(context.Background(), ids)
		assert.NoError(t, err)
		if assert.NotNil(t, resp) {
			assert.Empty(t, resp.Items)
		}
	}
}

// The ids go to Docker in one list, as filters it ORs; what it answers by
// prefix only is dropped, and the caller's options still apply.
func TestServicesByIDsAreAskedInOneListAndMatchedExactly(t *testing.T) {
	var asked struct {
		ids    []string
		status string
		calls  int
	}
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/services") {
			http.NotFound(w, r)
			return
		}
		asked.calls++
		var filters map[string]map[string]bool
		_ = json.Unmarshal([]byte(r.URL.Query().Get("filters")), &filters)
		for id := range filters["id"] {
			asked.ids = append(asked.ids, id)
		}
		asked.status = r.URL.Query().Get("status")
		// "svc1" is a prefix of "svc10": Docker answers both.
		_ = json.NewEncoder(w).Encode([]swarm.Service{{ID: "svc1"}, {ID: "svc10"}, {ID: "svc2"}})
	}))
	defer daemon.Close()
	c, err := client.New(client.WithHost("tcp://"+daemon.Listener.Addr().String()),
		client.WithAPIVersion(minAPIVersion))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	m := &manager{client: c}

	resp, err := m.ServiceListByIDs(context.Background(), []string{"svc1", "svc2", "svc1", ""},
		func(opts *client.ServiceListOptions) { opts.Status = true })
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, 1, asked.calls)
	assert.ElementsMatch(t, []string{"svc1", "svc2"}, asked.ids)
	assert.Equal(t, "true", asked.status)
	ids := make([]string, 0, len(resp.Items))
	for i := range resp.Items {
		ids = append(ids, resp.Items[i].ID)
	}
	assert.ElementsMatch(t, []string{"svc1", "svc2"}, ids)
}
