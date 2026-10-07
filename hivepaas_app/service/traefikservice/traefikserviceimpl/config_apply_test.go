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

	t.Run("HTTP, HTTP with TLS Passthrough, and TCP domains", func(t *testing.T) {
		app := &entity.App{
			ID:        "01K6APP",
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
					Domain:         "passthrough.myapp.com",
					Protocol:       base.NetworkProtocolHTTP,
					ContainerPort:  8443,
					TLSPassthrough: true,
					ForceHttps:     true,
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

		// 1. Standard HTTP Router
		assert.Equal(t, "Host(`app.myapp.com`)", labels["traefik.http.routers.router-01k6app-0.rule"])
		assert.Equal(t, "websecure", labels["traefik.http.routers.router-01k6app-0.entrypoints"])
		assert.Equal(t, "svc-01k6app-0", labels["traefik.http.routers.router-01k6app-0.service"])
		assert.Equal(t, labelValueTrue, labels["traefik.http.routers.router-01k6app-0.tls"])
		assert.Equal(t, "8080", labels["traefik.http.services.svc-01k6app-0.loadbalancer.server.port"])

		// 2. HTTP Domain with TLS Passthrough (routed via TCP router on websecure)
		assert.Equal(t, "HostSNI(`passthrough.myapp.com`)", labels["traefik.tcp.routers.tcp-router-01k6app-1.rule"])
		assert.Equal(t, "websecure", labels["traefik.tcp.routers.tcp-router-01k6app-1.entrypoints"])
		assert.Equal(t, "tcp-svc-01k6app-1", labels["traefik.tcp.routers.tcp-router-01k6app-1.service"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-01k6app-1.tls"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-01k6app-1.tls.passthrough"])
		assert.Equal(t, "8443", labels["traefik.tcp.services.tcp-svc-01k6app-1.loadbalancer.server.port"])
		// ForceHttps redirect router on web port 80
		assert.Equal(t, "Host(`passthrough.myapp.com`)", labels["traefik.http.routers.tcp-router-01k6app-1-forcehttps.rule"])
		assert.Equal(t, "web", labels["traefik.http.routers.tcp-router-01k6app-1-forcehttps.entrypoints"])

		// 3. TCP Router with HostSNI
		assert.Equal(t, "HostSNI(`db.myapp.com`)", labels["traefik.tcp.routers.tcp-router-01k6app-2.rule"])
		assert.Equal(t, "tcp-svc-5432", labels["traefik.tcp.routers.tcp-router-01k6app-2.entrypoints"])
		assert.Equal(t, "tcp-svc-01k6app-2", labels["traefik.tcp.routers.tcp-router-01k6app-2.service"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-01k6app-2.tls"])
		assert.Equal(t, labelValueTrue, labels["traefik.tcp.routers.tcp-router-01k6app-2.tls.passthrough"])
		assert.Equal(t, "5432", labels["traefik.tcp.services.tcp-svc-01k6app-2.loadbalancer.server.port"])
	})

	t.Run("TCP domain ending TLS accepts PostgreSQL's ALPN", func(t *testing.T) {
		data := &appConfigData{
			ApplyAppConfigReq: &traefikservice.ApplyAppConfigReq{
				App:             &entity.App{ID: "01K6DB", Key: "my_db"},
				RoutingSettings: &entity.AppRoutingSettings{Port: 5432},
			},
		}
		domain := &entity.AppDomain{
			Enabled:  true,
			Domain:   "db.myapp.com",
			Protocol: base.NetworkProtocolTCP,
		}
		labels := map[string]string{}
		traefikConfig := &AppTraefikConfig{}

		err := s.collectDomainConfig(domain, 0, labels, traefikConfig, data)
		assert.NoError(t, err)

		assert.Equal(t, "tcp-01k6db@file", labels["traefik.tcp.routers.tcp-router-01k6db-0.tls.options"])
		assert.NotContains(t, labels, "traefik.tcp.routers.tcp-router-01k6db-0.tls.passthrough")
		assert.True(t, data.hasFileConfig)
		assert.Contains(t, traefikConfig.TLS.Options["tcp-01k6db"].ALPNProtocols, "postgresql")
		assert.Contains(t, traefikConfig.TLS.Options["tcp-01k6db"].ALPNProtocols, "http/1.1")
	})

	t.Run("TCP domain with extra ALPN protocols gets options of its own", func(t *testing.T) {
		data := &appConfigData{
			ApplyAppConfigReq: &traefikservice.ApplyAppConfigReq{
				App:             &entity.App{ID: "01K6BROKER", Key: "my_broker"},
				RoutingSettings: &entity.AppRoutingSettings{Port: 8883},
			},
		}
		labels := map[string]string{}
		traefikConfig := &AppTraefikConfig{}

		for i, domain := range []*entity.AppDomain{
			{Enabled: true, Domain: "mqtt.myapp.com", Protocol: base.NetworkProtocolTCP},
			{Enabled: true, Domain: "iot.myapp.com", Protocol: base.NetworkProtocolTCP,
				ExtraALPNProtocols: []string{"x-amzn-mqtt-ca", "mqtt"}},
		} {
			assert.NoError(t, s.collectDomainConfig(domain, i, labels, traefikConfig, data))
		}

		assert.Equal(t, "tcp-01k6broker@file", labels["traefik.tcp.routers.tcp-router-01k6broker-0.tls.options"])
		assert.Equal(t, "tcp-01k6broker-1@file", labels["traefik.tcp.routers.tcp-router-01k6broker-1.tls.options"])
		assert.Equal(t, tcpTLSALPNProtocols, traefikConfig.TLS.Options["tcp-01k6broker"].ALPNProtocols)

		own := traefikConfig.TLS.Options["tcp-01k6broker-1"].ALPNProtocols
		assert.Equal(t, []string{"x-amzn-mqtt-ca", "mqtt"}, own[:2], "the domain's own come first")
		assert.Len(t, own, len(tcpTLSALPNProtocols)+1, "the usual ones too, without repeating mqtt")
	})

	// Without Force HTTPS a domain answers HTTP as well, rather than nothing: a
	// router of its own on web, with the main router's rule, service and
	// middlewares and no TLS - and so does each of its paths.
	t.Run("HTTP domain without Force HTTPS is served on HTTP too", func(t *testing.T) {
		data := &appConfigData{
			ApplyAppConfigReq: &traefikservice.ApplyAppConfigReq{
				App:             &entity.App{ID: "01K6WEB", Key: "my_web"},
				RoutingSettings: &entity.AppRoutingSettings{Port: 80},
				RefObjects:      entity.NewRefObjects(),
			},
		}
		domain := &entity.AppDomain{
			Enabled:  true,
			Domain:   "web.myapp.com",
			Protocol: base.NetworkProtocolHTTP,
			HeaderConfig: &entity.HTTPHeaderConfig{
				Enabled: true, ToAddToResponses: map[string]string{"X-Served": "yes"},
			},
			Paths: []*entity.HTTPPathConfig{{Enabled: true, Path: "/api", Mode: base.HTTPPathModePrefix}},
		}
		labels := map[string]string{}
		assert.NoError(t, s.collectDomainConfig(domain, 0, labels, &AppTraefikConfig{}, data))

		main, plain := "traefik.http.routers.router-01k6web-0", "traefik.http.routers.router-01k6web-0-http"
		assert.Equal(t, "web", labels[plain+".entrypoints"])
		assert.Equal(t, labels[main+".rule"], labels[plain+".rule"])
		assert.Equal(t, labels[main+".service"], labels[plain+".service"])
		assert.NotEmpty(t, labels[main+".middlewares"])
		assert.Equal(t, labels[main+".middlewares"], labels[plain+".middlewares"])
		assert.NotContains(t, labels, plain+".tls")
		assert.NotContains(t, labels, "traefik.http.routers.router-01k6web-0-forcehttps.rule")
		// Below the ACME challenge's router on web, whatever the domain's length,
		// and below its own paths'.
		assert.Equal(t, "1", labels[plain+".priority"])

		mainPath, plainPath := main+"-path-0", main+"-path-0-http"
		assert.Equal(t, "web", labels[plainPath+".entrypoints"])
		assert.Equal(t, labels[mainPath+".rule"], labels[plainPath+".rule"])
		assert.Equal(t, labels[mainPath+".service"], labels[plainPath+".service"])
		assert.Equal(t, labels[mainPath+".middlewares"], labels[plainPath+".middlewares"])
		assert.NotContains(t, labels, plainPath+".tls")
		assert.NotContains(t, labels, plainPath+".priority", "a path's ranks by its rule, above its domain's")
	})

	t.Run("HTTP domain with Force HTTPS is only sent to HTTPS on HTTP", func(t *testing.T) {
		data := &appConfigData{
			ApplyAppConfigReq: &traefikservice.ApplyAppConfigReq{
				App:             &entity.App{ID: "01K6WEB", Key: "my_web"},
				RoutingSettings: &entity.AppRoutingSettings{Port: 80},
				RefObjects:      entity.NewRefObjects(),
			},
		}
		domain := &entity.AppDomain{
			Enabled: true, Domain: "web.myapp.com", Protocol: base.NetworkProtocolHTTP, ForceHttps: true,
			Paths: []*entity.HTTPPathConfig{{Enabled: true, Path: "/api", Mode: base.HTTPPathModePrefix}},
		}
		labels := map[string]string{}
		assert.NoError(t, s.collectDomainConfig(domain, 0, labels, &AppTraefikConfig{}, data))

		assert.Equal(t, "web", labels["traefik.http.routers.router-01k6web-0-forcehttps.entrypoints"])
		assert.Equal(t, "1", labels["traefik.http.routers.router-01k6web-0-forcehttps.priority"],
			"below the ACME challenge's router: a long domain's Host rule would outrank its PathPrefix")
		assert.NotContains(t, labels, "traefik.http.routers.router-01k6web-0-http.rule")
		assert.NotContains(t, labels, "traefik.http.routers.router-01k6web-0-path-0-http.rule")
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
