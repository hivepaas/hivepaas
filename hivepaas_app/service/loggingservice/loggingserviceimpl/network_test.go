package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

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
	// The backend is the only thing that bridges the two: the collector runs on
	// every node, and the internal network is where the database lives.
	assert.ElementsMatch(t, []string{logNet, base.NetworkHivepaasLocal}, networkTargets(backend))
	assert.Equal(t, []string{logNet}, networkTargets(collector))
	assert.NotContains(t, networkTargets(backend), base.NetworkGlobalRouting)
}

// Deploying a backend onto a network the API is not on would leave it running
// and unreachable, which looks like success everywhere except the logs page.
func TestDeployRefusesWhenTheInternalNetworkIsMissing(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, storedSetting(t, enabledConfig()))
	fd.networks = nil // no hivepaas_local_net in this cluster

	err := s.Apply(context.Background(), nil)

	assert.ErrorIs(t, err, loggingservice.ErrAPINetworkMissing)
	assert.Empty(t, fd.created, "nothing should have been deployed")
}
