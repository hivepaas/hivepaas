package containerexecserviceimpl

import (
	"context"
	"errors"
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
