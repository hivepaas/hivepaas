package traefikserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDockerManager embeds the interface so only the one call this file makes
// needs a body; anything else would panic rather than quietly pass.
type fakeDockerManager struct {
	docker.Manager
	traefikSvc *swarm.Service
}

func (f *fakeDockerManager) ServiceGetByName(
	_ context.Context, _ string, _ bool,
) (*swarm.Service, error) {
	return f.traefikSvc, nil
}

func traefikSvcWithTrustedIPs(trusted bool) *swarm.Service {
	args := []string{"--entrypoints.web.address=:80"}
	if trusted {
		args = append(args, "--entrypoints.web.forwardedheaders.trustedips=173.245.48.0/20")
	}
	return &swarm.Service{Spec: swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Args: args}},
	}}
}

func configDataFor(trusted bool, hops int) *appConfigData {
	return &appConfigData{
		traefikSvc: traefikSvcWithTrustedIPs(trusted),
		proxySettings: &entity.HivePaaSProxySettings{
			ProxyProvider: "cloudflare",
			TrustedIPs:    []string{"173.245.48.0/20"},
			ProxyHops:     hops,
		},
	}
}

func TestClientIPStrategyDepth(t *testing.T) {
	svc := &service{}

	t.Run("uses the configured hops when the entrypoint trusts the proxy", func(t *testing.T) {
		assert.Equal(t, 3, svc.clientIPStrategyDepth(configDataFor(true, 3)))
	})

	// Without trusted IPs Traefik strips the forwarded headers before any
	// middleware runs, so a depth would be read off a header the caller wrote.
	t.Run("ignores the hops when the entrypoint trusts nobody", func(t *testing.T) {
		assert.Zero(t, svc.clientIPStrategyDepth(configDataFor(false, 3)))
	})

	t.Run("no proxy declared means the peer address", func(t *testing.T) {
		data := &appConfigData{traefikSvc: traefikSvcWithTrustedIPs(true)}
		assert.Zero(t, svc.clientIPStrategyDepth(data))
	})

	t.Run("a declared proxy with no hops is not guessed at", func(t *testing.T) {
		assert.Zero(t, svc.clientIPStrategyDepth(configDataFor(true, 0)))
	})
}

func TestCreateRateLimitConfigAppliesTheSameSourceToBothMiddlewares(t *testing.T) {
	svc := &service{}
	labels := map[string]string{}
	var middlewares []string

	svc.createRateLimitConfig(&entity.HTTPRateLimitConfig{
		Enabled: true, Average: 20, Burst: 30, MaxInFlightReq: 10,
	}, "router", labels, &middlewares, configDataFor(true, 2))

	// Two middlewares, two independent criteria: configuring one and not the other
	// counts requests per caller while counting concurrency per proxy.
	assert.Equal(t, "2", labels["traefik.http.middlewares.router-ratelimit.ratelimit.sourcecriterion.ipstrategy.depth"])
	assert.Equal(t, "2",
		labels["traefik.http.middlewares.router-inflightreq.inflightreq.sourcecriterion.ipstrategy.depth"])
}

// A criterion that reads a header nothing guarantees is worse than none: every
// request would land in a bucket of its own.
func TestCreateRateLimitConfigWritesNoCriterionWithoutTrustedIPs(t *testing.T) {
	svc := &service{}
	labels := map[string]string{}
	var middlewares []string

	svc.createRateLimitConfig(&entity.HTTPRateLimitConfig{
		Enabled: true, Average: 20, MaxInFlightReq: 10,
	}, "router", labels, &middlewares, configDataFor(false, 2))

	for key := range labels {
		assert.NotContains(t, key, "sourcecriterion")
	}
	assert.Equal(t, "20", labels["traefik.http.middlewares.router-ratelimit.ratelimit.average"])
}

// Traefik refuses a criterion that sets both, and takes the router down with it.
func TestCreateRateLimitConfigNeverWritesARequestHeaderName(t *testing.T) {
	svc := &service{}
	labels := map[string]string{}
	var middlewares []string

	svc.createRateLimitConfig(&entity.HTTPRateLimitConfig{
		Enabled: true, Average: 20, MaxInFlightReq: 10,
	}, "router", labels, &middlewares, configDataFor(true, 2))

	for key := range labels {
		assert.NotContains(t, key, "requestheadername")
	}
}

// The allowlist and the rate limits must agree on who the caller is; the
// allowlist is the one where disagreeing decides who gets in.
func TestCreateClientConfigUsesTheSameDepth(t *testing.T) {
	svc := &service{}
	labels := map[string]string{}
	var middlewares []string
	data := configDataFor(true, 3)

	svc.createClientConfig(&entity.HTTPClientConfig{
		Enabled: true, AllowedIPs: []string{"10.0.0.0/8"},
	}, "router", labels, &middlewares, data)

	assert.Equal(t, "3", labels["traefik.http.middlewares.router-allowed-ips.ipallowlist.ipstrategy.depth"])
}

func TestNeedsClientIPStrategy(t *testing.T) {
	t.Run("nothing asks who the caller is", func(t *testing.T) {
		assert.False(t, needsClientIPStrategy(&entity.AppRoutingSettings{
			ExposePublicly: true,
			Domains:        []*entity.AppDomain{{Domain: "x.example.com"}},
		}))
	})

	t.Run("a rate limit on a path asks", func(t *testing.T) {
		assert.True(t, needsClientIPStrategy(&entity.AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*entity.AppDomain{{Paths: []*entity.HTTPPathConfig{{
				RateLimitConfig: &entity.HTTPRateLimitConfig{Enabled: true, Average: 10},
			}}}},
		}))
	})

	t.Run("an allowlist asks", func(t *testing.T) {
		assert.True(t, needsClientIPStrategy(&entity.AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*entity.AppDomain{{
				ClientConfig: &entity.HTTPClientConfig{Enabled: true, AllowedIPs: []string{"10.0.0.0/8"}},
			}},
		}))
	})

	t.Run("a disabled rate limit does not", func(t *testing.T) {
		assert.False(t, needsClientIPStrategy(&entity.AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*entity.AppDomain{{
				RateLimitConfig: &entity.HTTPRateLimitConfig{Enabled: false, Average: 10},
			}},
		}))
	})
}

// Nothing loaded the settings, which means nothing needed them - and the answer
// then has to be the safe one, not a nil dereference.
func TestGetProxySettingsWithoutALoad(t *testing.T) {
	data := &appConfigData{}
	assert.Zero(t, data.getProxySettings().ClientIPDepth())
}
