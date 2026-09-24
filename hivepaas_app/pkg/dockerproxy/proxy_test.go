package dockerproxy

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefusesEveryEndpointOutsideTheTable(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodPost, "/v1.51/build"},
		{http.MethodPost, "/v1.51/session"},
		{http.MethodGet, "/v1.51/events"},
		{http.MethodPost, "/v1.51/containers/child1/update"},
		{http.MethodPost, "/v1.51/containers/prune"},
		{http.MethodGet, "/v1.51/containers/child1/export"},
		{http.MethodPost, "/v1.51/commit"},
		{http.MethodPost, "/v1.51/swarm/init"},
		{http.MethodGet, "/v1.51/services"},
		{http.MethodGet, "/v1.51/secrets"},
		{http.MethodPost, "/v1.51/plugins/pull"},
		{http.MethodGet, "/v1.51/system/df"},
		{http.MethodPost, "/v1.51/images/load"},
	} {
		status, raw := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusForbidden, status, endpoint.path)
		assert.Contains(t, refusalMessage(t, raw), "is not an endpoint this app may use", endpoint.path)
	}
	assert.Empty(t, w.daemon.requests)
}

func TestRefusesAPathThatIsNotInPlainForm(t *testing.T) {
	// A path the proxy reads one way and the daemon another would be judged by the
	// wrong rule.
	w := newWorld(t, testPolicy())
	for _, path := range []string{
		"/v1.51/containers/..%2Fbuild/json",
		"/v1.51/containers//json",
		"/v1.51/containers/child1/../../build",
	} {
		status, raw := w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusForbidden, status, path)
		assert.Contains(t, refusalMessage(t, raw), "is not in plain form", path)
	}
}

func TestPassesPingVersionAndImageReads(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodGet, "/_ping"},
		{http.MethodHead, "/_ping"},
		{http.MethodGet, "/v1.51/version"},
		{http.MethodGet, "/v1.51/images/json"},
		{http.MethodGet, "/v1.51/images/autobase/automation:2.11.0/json"},
	} {
		status, _ := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusOK, status, endpoint.path)
	}
}

func TestInfoLeavesOutTheCluster(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodGet, "/v1.51/info", nil)
	stop(t, assert.Equal(t, http.StatusOK, status))
	var info map[string]any
	stop(t, assert.NoError(t, json.Unmarshal(raw, &info)))
	assert.Equal(t, "29.8.0", info["ServerVersion"])
	assert.NotContains(t, info, "Swarm")
	assert.NotContains(t, info, "Labels")
	assert.NotContains(t, info, "RegistryConfig")
}

func TestPullTakesOnlyThePolicysImages(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, _ := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=alpine&tag=3", nil)
	assert.Equal(t, http.StatusOK, status)
	path, _ := w.forwarded(t, http.MethodPost, "/images/create")
	assert.Equal(t, "/v1.51/images/create", path)

	status, raw := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: image busybox:1 is not allowed", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/images/create?fromSrc=http://example.com/rootfs.tar", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: importing an image is not allowed", refusalMessage(t, raw))
}

func TestDecisionsAreReported(t *testing.T) {
	w := newWorld(t, testPolicy())
	w.do(t, http.MethodGet, "/_ping", nil)
	w.do(t, http.MethodPost, "/v1.51/build", nil)

	w.mu.Lock()
	defer w.mu.Unlock()
	stop(t, assert.Len(t, w.decisions, 2))
	assert.Equal(t, Decision{AppID: "app1", Method: http.MethodGet, Path: "/_ping", Allowed: true,
		Reason: "read-only endpoint"}, w.decisions[0])
	assert.False(t, w.decisions[1].Allowed)
	assert.Equal(t, "POST /build is not an endpoint this app may use", w.decisions[1].Reason)
}

func TestSetPolicyAppliesToTheNextRequest(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, _ := w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusForbidden, status)

	policy := testPolicy()
	policy.Images = []string{"busybox"}
	w.proxy.SetPolicy(policy)
	status, _ = w.do(t, http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1", nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestAnUnreachableDaemonIsNotARefusal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	stop(t, assert.NoError(t, err))
	addr := ln.Addr().String()
	stop(t, assert.NoError(t, ln.Close()))
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}
	server := httptest.NewServer(New(testPolicy(), Options{Upstream: upstream}))
	defer server.Close()

	w := &world{url: server.URL}
	status, raw := w.do(t, http.MethodGet, "/_ping", nil)
	assert.Equal(t, http.StatusBadGateway, status)
	assert.Contains(t, refusalMessage(t, raw), "hivepaas: the Docker daemon did not answer")
}
