package settingmountserviceimpl

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TestAgainstARealDaemon mounts a certificate and its key outside /run/secrets,
// renews them, and checks the container reads the new pair after one update and
// that the old objects are gone. It needs a swarm manager and alpine:3, so it
// runs only when HP_TEST_DOCKER is set.
func TestAgainstARealDaemon(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER") == "" {
		t.Skip("set HP_TEST_DOCKER=1 to run against this machine's Docker daemon")
	}
	ctx := context.Background()
	dockerManager, err := docker.New()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	id, err := ulid.NewStringULID()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	app := &entity.App{ID: "itest-" + id, GlobalKey: "itest_" + id}

	source := certSource(t, "cert_1", "CERT-1", "KEY-1")
	svc := fixture(t, []*entity.Setting{entry(t, "tls-cert", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"}, Files: []*entity.AppSettingMountFile{
			{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
			{Part: "privateKey", Path: "/etc/app/tls/key.pem", Mode: 0o400},
		}})}, source)
	svc.dockerManager, svc.logger = dockerManager, logging.GlobalLogger()
	svc.removalRetryDelay = time.Second

	spec := swarm.ServiceSpec{Annotations: swarm.Annotations{Name: app.GlobalKey},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
			Image: "alpine:3", Command: []string{"sleep", "3600"}}}}
	assert.NoError(t, svc.ApplyToService(ctx, nil, app, &spec))
	created, err := dockerManager.ServiceCreate(ctx, &spec)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	app.ServiceID = created.ID
	t.Cleanup(func() {
		_, _ = dockerManager.ServiceRemove(context.Background(), app.ServiceID)
		time.Sleep(3 * time.Second)
		_ = svc.RemoveApp(context.Background(), app.ID)
	})

	assert.Equal(t, "KEY-1", readFile(t, dockerManager, app.ServiceID, "/etc/app/tls/key.pem", "KEY-1"))

	assert.NoError(t, source.SetData(&entity.SSLCert{Certificate: "CERT-2",
		PrivateKey: entity.NewEncryptedField("KEY-2")}))
	assert.NoError(t, svc.Refresh(ctx, nil, app))
	assert.Equal(t, "KEY-2", readFile(t, dockerManager, app.ServiceID, "/etc/app/tls/key.pem", "KEY-2"))
	assert.Equal(t, "CERT-2", readFile(t, dockerManager, app.ServiceID, "/etc/app/tls/cert.pem", "CERT-2"))

	// The old task has stopped by now, and with it the last hold on the old pair.
	assert.NoError(t, svc.Sweep(ctx, app))
	left, err := dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		docker.FilterAdd(&opts.Filters, "label", "hivepaas.app.id="+app.ID)
	})
	assert.NoError(t, err)
	assert.Len(t, left.Items, 1, "only the new key is left")
}

// readFile reads path from the service's running container, waiting up to a
// minute for one that has want: a rolling update replaces the container. It
// runs cat: a secret lives on a tmpfs the archive endpoint does not see.
func readFile(t *testing.T, dockerManager docker.Manager, serviceID, path, want string) string {
	t.Helper()
	ctx := context.Background()
	var got string
	for deadline := time.Now().Add(time.Minute); time.Now().Before(deadline); time.Sleep(time.Second) {
		containers, err := dockerManager.ServiceContainerList(ctx, serviceID)
		if err != nil {
			continue
		}
		for _, c := range containers.Items {
			if c.State != "running" {
				continue
			}
			_, frames, err := dockerManager.ContainerExecWait(ctx, c.ID, func(opts *client.ExecCreateOptions) {
				opts.Cmd = []string{"cat", path}
				// A terminal, so that the output is not framed as stdout and stderr.
				opts.TTY, opts.AttachStdout, opts.AttachStderr = true, true, true
			})
			if err != nil {
				continue
			}
			var out strings.Builder
			for _, frame := range frames {
				out.WriteString(frame.Data)
			}
			got = strings.TrimSpace(out.String())
			if got == want {
				return got
			}
		}
	}
	return got
}
