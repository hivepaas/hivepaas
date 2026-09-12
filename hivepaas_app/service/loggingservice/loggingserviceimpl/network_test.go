package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	testRoutingNetID = "net-routing"
	testLocalNetID   = "net-local"
)

// fakeHpApp is HivePaaS's own service, attached to the routing network and one
// private network - the shape deployment/local/hivepaas.yaml gives it.
type fakeHpApp struct {
	hpappservice.Service
	networks []string
}

func (f *fakeHpApp) GetHpAppSwarmService(_ context.Context) (*swarm.Service, error) {
	svc := &swarm.Service{}
	for _, n := range f.networks {
		svc.Spec.TaskTemplate.Networks = append(svc.Spec.TaskTemplate.Networks,
			swarm.NetworkAttachmentConfig{Target: n})
	}
	return svc, nil
}

type fakeNetworkService struct {
	networkservice.Service
}

func (f *fakeNetworkService) GetGlobalRoutingNetworkID(_ context.Context) (string, error) {
	return testRoutingNetID, nil
}

func (f *fakeDocker) NetworkList(
	_ context.Context, _ ...docker.NetworkListOption,
) (*client.NetworkListResult, error) {
	return &client.NetworkListResult{Items: f.networks}, nil
}

func (f *fakeDocker) NetworkCreate(
	_ context.Context, name string, opts ...docker.NetworkCreateOption,
) (*client.NetworkCreateResult, error) {
	o := client.NetworkCreateOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	f.networksCreated = append(f.networksCreated, o)
	id := "id-" + name
	f.networks = append(f.networks, network.Summary{Network: network.Network{ID: id, Name: name}})
	return &client.NetworkCreateResult{ID: id}, nil
}

func networkTargets(spec *swarm.ServiceSpec) []string {
	out := []string{}
	for _, n := range spec.TaskTemplate.Networks {
		out = append(out, n.Target)
	}
	return out
}

func TestEnsureLoggingNetworkCreatesAnOverlayOnce(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)

	id1, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)
	id2, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, id1, id2)
	if assert.Len(t, fd.networksCreated, 1) {
		assert.Equal(t, docker.NetworkDriverOverlay, fd.networksCreated[0].Driver)
		assert.Equal(t, docker.NetworkScopeSwarm, fd.networksCreated[0].Scope)
		assert.False(t, fd.networksCreated[0].Attachable, "nothing outside the stack may join it")
	}
}

func TestEnsureLoggingNetworkIgnoresANameThatOnlyContainsIt(t *testing.T) {
	// docker's name filter matches substrings
	fd := &fakeDocker{networks: []network.Summary{{Network: network.Network{
		ID: "other", Name: "p1_" + NetworkLogging + "_copy",
	}}}}
	s := newTestService(fd, nil)

	id, err := s.ensureLoggingNetwork(context.Background())
	assert.NoError(t, err)
	assert.NotEqual(t, "other", id)
	assert.Len(t, fd.networksCreated, 1)
}

func TestAPIPrivateNetworksLeavesTheRoutingNetworkOut(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	nets, err := s.apiPrivateNetworks(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{testLocalNetID}, nets)
}

func TestAPIPrivateNetworksRefusesWhenOnlyTheRoutingNetworkIsLeft(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	s.hpAppService = &fakeHpApp{networks: []string{testRoutingNetID}}

	_, err := s.apiPrivateNetworks(context.Background())
	assert.ErrorIs(t, err, loggingservice.ErrAPINetworkMissing)
}

func TestDeployJoinsTheStackOnlyToNetworksItNeeds(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, storedSetting(t, enabledConfig()))

	assert.NoError(t, s.Apply(context.Background(), nil))

	var backend, collector *swarm.ServiceSpec
	for _, spec := range fd.created {
		switch spec.Name {
		case ServiceNameBackend:
			backend = spec
		case ServiceNameCollector:
			collector = spec
		}
	}
	if backend == nil || collector == nil {
		t.Fatalf("expected both services, created %d", len(fd.created))
	}
	logNet := "id-" + NetworkLogging
	assert.ElementsMatch(t, []string{logNet, testLocalNetID}, networkTargets(backend))
	assert.Equal(t, []string{logNet}, networkTargets(collector))
	assert.NotContains(t, networkTargets(backend), testRoutingNetID)
}
