package traefiksettingsdto

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestTransformStartupCommand_OpenPorts(t *testing.T) {
	traefikSvc := &swarm.Service{
		Spec: swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Args: []string{
						"traefik",
						"--entrypoints.web.address=:80",
						"--entrypoints.websecure.address=:443",
						"--entrypoints.tcp-svc-12312.address=:12312",
						"--entrypoints.tcp-svc-app-custom.address=:8443",
						"--entrypoints.tcp-svc-custom-tcp.address=:9000/tcp",
						"--entrypoints.udp-svc-10000.address=:10000/udp",
						"--entrypoints.udp-svc-dns.address=:53",
						"--log.level=INFO",
						"--accesslog=true",
						"--entrypoints.websecure.http3=true",
						"--experimental.fastproxy=true",
						"--providers.docker=true",
					},
				},
			},
		},
	}

	resp := TransformStartupCommand(traefikSvc)

	assert.Equal(t, "INFO", resp.LogLevel)
	assert.True(t, resp.AccessLog)
	assert.True(t, resp.HTTP3)
	assert.True(t, resp.FastProxy)

	expectedOpenPorts := []string{
		"12312/tcp",
		"8443/tcp",
		"9000/tcp",
		"10000/udp",
		"53/udp",
	}
	assert.Equal(t, expectedOpenPorts, resp.OpenPorts)

	assert.Equal(t, []string{"--providers.docker=true"}, resp.Args)
}
