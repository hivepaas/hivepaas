package traefikservice

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// An app's service pattern matches its own domains' services exactly: not
// another app's whose id starts the same, not a redirect router's.
func TestAppHTTPServicePatternIsExact(t *testing.T) {
	re := regexp.MustCompile(AppHTTPServicePattern("01K6A"))
	assert.True(t, re.MatchString("svc-01k6a-0@swarm"))
	assert.True(t, re.MatchString("svc-01k6a-12@swarm"))
	for _, other := range []string{"svc-01k6ab-0@swarm", "svc-01k6a-0-x@swarm", "svc-01k6a-0@file",
		"xsvc-01k6a-0@swarm", "svc-01k6a-@swarm"} {
		assert.False(t, re.MatchString(other), other)
	}
	assert.Equal(t, "svc-01k6a-3", AppHTTPServiceName("01K6A", 3))
}

// Labels written by key are stale; those by id are not; an operator's
// x-custom- labels and labels that name nothing are left alone.
func TestHasStaleRouteNames(t *testing.T) {
	byID := map[string]string{
		"traefik.enable": "true",
		"traefik.http.routers.router-01k6a-0.rule":                                 "Host(`a.test`)",
		"traefik.http.routers.router-01k6a-0-path-1.rule":                          "Host(`a.test`) && Path(`/x`)",
		"traefik.http.services.svc-01k6a-0.loadbalancer.server.port":               "8080",
		"traefik.tcp.routers.tcp-router-01k6a-1.rule":                              "HostSNI(`b.test`)",
		"traefik.http.middlewares.router-01k6a-0-forcehttps.redirectscheme.scheme": "https",
		"traefik.http.routers.x-custom-acme.rule":                                  "Host(`c.test`)",
	}
	assert.False(t, HasStaleRouteNames(byID, "01K6A"))

	byKey := map[string]string{"traefik.http.routers.router-app-0.rule": "Host(`a.test`)"}
	assert.True(t, HasStaleRouteNames(byKey, "01K6A"))
	assert.False(t, HasStaleRouteNames(map[string]string{"traefik.enable": "true"}, "01K6A"))
}
