package hpappserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker hands back one service and records what it was asked to update.
// Embedding the interface means only the two methods this path calls need a
// body; anything else panics loudly rather than passing quietly.
type fakeDocker struct {
	docker.Manager
	service  *swarm.Service
	updates  int
	replicas uint64
}

func (f *fakeDocker) ServiceGetByName(
	_ context.Context, _ string, _ bool,
) (*swarm.Service, error) {
	return f.service, nil
}

func (f *fakeDocker) ServiceUpdate(
	_ context.Context, _ string, _ *swarm.Version, spec *swarm.ServiceSpec,
	_ ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	f.record(spec)
	return &client.ServiceUpdateResult{}, nil
}

// ServiceUpdateFunc mirrors what the manager does: hand the callback the service,
// respect its "nothing to do" answer, and only then record an update.
func (f *fakeDocker) ServiceUpdateFunc(
	_ context.Context, _ string, service *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), _ int, _ time.Duration,
	_ ...docker.ServiceUpdateOption,
) error {
	apply, err := fn(0, service)
	if err != nil || !apply {
		return err
	}
	f.record(&service.Spec)
	return nil
}

func (f *fakeDocker) record(spec *swarm.ServiceSpec) {
	f.updates++
	f.replicas = *spec.Mode.Replicated.Replicas
}

func appServiceAt(replicas uint64) *swarm.Service {
	return &swarm.Service{
		ID: "hivepaas_app",
		Spec: swarm.ServiceSpec{
			Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &replicas}},
		},
	}
}

// The state this exists for: an update stopped the app and never put it back,
// so nothing is running that could be asked to fix it.
func TestEnsureHpAppRunningBringsBackAStoppedApp(t *testing.T) {
	f := &fakeDocker{service: appServiceAt(0)}
	s := &service{dockerManager: f}

	scaled, err := s.EnsureHpAppRunning(context.Background())

	assert.NoError(t, err)
	assert.True(t, scaled)
	assert.Equal(t, uint64(1), f.replicas)
}

// It runs after every update, successful ones included, so doing nothing has to
// be the common case - not a second update of a service that is already fine.
func TestEnsureHpAppRunningLeavesARunningAppAlone(t *testing.T) {
	f := &fakeDocker{service: appServiceAt(3)}
	s := &service{dockerManager: f}

	scaled, err := s.EnsureHpAppRunning(context.Background())

	assert.NoError(t, err)
	assert.False(t, scaled)
	assert.Zero(t, f.updates)
}

// Not a shape HivePaaS deploys, but reading replicas through it would panic in
// the one path that must never panic.
func TestEnsureHpAppRunningIgnoresANonReplicatedService(t *testing.T) {
	f := &fakeDocker{service: &swarm.Service{ID: "hivepaas_app"}}
	s := &service{dockerManager: f}

	scaled, err := s.EnsureHpAppRunning(context.Background())

	assert.NoError(t, err)
	assert.False(t, scaled)
	assert.Zero(t, f.updates)
}
