package traefikserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func TestApplyAppConfig(t *testing.T) {
	s := &service{}

	t.Run("HTTP and TCP domains", func(t *testing.T) {
		app := &entity.App{
			Key:       "my_app",
			ServiceID: "svc-123",
		}
		service := &swarm.Service{
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{
					Labels: map[string]string{
						"traefik.http.routers.old.rule": "Host(`old.com`)",
					},
				},
			},
		}

		routingSettings := &entity.AppRoutingSettings{
			Port:           8080,
			ExposePublicly: true,
			Domains: []*entity.AppDomain{
				{
					Enabled:       true,
					Domain:        "app.myapp.com",
					Protocol:      base.NetworkProtocolHTTP,
					ContainerPort: 8080,
				},
				{
					Enabled:        true,
					Domain:         "db.myapp.com",
					Protocol:       base.NetworkProtocolTCP,
					ContainerPort:  5432,
					TLSPassthrough: true,
				},
			},
		}

		req := &traefikservice.ApplyAppConfigReq{
			App:             app,
			Service:         service,
			RoutingSettings: routingSettings,
			RefObjects:      entity.NewRefObjects(),
		}

		resp, err := s.ApplyAppConfig(context.Background(), nil, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		labels := resp.Service.Spec.Labels
		assert.Equal(t, labelValueTrue, labels["traefik.enable"])
		assert.Equal(t, base.NetworkGlobalRouting, labels["traefik.swarm.network"])

		// Old HTTP router cleaned
		assert.NotContains(t, labels, "traefik.http.routers.old.rule")

		// 1. HTTP Router
		assert.Equal(t, "Host(`app.myapp.com`)", labels["traefik.http.routers.router-my-app-0.rule"])
		assert.Equal(t, "websecure", labels["traefik.http.routers.router-my-app-0.entrypoints"])
		assert.Equal(t, "svc-my-app-0", labels["traefik.http.routers.router-my-app-0.service"])
		assert.Equal(t, labelValueTrue, labels["traefik.http.routers.router-my-app-0.tls"])
		assert.Equal(t, "8080", labels["traefik.http.services.svc-my-app-0.loadbalancer.server.port"])

		// 2. TCP Router with HostSNI
		assert.Equal(t, "HostSNI(`db.myapp.com`)", labels["traefik.tcp.routers.tcp-router-my-app-1.rule"])
		assert.Equal(t, "tcp-svc-5432", labels["traefik.tcp.routers.tcp-router-my-app-1.entrypoints"])
		assert.Equal(t, "tcp-svc-my-app-1", labels["traefik.tcp.routers.tcp-router-my-app-1.service"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-my-app-1.tls"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-my-app-1.tls.passthrough"])
		assert.Equal(t, "5432", labels["traefik.tcp.services.tcp-svc-my-app-1.loadbalancer.server.port"])
	})

	t.Run("ExposePublicly disabled cleans labels", func(t *testing.T) {
		app := &entity.App{
			Key: "my_app",
		}
		service := &swarm.Service{
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{
					Labels: map[string]string{
						"traefik.enable":                "true",
						"traefik.http.routers.old.rule": "Host(`old.com`)",
						"traefik.tcp.routers.old.rule":  "HostSNI(`old.com`)",
						"custom.label":                  "preserve-me",
						"traefik.x-custom-header":       "preserve-me-too",
					},
				},
			},
		}

		req := &traefikservice.ApplyAppConfigReq{
			App:             app,
			Service:         service,
			RoutingSettings: &entity.AppRoutingSettings{ExposePublicly: false},
			RefObjects:      entity.NewRefObjects(),
		}

		resp, err := s.ApplyAppConfig(context.Background(), nil, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		labels := resp.Service.Spec.Labels
		assert.NotContains(t, labels, "traefik.enable")
		assert.NotContains(t, labels, "traefik.http.routers.old.rule")
		assert.NotContains(t, labels, "traefik.tcp.routers.old.rule")
		assert.Equal(t, "preserve-me", labels["custom.label"])
		assert.Equal(t, "preserve-me-too", labels["traefik.x-custom-header"])
	})
}
