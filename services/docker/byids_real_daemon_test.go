package docker

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

const unknownID = "zzzzzzzzzzzzzzzzzzzzzzzzz"

// TestByIDsAgainstARealSwarm creates three services, two networks and two volumes
// on the daemon DOCKER_HOST names - this machine's without it - and removes them.
// The services have no replica, so nothing runs and no image is pulled. It runs
// only when HP_TEST_DOCKER_SWARM is set.
func TestByIDsAgainstARealSwarm(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER_SWARM") == "" {
		t.Skip("set HP_TEST_DOCKER_SWARM=1 to create services, networks and volumes on DOCKER_HOST's swarm")
	}
	c, err := client.New(client.FromEnv, client.WithAPIVersion(minAPIVersion))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	m := &manager{client: c}
	t.Cleanup(func() { _ = m.Close() })
	prefix := "hp-itest-" + strconv.FormatInt(time.Now().UnixNano(), 36) + "-"

	t.Run("services", func(t *testing.T) { testServicesByIDs(t, m, prefix) })
	t.Run("nodes", func(t *testing.T) { testNodesByIDs(t, m) })
	t.Run("networks", func(t *testing.T) { testNetworksByIDs(t, m, prefix) })
	t.Run("volumes", func(t *testing.T) { testVolumesByIDs(t, m, prefix) })
}

func testServicesByIDs(t *testing.T, m *manager, prefix string) {
	ctx := context.Background()
	ids := make([]string, 0, 3) //nolint:mnd
	for _, name := range []string{"a", "b", "c"} {
		zero := uint64(0)
		created, err := m.ServiceCreate(ctx, &swarm.ServiceSpec{
			Annotations:  swarm.Annotations{Name: prefix + name, Labels: map[string]string{"hivepaas.itest": "1"}},
			TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:3"}},
			Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &zero}},
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		t.Cleanup(func() { _, _ = m.ServiceRemove(context.Background(), created.ID) })
		ids = append(ids, created.ID)
	}
	a, b, c := ids[0], ids[1], ids[2]

	// What Docker does with id filters, which the ...ByIDs functions build on:
	// the values of one filter are ORed, and each is a prefix - an empty one of
	// all.
	listed := func(values ...string) []string {
		resp, err := m.ServiceList(ctx, func(opts *client.ServiceListOptions) {
			for _, v := range values {
				FilterAdd(&opts.Filters, "id", v)
			}
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return idsOf(resp.Items, func(s *swarm.Service) string { return s.ID })
	}
	both := listed(a, b)
	assert.Contains(t, both, a)
	assert.Contains(t, both, b)
	assert.NotContains(t, both, c)
	assert.Contains(t, listed(a[:5]), a)
	assert.Subset(t, listed(""), ids)

	resp, err := m.ServiceListByIDs(ctx, []string{a, b, a[:5], prefix + "a", "", unknownID},
		func(opts *client.ServiceListOptions) { opts.Status = true })
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	for i := range resp.Items {
		assert.NotNil(t, resp.Items[i].ServiceStatus)
	}
	assert.ElementsMatch(t, []string{a, b}, idsOf(resp.Items, func(s *swarm.Service) string { return s.ID }))
}

// A node is matched on its id only: not on a prefix of it, and not on its
// hostname, which Docker also takes for its name.
func testNodesByIDs(t *testing.T, m *manager) {
	ctx := context.Background()
	info, err := m.SystemInfo(ctx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	id := info.Info.Swarm.NodeID
	node, err := m.NodeInspect(ctx, id)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	hostname := node.Node.Description.Hostname
	for _, tc := range []struct{ asked, want []string }{
		{[]string{id}, []string{id}},
		{[]string{id[:4]}, nil},
		{[]string{hostname}, nil},
		{[]string{""}, nil},
		{[]string{id, id[:4], hostname, "", unknownID}, []string{id}},
	} {
		resp, err := m.NodeListByIDs(ctx, tc.asked)
		if assert.NoError(t, err, tc.asked) {
			assert.ElementsMatch(t, tc.want, idsOf(resp.Items, func(n *swarm.Node) string { return n.ID }), tc.asked)
		}
	}
}

func testNetworksByIDs(t *testing.T, m *manager, prefix string) {
	ctx := context.Background()
	ids := make([]string, 0, 2) //nolint:mnd
	for _, name := range []string{"net-a", "net-b"} {
		created, err := m.NetworkCreate(ctx, prefix+name, func(opts *client.NetworkCreateOptions) {
			opts.Driver = NetworkDriverOverlay
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		t.Cleanup(func() { _, _ = m.NetworkRemove(context.Background(), created.ID) })
		ids = append(ids, created.ID)
	}
	a, b := ids[0], ids[1]
	for _, tc := range []struct{ asked, want []string }{
		{[]string{a}, []string{a}},
		{[]string{a[:5]}, nil},
		{[]string{prefix + "net-a"}, nil},
		{[]string{""}, nil},
		{[]string{a, b, a[:5], prefix + "net-a", "", unknownID}, []string{a, b}},
	} {
		resp, err := m.NetworkListByIDs(ctx, tc.asked)
		if assert.NoError(t, err, tc.asked) {
			assert.ElementsMatch(t, tc.want, idsOf(resp.Items, func(n *network.Summary) string { return n.ID }),
				tc.asked)
		}
	}
}

// A local volume's id is its name.
func testVolumesByIDs(t *testing.T, m *manager, prefix string) {
	ctx := context.Background()
	names := []string{prefix + "vol-a", prefix + "vol-b"}
	for _, name := range names {
		_, err := m.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) { opts.Name = name })
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		t.Cleanup(func() { _, _ = m.VolumeRemove(context.Background(), name, true) })
	}
	for _, tc := range []struct{ asked, want []string }{
		{[]string{names[0]}, names[:1]},
		{[]string{""}, nil},
		{[]string{names[0], names[1], prefix + "vol", "", unknownID}, names},
	} {
		resp, err := m.VolumeListByIDs(ctx, tc.asked)
		if assert.NoError(t, err, tc.asked) {
			assert.ElementsMatch(t, tc.want, idsOf(resp.Items, func(v *volume.Volume) string { return v.Name }), tc.asked)
		}
	}
}
