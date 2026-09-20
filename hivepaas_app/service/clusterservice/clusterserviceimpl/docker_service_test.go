package clusterserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeDockerManager struct {
	docker.Manager
	services []swarm.Service
}

func (f *fakeDockerManager) ServiceList(
	_ context.Context, _ ...docker.ServiceListOption,
) (*client.ServiceListResult, error) {
	return &client.ServiceListResult{Items: f.services}, nil
}

func publishing(name string, ports ...swarm.PortConfig) swarm.Service {
	return swarm.Service{
		ID: "svc-" + name,
		Spec: swarm.ServiceSpec{
			Annotations:  swarm.Annotations{Name: name},
			EndpointSpec: &swarm.EndpointSpec{Ports: ports},
		},
	}
}

func portsService(services ...swarm.Service) *service {
	return &service{dockerManager: &fakeDockerManager{services: services}}
}

func TestVerifyPortsAvailableAllowsAFreePort(t *testing.T) {
	svc := portsService(publishing("traefik",
		swarm.PortConfig{PublishedPort: 443, Protocol: network.TCP}))

	assert.NoError(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}, nil))
}

func TestVerifyPortsAvailableRefusesAPortAnotherServiceHas(t *testing.T) {
	svc := portsService(publishing("project_dev_vpn",
		swarm.PortConfig{PublishedPort: 51820, Protocol: network.UDP}))

	err := svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}, nil)

	assert.ErrorIs(t, err, hperrors.ErrPortInUse)
	var hpErr hperrors.HPError
	assert.ErrorAs(t, err, &hpErr)
	assert.Contains(t, hpErr.Build("en").Detail, "project_dev_vpn")
}

// A port is only taken for the protocol it was published with: a DNS server
// answers on 53 over both, and two apps may split them.
func TestVerifyPortsAvailableSeparatesTheProtocols(t *testing.T) {
	svc := portsService(publishing("dns",
		swarm.PortConfig{PublishedPort: 53, Protocol: network.UDP}))

	assert.NoError(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 53, Protocol: network.TCP}}, nil))
	assert.Error(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 53, Protocol: network.UDP}}, nil))
}

// A port with no protocol is tcp, as docker reads it, on both sides of the
// comparison.
func TestVerifyPortsAvailableTreatsAnEmptyProtocolAsTCP(t *testing.T) {
	svc := portsService(publishing("ssh", swarm.PortConfig{PublishedPort: 2222}))

	assert.ErrorIs(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 2222, Protocol: network.TCP}}, nil), hperrors.ErrPortInUse)
	assert.ErrorIs(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 2222}}, nil), hperrors.ErrPortInUse)
}

// An app keeping the port it already has is not taking it from anybody.
func TestVerifyPortsAvailableIgnoresTheServiceBeingUpdated(t *testing.T) {
	svc := portsService(publishing("vpn",
		swarm.PortConfig{PublishedPort: 51820, Protocol: network.UDP}))

	assert.NoError(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}, []string{"svc-vpn"}))
}

// A published port of zero is docker choosing one, which claims nothing.
func TestVerifyPortsAvailableIgnoresPortsNobodyChose(t *testing.T) {
	svc := portsService(publishing("vpn",
		swarm.PortConfig{TargetPort: 51820, PublishedPort: 0, Protocol: network.UDP}))

	assert.NoError(t, svc.VerifyPortsAvailable(context.Background(),
		[]clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}, nil))
}
