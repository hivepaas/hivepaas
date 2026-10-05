package docker

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

// TestServicesByIDsAgainstARealSwarm creates three services on the swarm of the
// daemon DOCKER_HOST names - this machine's without it - and removes them. They
// have no replica, so nothing runs and no image is pulled. It runs only when
// HP_TEST_DOCKER_SWARM is set.
func TestServicesByIDsAgainstARealSwarm(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER_SWARM") == "" {
		t.Skip("set HP_TEST_DOCKER_SWARM=1 to create services on the swarm of DOCKER_HOST's daemon")
	}
	ctx := context.Background()
	c, err := client.New(client.FromEnv, client.WithAPIVersion(minAPIVersion))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	m := &manager{client: c}
	t.Cleanup(func() { _ = m.Close() })

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	ids := make([]string, 0, 3) //nolint:mnd
	for _, name := range []string{"a", "b", "c"} {
		zero := uint64(0)
		created, err := m.ServiceCreate(ctx, &swarm.ServiceSpec{
			Annotations: swarm.Annotations{Name: "hp-itest-" + suffix + "-" + name,
				Labels: map[string]string{"hivepaas.itest": "1"}},
			TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:3"}},
			Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &zero}},
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		t.Cleanup(func() { _, _ = m.ServiceRemove(context.Background(), created.ID) })
		ids = append(ids, created.ID)
	}
	a, b, c3 := ids[0], ids[1], ids[2]

	// What Docker does with id filters, which ServiceListByIDs builds on: the
	// values of one filter are ORed, and each is a prefix - an empty one of all.
	listed := func(values ...string) []string {
		resp, err := m.ServiceList(ctx, func(opts *client.ServiceListOptions) {
			for _, v := range values {
				FilterAdd(&opts.Filters, "id", v)
			}
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		out := make([]string, 0, len(resp.Items))
		for i := range resp.Items {
			out = append(out, resp.Items[i].ID)
		}
		return out
	}
	both := listed(a, b)
	assert.Contains(t, both, a)
	assert.Contains(t, both, b)
	assert.NotContains(t, both, c3)
	assert.Contains(t, listed(a[:5]), a)
	assert.Subset(t, listed(""), ids)

	// ServiceListByIDs keeps only the ids given, with their task counts.
	resp, err := m.ServiceListByIDs(ctx, []string{a, b, a[:5], "", "zzzzzzzzzzzzzzzzzzzzzzzzz"},
		func(opts *client.ServiceListOptions) { opts.Status = true })
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	got := make([]string, 0, len(resp.Items))
	for i := range resp.Items {
		got = append(got, resp.Items[i].ID)
		assert.NotNil(t, resp.Items[i].ServiceStatus)
	}
	assert.ElementsMatch(t, []string{a, b}, got)
}
