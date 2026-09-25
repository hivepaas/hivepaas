package dockerapiagentuc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeVolumes creates socket volumes in a temporary directory, as a node's
// daemon would under its data root.
type fakeVolumes struct {
	docker.Manager
	mu      sync.Mutex
	root    string
	created map[string]map[string]string
}

func (f *fakeVolumes) VolumeCreate(_ context.Context, options ...docker.VolumeCreateOption) (
	*client.VolumeCreateResult, error) {
	opts := client.VolumeCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	// Docker returns a volume that exists as it is, labels and all, rather than
	// creating it again.
	labels, exists := f.created[opts.Name]
	if !exists {
		labels = opts.Labels
		f.created[opts.Name] = labels
	}
	dir := filepath.Join(f.root, opts.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &client.VolumeCreateResult{
		Volume: volume.Volume{Name: opts.Name, Mountpoint: dir, Labels: labels},
	}, nil
}

// shortTempDir is a directory with a path short enough for a unix socket, which
// macOS limits to 104 bytes and t.TempDir can exceed.
func shortTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dapi")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// newTestHost is a host over fake volumes, in front of a daemon that answers
// every request with 200 and counts them.
func newTestHost(t *testing.T) (*socketHost, *fakeVolumes, *int) {
	t.Helper()
	requests := 0
	var mu sync.Mutex
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		requests++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(daemon.Close)
	addr := daemon.Listener.Addr().String()
	upstream := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}}

	volumes := &fakeVolumes{root: shortTempDir(t), created: map[string]map[string]string{}}
	host := newSocketHost(logging.GlobalLogger(), volumes, upstream)
	t.Cleanup(host.closeAll)
	return host, volumes, &requests
}

func testPolicy(appID string) *dockerproxy.Policy {
	return &dockerproxy.Policy{
		AppID: appID, Images: []string{"alpine"}, Network: dockerapiservice.NetworkName(appID),
		SocketVolume: dockerapiservice.SocketVolumeName(appID),
		Limits:       dockerproxy.Limits{Containers: 1, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
	}
}

func socketOf(volumes *fakeVolumes, appID string) string {
	return filepath.Join(volumes.root, dockerapiservice.SocketVolumeName(appID), dockerproxy.SocketFile)
}

// send makes a request over an app's socket.
func send(t *testing.T, socket, method, path string) (int, error) {
	t.Helper()
	httpClient := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}}
	req, err := http.NewRequestWithContext(context.Background(), method, "http://docker"+path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	_ = resp.Body.Close()
	return resp.StatusCode, nil
}

func TestReconcileServesEachAppOnASocketInItsVolume(t *testing.T) {
	host, volumes, requests := newTestHost(t)
	err := host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("app1"), testPolicy("app2")})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []string{"app1", "app2"}, host.served())
	assert.Equal(t, map[string]string{dockerapiservice.SocketVolumeLabel: "app1"},
		volumes.created[dockerapiservice.SocketVolumeName("app1")])

	for _, appID := range []string{"app1", "app2"} {
		status, err := send(t, socketOf(volumes, appID), http.MethodPost, "/v1.51/images/create?fromImage=alpine&tag=3")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, 2, *requests)
}

func TestReconcileGivesAChangedPolicyToTheOpenSocket(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	ctx := context.Background()
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{testPolicy("app1")}))
	status, _ := send(t, socketOf(volumes, "app1"), http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1")
	assert.Equal(t, http.StatusForbidden, status)

	changed := testPolicy("app1")
	changed.Images = []string{"busybox"}
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{changed}))
	status, _ = send(t, socketOf(volumes, "app1"), http.MethodPost, "/v1.51/images/create?fromImage=busybox&tag=1")
	assert.Equal(t, http.StatusOK, status)
}

func TestReconcileClosesTheSocketOfAnAppNoLongerListed(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	ctx := context.Background()
	assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{testPolicy("app1")}))
	assert.NoError(t, host.reconcile(ctx, nil))

	assert.Empty(t, host.served())
	_, err := os.Stat(socketOf(volumes, "app1"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = send(t, socketOf(volumes, "app1"), http.MethodGet, "/_ping")
	assert.Error(t, err)
}

func TestReconcileReplacesASocketAnEarlierAgentLeft(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	socket := socketOf(volumes, "app1")
	if !assert.NoError(t, os.MkdirAll(filepath.Dir(socket), 0o755)) {
		t.FailNow()
	}
	assert.NoError(t, os.WriteFile(socket, []byte("stale"), 0o600))

	assert.NoError(t, host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("app1")}))
	status, err := send(t, socket, http.MethodGet, "/_ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
}

func TestOneAppThatCannotBeServedDoesNotStopTheOthers(t *testing.T) {
	host, volumes, _ := newTestHost(t)
	unreachable := errors.New("socket volume unreachable")
	serveInVolume := host.socketDir
	host.socketDir = func(ctx context.Context, policy *dockerproxy.Policy) (string, error) {
		if policy.AppID == "broken" {
			return "", unreachable
		}
		return serveInVolume(ctx, policy)
	}

	err := host.reconcile(context.Background(), []*dockerproxy.Policy{testPolicy("broken"), testPolicy("app1")})
	assert.ErrorIs(t, err, unreachable)
	assert.Equal(t, []string{"app1"}, host.served())
	status, _ := send(t, socketOf(volumes, "app1"), http.MethodGet, "/_ping")
	assert.Equal(t, http.StatusOK, status)
}

func TestCloseAppStopsServingOneApp(t *testing.T) {
	host, _, _ := newTestHost(t)
	assert.NoError(t, host.reconcile(context.Background(),
		[]*dockerproxy.Policy{testPolicy("app1"), testPolicy("app2")}))
	host.closeApp("app1")
	assert.Equal(t, []string{"app2"}, host.served())
}

// A socket volume under an app's name that something else made is not the
// app's: putting its socket there would leave it where that something else can
// reach it, which is the Docker API of the whole app.
func TestReconcileRefusesASocketVolumeLabeledForAnotherApp(t *testing.T) {
	root := shortTempDir(t)
	volumes := &fakeVolumes{root: root, created: map[string]map[string]string{
		"hp-dapi-sock-app1": {dockerapiservice.SocketVolumeLabel: "app2"},
	}}

	_, err := volumeSocketDir(context.Background(), volumes, &dockerproxy.Policy{
		AppID: "app1", SocketVolume: "hp-dapi-sock-app1",
	})

	assert.Error(t, err)
}
