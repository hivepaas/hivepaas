package dockerapiserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeNetworks holds networks by name, and records what is created and removed.
type fakeNetworks struct {
	docker.Manager
	byName  map[string]string
	created map[string]client.NetworkCreateOptions
	removed []string
}

func (f *fakeNetworks) NetworkInspect(_ context.Context, name string, _ ...docker.NetworkInspectOption) (
	*client.NetworkInspectResult, error) {
	id, found := f.byName[name]
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	return &client.NetworkInspectResult{Network: network.Inspect{Network: network.Network{ID: id, Name: name}}}, nil
}

func (f *fakeNetworks) NetworkCreate(_ context.Context, name string, options ...docker.NetworkCreateOption) (
	*client.NetworkCreateResult, error) {
	opts := client.NetworkCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	f.created[name] = opts
	f.byName[name] = "id-" + name
	return &client.NetworkCreateResult{ID: "id-" + name}, nil
}

func (f *fakeNetworks) NetworkRemove(_ context.Context, name string, _ ...docker.NetworkRemoveOption) (
	*client.NetworkRemoveResult, error) {
	f.removed = append(f.removed, name)
	return &client.NetworkRemoveResult{}, nil
}

func accessService(settings ...*entity.Setting) (*service, *fakeNetworks) {
	networks := &fakeNetworks{byName: map[string]string{}, created: map[string]client.NetworkCreateOptions{}}
	return &service{settingRepo: &fakeSettingRepo{settings: settings}, dockerManager: networks}, networks
}

func TestEnsureNetworkCreatesTheAppsNetworkOnce(t *testing.T) {
	svc, networks := accessService()
	id, err := svc.EnsureNetwork(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Equal(t, "id-hp-dapi-app1", id)
	created := networks.created["hp-dapi-app1"]
	assert.Equal(t, docker.NetworkDriverOverlay, created.Driver)
	assert.True(t, created.Attachable)
	assert.Equal(t, map[string]string{dockerapiservice.NetworkLabel: "app1"}, created.Labels)

	networks.created = map[string]client.NetworkCreateOptions{}
	id, err = svc.EnsureNetwork(context.Background(), "app1")
	assert.NoError(t, err)
	assert.Equal(t, "id-hp-dapi-app1", id)
	assert.Empty(t, networks.created)
}

func TestApplyToServiceAttachesAnAppWithAccess(t *testing.T) {
	svc, _ := accessService(dockerAPISetting("app1", `{"images":["alpine"]}`))
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}}
	assert.NoError(t, svc.ApplyToService(context.Background(), nil, "app1", spec))
	assert.Equal(t, []mount.Mount{dockerapiservice.SocketMount("app1")}, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "id-hp-dapi-app1"}}, spec.TaskTemplate.Networks)
}

func TestApplyToServiceDetachesAnAppWithout(t *testing.T) {
	svc, networks := accessService()
	networks.byName["hp-dapi-app1"] = "id-hp-dapi-app1"
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{dockerapiservice.SocketMount("app1")}},
		Networks:      []swarm.NetworkAttachmentConfig{{Target: "id-hp-dapi-app1"}},
	}}
	assert.NoError(t, svc.ApplyToService(context.Background(), nil, "app1", spec))
	assert.Empty(t, spec.TaskTemplate.ContainerSpec.Mounts)
	assert.Empty(t, spec.TaskTemplate.Networks)
}

// fakeCluster is a cluster's nodes and networks at once, for a call that needs
// both.
type fakeCluster struct {
	fakeNetworks
	nodes []swarm.Node
}

func (f *fakeCluster) NodeList(_ context.Context, _ ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{Items: f.nodes}, nil
}

func TestRemoveAppRemovesItsNetworkAndAsksTheAgents(t *testing.T) {
	var calls []string
	svc := fanOutService(&calls, "")
	cluster := &fakeCluster{fakeNetworks: fakeNetworks{byName: map[string]string{}},
		nodes: svc.dockerManager.(*fakeNodes).nodes}
	svc.dockerManager = cluster
	networks := &cluster.fakeNetworks

	assert.NoError(t, svc.RemoveApp(context.Background(), "app1"))
	assert.Equal(t, []string{"remove app1 agent-on-n1", "remove app1 agent-on-n2"}, calls)
	assert.Equal(t, []string{"hp-dapi-app1"}, networks.removed)
}
