package dockerapiagentuc

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TestAgainstARealDaemon runs a child through an app's socket against the
// Docker daemon of this machine, and sweeps it away. It needs a daemon, so it
// runs only when HP_TEST_DOCKER is set, and needs alpine:3 pullable.
func TestAgainstARealDaemon(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER") == "" {
		t.Skip("set HP_TEST_DOCKER=1 to run against this machine's Docker daemon")
	}
	ctx := context.Background()
	dockerManager, err := docker.New()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	upstream, err := newUpstream()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	id, err := ulid.NewStringULID()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	appID := "itest-" + id
	networkName := "hp-dapi-" + appID
	_, err = dockerManager.NetworkCreate(ctx, networkName, func(opts *client.NetworkCreateOptions) {
		opts.Driver = "bridge"
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Cleanup(func() { _, _ = dockerManager.NetworkRemove(context.Background(), networkName) })

	dir := shortTempDir(t)
	host := newSocketHost(logging.GlobalLogger(), dockerManager, upstream)
	host.socketDir = func(context.Context, *dockerproxy.Policy) (string, error) { return dir, nil }
	t.Cleanup(host.closeAll)
	policy := &dockerproxy.Policy{AppID: appID, Images: []string{"alpine"}, Network: networkName,
		Limits: dockerproxy.Limits{Containers: 2, Memory: 64 << 20, NanoCPUs: 500_000_000}}
	if !assert.NoError(t, host.reconcile(ctx, []*dockerproxy.Policy{policy})) {
		t.FailNow()
	}

	// The app's side: a Docker client that knows nothing of the proxy.
	app, err := client.New(client.WithHost("unix://" + filepath.Join(dir, dockerproxy.SocketFile)))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	pull, err := app.ImagePull(ctx, "alpine:3", client.ImagePullOptions{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, _ = io.Copy(io.Discard, pull)
	_ = pull.Close()

	created, err := app.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{Image: "alpine:3", Cmd: []string{"true"}},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = app.ContainerStart(ctx, created.ID, client.ContainerStartOptions{})
	assert.NoError(t, err)
	wait := app.ContainerWait(ctx, created.ID, client.ContainerWaitOptions{})
	select {
	case <-wait.Result:
	case err = <-wait.Error:
		assert.NoError(t, err)
	case <-time.After(time.Minute):
		t.Fatal("the child did not finish")
	}

	inspect, err := dockerManager.ContainerInspect(ctx, created.ID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, appID, inspect.Container.Config.Labels[dockerproxy.OwnerLabel])
	assert.Equal(t, networkName, string(inspect.Container.HostConfig.NetworkMode))
	assert.Equal(t, int64(64<<20), inspect.Container.HostConfig.Memory)

	_, err = app.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     &container.Config{Image: "alpine:3"},
		HostConfig: &container.HostConfig{Privileged: true},
	})
	assert.ErrorContains(t, err, "HostConfig.Privileged is not allowed")

	s := &sweep{docker: dockerManager, now: time.Now(), only: appID,
		gone: func(id string) bool { return id == appID }}
	removed, err := s.run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, removed.Containers)
	_, err = dockerManager.ContainerInspect(ctx, created.ID)
	assert.Error(t, err)
}
