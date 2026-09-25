package dockerproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// world is a proxy in front of a fake daemon holding:
//   - task1, the app's own task (service svc1), with its volume hp-vol-data
//     mounted at /var/lib/autobase under subpath app1-key;
//   - child1, a container the app started, with exec exec-child;
//   - other1, a container of another app, with exec exec-other;
//   - volumes cache-app1 (the app's), cache-app2 (another app's) and the app's
//     socket volume;
//   - networks hp-dapi-app1 (the app's, n-app), jobnet (created by the app,
//     n-job), othernet (another app's), proj_env_net (the app's env) and
//     hivepaas_net.
type world struct {
	daemon *fakeDaemon
	proxy  *Proxy
	url    string

	mu        sync.Mutex
	decisions []Decision
}

func testPolicy() *Policy {
	return &Policy{
		AppID:          "app1",
		ServiceID:      "svc1",
		Images:         []string{"alpine", "autobase/automation:2.11.0"},
		SharedDirs:     []string{"/var/lib/autobase/ansible"},
		Network:        "hp-dapi-app1",
		Networks:       []string{"proj_env_net"},
		SocketVolume:   "hp-dapi-sock-app1",
		ReservedPrefix: "hp-dapi-",
		Limits:         Limits{Containers: 3, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
	}
}

func newWorld(t testing.TB, policy *Policy) *world {
	t.Helper()
	fake := &fakeDaemon{
		containers: map[string]*fakeContainer{
			"task1": {labels: map[string]string{"com.docker.swarm.service.id": "svc1"}, running: true,
				mounts: []map[string]any{{
					"Type": "volume", "Source": "hp-vol-data", "Target": "/var/lib/autobase",
					"VolumeOptions": map[string]any{"Subpath": "app1-key"},
				}}},
			"child1": {labels: map[string]string{OwnerLabel: "app1"}, running: true},
			"other1": {labels: map[string]string{OwnerLabel: "app2"}, running: true},
		},
		execs: map[string]string{"exec-child": "child1", "exec-other": "other1"},
		volumes: map[string]map[string]string{
			"cache-app1": {OwnerLabel: "app1"}, "cache-app2": {OwnerLabel: "app2"}, "hp-dapi-sock-app1": {},
		},
		networks: map[string]*fakeNetwork{
			"n-app":   {name: "hp-dapi-app1", labels: map[string]string{}},
			"n-job":   {name: "jobnet", labels: map[string]string{OwnerLabel: "app1"}},
			"n-other": {name: "othernet", labels: map[string]string{OwnerLabel: "app2"}},
			"n-env":   {name: "proj_env_net", labels: map[string]string{}},
			"n-hp":    {name: "hivepaas_net", labels: map[string]string{}},
		},
	}
	daemonServer := httptest.NewServer(fake)
	t.Cleanup(daemonServer.Close)
	addr := daemonServer.Listener.Addr().String()
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}

	w := &world{daemon: fake}
	w.proxy = New(policy, Options{Upstream: upstream, OnDecision: func(d Decision) {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.decisions = append(w.decisions, d)
	}})
	proxyServer := httptest.NewServer(w.proxy)
	t.Cleanup(proxyServer.Close)
	w.url = proxyServer.URL
	return w
}

// do sends a request to the proxy. body is sent as it is when it is bytes, and
// as JSON otherwise.
func (w *world) do(t testing.TB, method, path string, body any) (int, []byte) {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case []byte:
		reader = bytes.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		stop(t, assert.NoError(t, err))
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, w.url+path, reader)
	stop(t, assert.NoError(t, err))
	resp, err := http.DefaultClient.Do(req)
	stop(t, assert.NoError(t, err))
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	stop(t, assert.NoError(t, err))
	return resp.StatusCode, raw
}

// refusalMessage is the message of an answer the proxy wrote itself.
func refusalMessage(t testing.TB, raw []byte) string {
	t.Helper()
	var answer struct {
		Message string `json:"message"`
	}
	stop(t, assert.NoError(t, json.Unmarshal(raw, &answer), string(raw)))
	return answer.Message
}

// reached reports whether a request with that method and path suffix reached
// the daemon.
func (w *world) reached(method, suffix string) bool {
	w.daemon.mu.Lock()
	defer w.daemon.mu.Unlock()
	for _, req := range w.daemon.requests {
		if req.method == method && strings.HasSuffix(req.path, suffix) {
			return true
		}
	}
	return false
}

// posted returns the path and decoded body of the last POST with that path
// suffix that reached the daemon.
func (w *world) posted(t testing.TB, suffix string) (string, map[string]any) {
	t.Helper()
	w.daemon.mu.Lock()
	defer w.daemon.mu.Unlock()
	for i := len(w.daemon.requests) - 1; i >= 0; i-- {
		req := w.daemon.requests[i]
		if req.method != http.MethodPost || !strings.HasSuffix(req.path, suffix) {
			continue
		}
		body := map[string]any{}
		if len(req.body) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(req.body))
			decoder.UseNumber()
			stop(t, assert.NoError(t, decoder.Decode(&body)))
		}
		return req.path, body
	}
	t.Fatalf("no POST ...%s reached the daemon", suffix)
	return "", nil
}

// fixture decodes a request body recorded from a real client.
func fixture(t testing.TB, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	stop(t, assert.NoError(t, err))
	body := map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	stop(t, assert.NoError(t, decoder.Decode(&body)))
	return body
}

// set changes a field of a decoded body by its dotted path, and returns the body.
func set(body map[string]any, dotted string, value any) map[string]any {
	keys := strings.Split(dotted, ".")
	obj := body
	for _, key := range keys[:len(keys)-1] {
		next, ok := obj[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			obj[key] = next
		}
		obj = next
	}
	obj[keys[len(keys)-1]] = value
	return body
}
