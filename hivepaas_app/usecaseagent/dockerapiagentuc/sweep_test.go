package dockerapiagentuc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var sweepNow = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

// fakeNode is a node's daemon as far as a sweep asks: the listings honor the
// label and name filters a sweep uses, and removals are recorded.
type fakeNode struct {
	docker.Manager
	containers []container.Summary
	networks   []network.Summary
	attached   map[string]int
	volumes    []volume.Volume
	removed    []string
}

func owned(appID string) map[string]string {
	return map[string]string{dockerproxy.OwnerLabel: appID}
}

// matches applies a filters value of the kinds a sweep sends: "label" as key or
// key=value, and "name" as a substring.
func matches(filters client.Filters, name string, labels map[string]string) bool {
	for want := range filters["label"] {
		key, value, withValue := strings.Cut(want, "=")
		got, found := labels[key]
		if !found || (withValue && got != value) {
			return false
		}
	}
	for want := range filters["name"] {
		if !strings.Contains(name, want) {
			return false
		}
	}
	return true
}

func (f *fakeNode) ContainerList(_ context.Context, options ...docker.ContainerListOption) (
	*client.ContainerListResult, error) {
	opts := client.ContainerListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.ContainerListResult{}
	for _, c := range f.containers {
		if matches(opts.Filters, c.ID, c.Labels) {
			out.Items = append(out.Items, c)
		}
	}
	return out, nil
}

func (f *fakeNode) ContainerRemove(_ context.Context, id string, _ ...docker.ContainerRemoveOption) (
	*client.ContainerRemoveResult, error) {
	f.removed = append(f.removed, "container "+id)
	return &client.ContainerRemoveResult{}, nil
}

func (f *fakeNode) NetworkList(_ context.Context, options ...docker.NetworkListOption) (
	*client.NetworkListResult, error) {
	opts := client.NetworkListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.NetworkListResult{}
	for _, n := range f.networks {
		if matches(opts.Filters, n.Name, n.Labels) {
			out.Items = append(out.Items, n)
		}
	}
	return out, nil
}

func (f *fakeNode) NetworkInspect(_ context.Context, id string, _ ...docker.NetworkInspectOption) (
	*client.NetworkInspectResult, error) {
	attached := map[string]network.EndpointResource{}
	for i := range f.attached[id] {
		attached[string(rune('a'+i))] = network.EndpointResource{}
	}
	return &client.NetworkInspectResult{Network: network.Inspect{Containers: attached}}, nil
}

func (f *fakeNode) NetworkRemove(_ context.Context, id string, _ ...docker.NetworkRemoveOption) (
	*client.NetworkRemoveResult, error) {
	f.removed = append(f.removed, "network "+id)
	return &client.NetworkRemoveResult{}, nil
}

func (f *fakeNode) VolumeList(_ context.Context, options ...docker.VolumeListOption) (
	*client.VolumeListResult, error) {
	opts := client.VolumeListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.VolumeListResult{}
	for _, v := range f.volumes {
		if matches(opts.Filters, v.Name, v.Labels) {
			out.Items = append(out.Items, v)
		}
	}
	return out, nil
}

func (f *fakeNode) VolumeRemove(_ context.Context, id string, _ bool, _ ...docker.VolumeRemoveOption) (
	*client.VolumeRemoveResult, error) {
	f.removed = append(f.removed, "volume "+id)
	return &client.VolumeRemoveResult{}, nil
}

func networkSummary(id string, labels map[string]string, created time.Time) network.Summary {
	return network.Summary{Network: network.Network{ID: id, Name: id, Labels: labels, Created: created}}
}

// aNode holds, for app1 (which keeps its access) and app2 (which lost it):
// running, recently exited and long-exited children, fresh, idle and busy
// networks, cache volumes and socket volumes - and objects of nobody's.
func aNode() *fakeNode {
	hour := time.Hour
	return &fakeNode{
		containers: []container.Summary{
			{ID: "app1-running", Labels: owned("app1"), State: container.StateRunning,
				Created: sweepNow.Add(-48 * hour).Unix()},
			{ID: "app1-exited-now", Labels: owned("app1"), State: container.StateExited,
				Created: sweepNow.Add(-hour).Unix()},
			{ID: "app1-exited-old", Labels: owned("app1"), State: container.StateExited,
				Created: sweepNow.Add(-25 * hour).Unix()},
			{ID: "app2-running", Labels: owned("app2"), State: container.StateRunning,
				Created: sweepNow.Unix()},
			{ID: "hivepaas-db", Labels: map[string]string{}, State: container.StateExited,
				Created: sweepNow.Add(-100 * hour).Unix()},
		},
		networks: []network.Summary{
			networkSummary("app1-fresh", owned("app1"), sweepNow.Add(-10*time.Minute)),
			networkSummary("app1-idle", owned("app1"), sweepNow.Add(-2*hour)),
			networkSummary("app1-busy", owned("app1"), sweepNow.Add(-2*hour)),
			networkSummary("app2-net", owned("app2"), sweepNow),
			networkSummary("hivepaas_net", map[string]string{}, sweepNow.Add(-100*hour)),
		},
		attached: map[string]int{"app1-busy": 1},
		volumes: []volume.Volume{
			{Name: "app1-cache", Labels: owned("app1")},
			{Name: "app2-cache", Labels: owned("app2")},
			{Name: dockerapiservice.SocketVolumeName("app1"), Labels: map[string]string{}},
			{Name: dockerapiservice.SocketVolumeName("app2"), Labels: map[string]string{}},
			{Name: "hp-vol-data", Labels: map[string]string{}},
		},
	}
}

func TestCollectRemovesWhatIsLeftBehind(t *testing.T) {
	node := aNode()
	access := map[string]bool{"app1": true}
	s := &sweep{docker: node, now: sweepNow, ageOut: true, gone: func(appID string) bool { return !access[appID] }}

	removed, err := s.run(context.Background())
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"container app1-exited-old", "container app2-running",
		"network app1-idle", "network app2-net",
		"volume app2-cache", "volume " + dockerapiservice.SocketVolumeName("app2"),
	}, node.removed)
	assert.Equal(t, Removed{Containers: 2, Networks: 2, Volumes: 2}, removed)
}

func TestRemovingAnAppTouchesOnlyThatApp(t *testing.T) {
	node := aNode()
	s := &sweep{docker: node, now: sweepNow, only: "app1", gone: func(appID string) bool { return appID == "app1" }}

	removed, err := s.run(context.Background())
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"container app1-running", "container app1-exited-now", "container app1-exited-old",
		"network app1-fresh", "network app1-idle", "network app1-busy",
		"volume app1-cache", "volume " + dockerapiservice.SocketVolumeName("app1"),
	}, node.removed)
	assert.Equal(t, Removed{Containers: 3, Networks: 3, Volumes: 2}, removed)
}
