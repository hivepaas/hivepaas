package containerexecserviceimpl

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeExecDocker struct {
	docker.Manager
	execErr error
}

func (f *fakeExecDocker) NodeCurrentID(context.Context) (string, error) { return "node-1", nil }

func (f *fakeExecDocker) ContainerExec(
	context.Context, string, ...docker.ExecCreateOption,
) (*client.ExecCreateResult, *client.ExecAttachResult, *client.ExecStartResult, error) {
	return nil, nil, nil, f.execErr
}

func (f *fakeExecDocker) CanRetryExec(context.Context, string) (bool, error) { return false, nil }

// An exec that fails to start says why: the error is its own, not a panic
// raised while cleaning up after it.
func TestContainerExecReportsTheFailureItself(t *testing.T) {
	previous := config.Current()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(previous) })

	execErr := errors.New("no such container")
	svc := &service{dockerManager: &fakeExecDocker{execErr: execErr}}

	_, err := svc.ContainerExec(context.Background(), &containerexecservice.ContainerExecReq{
		ContainerID: "c1",
		NodeID:      "node-1",
		ExecOptions: func(*client.ExecCreateOptions) {},
	})

	assert.ErrorIs(t, err, execErr)
	assert.NotErrorIs(t, err, hperrors.ErrPanic)
}

// A stdin read from is not read again from its start: the guard says so once
// a byte of it is gone, which keeps a failed exec from being retried with what
// is left of it.
func TestStdinGuardTellsAReadStdin(t *testing.T) {
	guard := &stdinGuard{r: strings.NewReader("rows")}
	assert.False(t, guard.read())
	buf := make([]byte, 2)
	n, err := guard.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, guard.read())
}
