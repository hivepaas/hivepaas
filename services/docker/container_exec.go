package docker

import (
	"context"
	"io"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

var (
	DefaultConsoleSize = client.ConsoleSize{
		Height: 40,  //nolint
		Width:  120, //nolint
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

	if opts.ConsoleSize.Width > 0 || opts.ConsoleSize.Height > 0 {
		opts.TTY = true
	}

	_, err := m.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, nil, nil, apperrors.New(err)
	}

	createResp, err := m.client.ExecCreate(ctx, containerID, opts)
	if err != nil {
		return nil, nil, nil, apperrors.NewInfra(err)
	}
	execID := createResp.ID
	if execID == "" {
		return nil, nil, nil, apperrors.New(apperrors.ErrInfraInternal)
	}

	attachResp, err := m.client.ExecAttach(ctx, execID, client.ExecAttachOptions{
		TTY:         opts.TTY,
		ConsoleSize: opts.ConsoleSize,
	})
	if err != nil {
		return nil, nil, nil, apperrors.NewInfra(err)
	}

	startResp, err := m.client.ExecStart(ctx, execID, client.ExecStartOptions{
		Detach:      false, // TODO: handle this
		TTY:         opts.TTY,
		ConsoleSize: opts.ConsoleSize,
	})
	if err != nil {
		return nil, nil, nil, apperrors.NewInfra(err)
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
		return nil, nil, apperrors.New(err)
	}

	logChan, _ := StartScanningLog(ctx, io.NopCloser(attachResp.Reader), WithParseLogHeader(false))
	defer attachResp.Close()

	logs := make([]*tasklog.LogFrame, 0, 20) //nolint
	for msgs := range logChan {
		logs = append(logs, msgs...)
	}

	inspectResp, err := m.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return nil, nil, apperrors.New(err)
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
