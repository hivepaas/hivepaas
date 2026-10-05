package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

// No id, nothing listed: Docker would read an empty id as a prefix of every
// object, so it is not asked at all.
func TestNoIDsListNothing(t *testing.T) {
	m := &manager{} // no client: reaching Docker would panic
	ctx := context.Background()
	for _, ids := range [][]string{nil, {}, {"", "  "}} {
		services, err := m.ServiceListByIDs(ctx, ids)
		if assert.NoError(t, err) {
			assert.Empty(t, services.Items)
		}
		nodes, err := m.NodeListByIDs(ctx, ids)
		if assert.NoError(t, err) {
			assert.Empty(t, nodes.Items)
		}
		networks, err := m.NetworkListByIDs(ctx, ids)
		if assert.NoError(t, err) {
			assert.Empty(t, networks.Items)
		}
		volumes, err := m.VolumeListByIDs(ctx, ids)
		if assert.NoError(t, err) {
			assert.Empty(t, volumes.Items)
		}
	}
}

// The ids asked for: "id1" twice, and a blank one dropped.
var askedIDs = []string{"id1", "id2", "id1", ""}

// Services, nodes and networks are asked for in one list, with an id filter
// Docker ORs; what it answers by prefix only - "id10" for "id1" - is dropped,
// and the caller's options still apply.
func TestServicesByIDsAreAskedOnceAndMatchedExactly(t *testing.T) {
	m, d := newDaemon(t, "/services", []swarm.Service{{ID: "id1"}, {ID: "id10"}, {ID: "id2"}})
	resp, err := m.ServiceListByIDs(context.Background(), askedIDs, func(opts *client.ServiceListOptions) {
		opts.Status = true
		FilterAdd(&opts.Filters, "label", "team=a")
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	d.assertAskedByID(t)
	assert.Equal(t, "true", d.status)
	assert.ElementsMatch(t, []string{"id1", "id2"}, idsOf(resp.Items, func(s *swarm.Service) string { return s.ID }))
}

func TestNodesByIDsAreAskedOnceAndMatchedExactly(t *testing.T) {
	m, d := newDaemon(t, "/nodes", []swarm.Node{{ID: "id1"}, {ID: "id10"}, {ID: "id2"}})
	resp, err := m.NodeListByIDs(context.Background(), askedIDs, func(opts *client.NodeListOptions) {
		FilterAdd(&opts.Filters, "label", "team=a")
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	d.assertAskedByID(t)
	assert.ElementsMatch(t, []string{"id1", "id2"}, idsOf(resp.Items, func(n *swarm.Node) string { return n.ID }))
}

func TestNetworksByIDsAreAskedOnceAndMatchedExactly(t *testing.T) {
	m, d := newDaemon(t, "/networks", []network.Summary{
		{Network: network.Network{ID: "id1"}}, {Network: network.Network{ID: "id10"}}, {Network: network.Network{ID: "id2"}},
	})
	resp, err := m.NetworkListByIDs(context.Background(), askedIDs, func(opts *client.NetworkListOptions) {
		FilterAdd(&opts.Filters, "label", "team=a")
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	d.assertAskedByID(t)
	assert.ElementsMatch(t, []string{"id1", "id2"}, idsOf(resp.Items, func(n *network.Summary) string { return n.ID }))
}

// Volumes are listed once and matched on their id: a cluster volume's cluster
// id, any other's name. Docker has no filter for a cluster id, so none is sent
// but the caller's.
func TestVolumesByIDsAreListedOnceAndMatchedOnTheirID(t *testing.T) {
	m, d := newDaemon(t, "/volumes", volume.ListResponse{Volumes: []volume.Volume{
		{Name: "vol1"}, {Name: "vol10"},
		{Name: "named", ClusterVolume: &volume.ClusterVolume{ID: "cv1"}},
		{Name: "cv2", ClusterVolume: &volume.ClusterVolume{ID: "other"}},
	}})
	resp, err := m.VolumeListByIDs(context.Background(), []string{"vol1", "cv1", "cv2", "named", ""},
		func(opts *client.VolumeListOptions) { FilterAdd(&opts.Filters, "label", "team=a") })
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, 1, d.calls)
	assert.Equal(t, map[string]map[string]bool{"label": {"team=a": true}}, d.filters)
	assert.ElementsMatch(t, []string{"vol1", "named"}, idsOf(resp.Items, func(v *volume.Volume) string { return v.Name }))
}

// daemon answers one kind of list with what it is given, and keeps what it was
// asked.
type daemon struct {
	calls   int
	filters map[string]map[string]bool
	status  string
}

func newDaemon(t *testing.T, path string, answer any) (*manager, *daemon) {
	t.Helper()
	d := &daemon{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, path) {
			http.NotFound(w, r)
			return
		}
		d.calls++
		d.filters = nil
		_ = json.Unmarshal([]byte(r.URL.Query().Get("filters")), &d.filters)
		d.status = r.URL.Query().Get("status")
		_ = json.NewEncoder(w).Encode(answer)
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.WithHost("tcp://"+srv.Listener.Addr().String()), client.WithAPIVersion(minAPIVersion))
	if err != nil {
		t.Fatal(err)
	}
	return &manager{client: c}, d
}

// assertAskedByID checks the one list asked for askedIDs by id, with the
// caller's label filter beside them.
func (d *daemon) assertAskedByID(t *testing.T) {
	t.Helper()
	assert.Equal(t, 1, d.calls)
	assert.Equal(t, map[string]map[string]bool{
		"id":    {"id1": true, "id2": true},
		"label": {"team=a": true},
	}, d.filters)
}

func idsOf[T any](items []T, id func(*T) string) []string {
	out := make([]string, 0, len(items))
	for i := range items {
		out = append(out, id(&items[i]))
	}
	return out
}
