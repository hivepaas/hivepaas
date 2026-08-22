package traefikserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestParsePortString(t *testing.T) {
	t.Run("standard TCP port", func(t *testing.T) {
		p, ok := parsePortString("9000")
		assert.True(t, ok)
		assert.Equal(t, uint32(9000), p.portNum)
		assert.Equal(t, network.TCP, p.protocol)
		assert.Equal(t, "entrypoints.tcp-svc-9000.address", p.argKey)
		assert.Equal(t, "--entrypoints.tcp-svc-9000.address=:9000", p.argString)
		assert.Equal(t, "9000/tcp", p.key())
	})

	t.Run("TCP port with prefix and protocol", func(t *testing.T) {
		p, ok := parsePortString(":9000/tcp")
		assert.True(t, ok)
		assert.Equal(t, uint32(9000), p.portNum)
		assert.Equal(t, network.TCP, p.protocol)
		assert.Equal(t, "--entrypoints.tcp-svc-9000.address=:9000", p.argString)
	})

	t.Run("UDP port", func(t *testing.T) {
		p, ok := parsePortString("10000/udp")
		assert.True(t, ok)
		assert.Equal(t, uint32(10000), p.portNum)
		assert.Equal(t, network.UDP, p.protocol)
		assert.Equal(t, "entrypoints.udp-svc-10000.address", p.argKey)
		assert.Equal(t, "--entrypoints.udp-svc-10000.address=:10000/udp", p.argString)
		assert.Equal(t, "10000/udp", p.key())
	})

	t.Run("invalid inputs", func(t *testing.T) {
		_, ok := parsePortString("")
		assert.False(t, ok)
		_, ok = parsePortString("abc")
		assert.False(t, ok)
		_, ok = parsePortString("0")
		assert.False(t, ok)
	})
}

func TestProcessPortConfigInArgs(t *testing.T) {
	currentArgs := []string{
		"--api.dashboard=true",
		"--entrypoints.web.address=:80",
		"--entrypoints.websecure.address=:443",
		"--entrypoints.tcp-svc-9000.address=:9000",
		"--entrypoints.udp-svc-10000.address=:10000/udp",
	}

	t.Run("no changes", func(t *testing.T) {
		newArgs, hasChanges := processPortConfigInArgs(currentArgs, []string{"9000/tcp", "10000/udp"}, nil)
		assert.False(t, hasChanges)
		assert.Equal(t, currentArgs, newArgs)
	})

	t.Run("open new TCP and UDP ports", func(t *testing.T) {
		newArgs, hasChanges := processPortConfigInArgs(currentArgs, []string{"9001", "10001/udp"}, nil)
		assert.True(t, hasChanges)
		assert.Contains(t, newArgs, "--entrypoints.tcp-svc-9001.address=:9001")
		assert.Contains(t, newArgs, "--entrypoints.udp-svc-10001.address=:10001/udp")
		assert.Contains(t, newArgs, "--entrypoints.tcp-svc-9000.address=:9000")
		assert.Contains(t, newArgs, "--entrypoints.udp-svc-10000.address=:10000/udp")
		assert.Contains(t, newArgs, "--entrypoints.web.address=:80")
	})

	t.Run("close existing ports", func(t *testing.T) {
		newArgs, hasChanges := processPortConfigInArgs(currentArgs, nil, []string{"9000/tcp", "10000/udp"})
		assert.True(t, hasChanges)
		assert.NotContains(t, newArgs, "--entrypoints.tcp-svc-9000.address=:9000")
		assert.NotContains(t, newArgs, "--entrypoints.udp-svc-10000.address=:10000/udp")
		assert.Contains(t, newArgs, "--entrypoints.web.address=:80")
	})
}

func TestProcessPortConfigInEndpointSpec(t *testing.T) {
	endpointSpec := &swarm.EndpointSpec{
		Ports: []swarm.PortConfig{
			{TargetPort: 80, PublishedPort: 80, Protocol: network.TCP},
			{TargetPort: 443, PublishedPort: 443, Protocol: network.TCP},
			{TargetPort: 9000, PublishedPort: 9000, Protocol: network.TCP},
			{TargetPort: 10000, PublishedPort: 10000, Protocol: network.UDP},
		},
	}

	t.Run("no changes", func(t *testing.T) {
		newPorts, hasChanges := processPortConfigInEndpointSpec(endpointSpec, []string{"9000/tcp", "10000/udp"}, nil)
		assert.False(t, hasChanges)
		assert.Equal(t, len(endpointSpec.Ports), len(newPorts))
	})

	t.Run("open new ports", func(t *testing.T) {
		newPorts, hasChanges := processPortConfigInEndpointSpec(endpointSpec, []string{"9001", "10001/udp"}, nil)
		assert.True(t, hasChanges)
		assert.Equal(t, 6, len(newPorts))

		var foundTCP9001, foundUDP10001 bool
		for _, p := range newPorts {
			if p.PublishedPort == 9001 && p.Protocol == network.TCP {
				foundTCP9001 = true
			}
			if p.PublishedPort == 10001 && p.Protocol == network.UDP {
				foundUDP10001 = true
			}
		}
		assert.True(t, foundTCP9001)
		assert.True(t, foundUDP10001)
	})

	t.Run("close ports", func(t *testing.T) {
		newPorts, hasChanges := processPortConfigInEndpointSpec(endpointSpec, nil, []string{"9000/tcp", "10000/udp"})
		assert.True(t, hasChanges)
		assert.Equal(t, 2, len(newPorts))
		assert.Equal(t, uint32(80), newPorts[0].PublishedPort)
		assert.Equal(t, uint32(443), newPorts[1].PublishedPort)
	})
}
