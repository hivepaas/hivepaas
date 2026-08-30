package docker

import (
	"context"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

var (
	DefaultConsoleSize = client.ConsoleSize{
		Height: 40,  //nolint:mnd
		Width:  120, //nolint:mnd
	}
)

type ExecCreateOption func(*client.ExecCreateOptions)

func (m *manager) ContainerExec(
	ctx context.Context,
	containerID string,
	options ...ExecCreateOption,
) (*client.ExecCreateResult, *client.ExecAttachResult, *client.ExecStartResult, error) {
	opts := client.ExecCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	if !opts.TTY {
		opts.ConsoleSize = client.ConsoleSize{}
	}

	_, err := m.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}

	createResp, err := m.client.ExecCreate(ctx, containerID, opts)
	if err != nil {
		return nil, nil, nil, hperrors.NewInfra(err)
	}
	execID := createResp.ID
	if execID == "" {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrInfraInternal)
	}

	attachResp, err := m.client.ExecAttach(ctx, execID, client.ExecAttachOptions{
		TTY:         opts.TTY,
		ConsoleSize: opts.ConsoleSize,
	})
	if err != nil {
		return &createResp, &attachResp, nil, hperrors.NewInfra(err)
	}

	startResp, err := m.client.ExecStart(ctx, execID, client.ExecStartOptions{
		Detach:      false, // TODO: handle this
		TTY:         opts.TTY,
		ConsoleSize: opts.ConsoleSize,
	})
	if err != nil {
		return &createResp, &attachResp, &startResp, hperrors.NewInfra(err)
	}

	return &createResp, &attachResp, &startResp, nil
}

func (m *manager) ContainerExecWait(
	ctx context.Context,
	containerID string,
	options ...ExecCreateOption,
) (*client.ExecInspectResult, []*tasklog.LogFrame, error) {
	createResp, attachResp, _, err := m.ContainerExec(ctx, containerID, options...)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	logChan, _ := StartScanningLog(ctx, io.NopCloser(attachResp.Reader), WithParseLogHeader(false))
	defer attachResp.Close()

	logs := make([]*tasklog.LogFrame, 0, 20) //nolint:mnd
	for msgs := range logChan {
		logs = append(logs, msgs...)
	}

	inspectResp, err := m.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	return inspectResp, logs, nil
}

func (m *manager) ContainerExecResize(
	ctx context.Context,
	execID string,
	width, height uint,
) (*client.ExecResizeResult, error) {
	resp, err := m.client.ExecResize(ctx, execID, client.ExecResizeOptions{
		Width:  width,
		Height: height,
	})
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ExecInspectOption func(*client.ExecInspectOptions)

func (m *manager) ContainerExecInspect(
	ctx context.Context,
	execID string,
	options ...ExecInspectOption,
) (*client.ExecInspectResult, error) {
	opts := client.ExecInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ExecInspect(ctx, execID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) CanRetryExec(
	ctx context.Context,
	execID string,
) (bool, error) {
	if execID == "" {
		return true, nil
	}

	inspectResp, err := m.ContainerExecInspect(ctx, execID)
	if err != nil {
		if hperrors.IsInfraNotFound(err) {
			return true, nil
		}
		return false, hperrors.Wrap(err)
	}
	if inspectResp == nil {
		return true, nil
	}

	// If the process is currently running or was already started (PID > 0),
	// it should not be retried to prevent duplicate execution.
	if inspectResp.Running || inspectResp.PID > 0 {
		return false, nil
	}

	return true, nil
}

func (m *manager) ContainerCreateToExec(
	ctx context.Context,
	image string,
	cmd []string,
	options ...ContainerCreateOption,
) (createResp *client.ContainerCreateResult, statusCode int64, err error) {
	createOpts := client.ContainerCreateOptions{
		Config: &container.Config{
			Image: image,
			Cmd:   cmd,
		},
		HostConfig: &container.HostConfig{
			AutoRemove: true,
		},
		Name: TempContainerPrefix + gofn.RandString(5), //nolint:mnd
	}
	for _, opt := range options {
		opt(&createOpts)
	}

	if createOpts.HostConfig == nil {
		createOpts.HostConfig = &container.HostConfig{}
	}
	createOpts.HostConfig.AutoRemove = true

	if createOpts.Config == nil {
		createOpts.Config = &container.Config{}
	}
	if createOpts.Config.Labels == nil {
		createOpts.Config.Labels = make(map[string]string)
	}
	createOpts.Config.Labels[LabelTempResource] = LabelTempResourceVal
	createOpts.Config.Labels[LabelTempCreatedAt] = time.Now().UTC().Format(time.RFC3339)

	createResp, err = m.ContainerCreate(ctx, func(opts *client.ContainerCreateOptions) {
		*opts = createOpts
	})
	if err != nil {
		return nil, 0, hperrors.Wrap(err)
	}

	defer func() { //nolint:contextcheck
		if err != nil && createResp != nil && createResp.ID != "" {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
			defer cancel()
			_, _ = m.ContainerRemove(cleanupCtx, createResp.ID, func(opts *client.ContainerRemoveOptions) {
				opts.Force = true
			})
		}
	}()

	_, err = m.ContainerStart(ctx, createResp.ID)
	if err != nil {
		return createResp, 0, hperrors.Wrap(err)
	}

	waitRes := m.ContainerWait(ctx, createResp.ID, func(opts *client.ContainerWaitOptions) {
		opts.Condition = container.WaitConditionNotRunning
	})
	select {
	case waitErr := <-waitRes.Error:
		if waitErr != nil && hperrors.IsInfraNotFound(waitErr) {
			waitErr = nil
		}
		return createResp, 0, hperrors.Wrap(waitErr)
	case waitResp := <-waitRes.Result:
		return createResp, waitResp.StatusCode, nil
	case <-ctx.Done():
		return createResp, 0, hperrors.Wrap(ctx.Err())
	}
}
