package docker

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

type ContainerCreateOption func(*client.ContainerCreateOptions)

func (m *manager) ContainerCreate(
	ctx context.Context,
	options ...ContainerCreateOption,
) (*client.ContainerCreateResult, error) {
	opts := client.ContainerCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerCreate(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ContainerStartOption func(*client.ContainerStartOptions)

func (m *manager) ContainerStart(
	ctx context.Context,
	containerID string,
	options ...ContainerStartOption,
) (*client.ContainerStartResult, error) {
	opts := client.ContainerStartOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerStart(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ContainerWaitOption func(*client.ContainerWaitOptions)

func (m *manager) ContainerWait(
	ctx context.Context,
	containerID string,
	options ...ContainerWaitOption,
) *client.ContainerWaitResult {
	opts := client.ContainerWaitOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp := m.client.ContainerWait(ctx, containerID, opts)
	return &resp
}

type ContainerRemoveOption func(*client.ContainerRemoveOptions)

func (m *manager) ContainerRemove(
	ctx context.Context,
	containerID string,
	options ...ContainerRemoveOption,
) (*client.ContainerRemoveResult, error) {
	opts := client.ContainerRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerRemove(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ContainerListOption func(*client.ContainerListOptions)

func (m *manager) ContainerList(
	ctx context.Context,
	options ...ContainerListOption,
) (*client.ContainerListResult, error) {
	opts := client.ContainerListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceContainerList(
	ctx context.Context,
	serviceID string,
	options ...ContainerListOption,
) (*client.ContainerListResult, error) {
	options = append(options, func(opts *client.ContainerListOptions) {
		FilterAdd(&opts.Filters, "label", "com.docker.swarm.service.id="+serviceID)
	})
	return m.ContainerList(ctx, options...)
}

type ContainerInspectOption func(*client.ContainerInspectOptions)

func (m *manager) ContainerInspect(
	ctx context.Context,
	containerID string,
	options ...ContainerInspectOption,
) (*client.ContainerInspectResult, error) {
	opts := client.ContainerInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerInspect(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ContainerInspectMulti(
	ctx context.Context,
	containerIDs []string,
	options ...ContainerInspectOption,
) (map[string]*client.ContainerInspectResult, map[string]error) {
	if len(containerIDs) == 1 {
		resp, err := m.ContainerInspect(ctx, containerIDs[0], options...)
		if err != nil {
			return nil, map[string]error{containerIDs[0]: hperrors.Wrap(err)}
		}
		return map[string]*client.ContainerInspectResult{containerIDs[0]: resp}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := make(map[string]*client.ContainerInspectResult, len(containerIDs))
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			defer safego.Recover("docker.containerInspectMulti")
			resp, err := m.ContainerInspect(ctx, containerID, options...)
			mu.Lock()
			if err != nil {
				allErrors[containerID] = hperrors.Wrap(err)
			} else {
				allResults[containerID] = resp
			}
			mu.Unlock()
		})
	}
	wg.Wait()
	return allResults, allErrors
}

type ContainerLogsOption func(*client.ContainerLogsOptions)

func (m *manager) ContainerLogs(
	ctx context.Context,
	containerID string,
	options ...ContainerLogsOption,
) (client.ContainerLogsResult, error) {
	if containerID == "" {
		return nil, nil
	}

	opts := client.ContainerLogsOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return resp, nil
}

type ContainerRestartOption func(options *client.ContainerRestartOptions)

func (m *manager) ContainerRestart(
	ctx context.Context,
	containerID string,
	options ...ContainerRestartOption,
) (*client.ContainerRestartResult, error) {
	opts := client.ContainerRestartOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerRestart(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ContainerRestartMulti(
	ctx context.Context,
	containerIDs []string,
	options ...ContainerRestartOption,
) map[string]error {
	if len(containerIDs) == 1 {
		_, err := m.ContainerRestart(ctx, containerIDs[0], options...)
		if err != nil {
			return map[string]error{containerIDs[0]: hperrors.Wrap(err)}
		}
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			defer safego.Recover("docker.containerRestartMulti")
			_, err := m.ContainerRestart(ctx, containerID, options...)
			if err != nil {
				mu.Lock()
				allErrors[containerID] = hperrors.Wrap(err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	return allErrors
}

type ContainerKillOption func(options *client.ContainerKillOptions)

func (m *manager) ContainerKill(
	ctx context.Context,
	containerID string,
	signal string,
	options ...ContainerKillOption,
) (*client.ContainerKillResult, error) {
	opts := client.ContainerKillOptions{}
	opts.Signal = signal
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerKill(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ContainerKillMulti(
	ctx context.Context,
	containerIDs []string,
	signal string,
	options ...ContainerKillOption,
) map[string]error {
	if len(containerIDs) == 1 {
		_, err := m.ContainerKill(ctx, containerIDs[0], signal, options...)
		if err != nil {
			return map[string]error{containerIDs[0]: hperrors.Wrap(err)}
		}
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			defer safego.Recover("docker.containerKillMulti")
			_, err := m.ContainerKill(ctx, containerID, signal, options...)
			if err != nil {
				mu.Lock()
				allErrors[containerID] = hperrors.Wrap(err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	return allErrors
}

type ContainerPruneOption func(options *client.ContainerPruneOptions)

func (m *manager) ContainerPrune(
	ctx context.Context,
	generalRetention time.Duration,
	options ...ContainerPruneOption,
) (*client.ContainerPruneResult, error) {
	opts := client.ContainerPruneOptions{}
	if generalRetention > 0 {
		FilterAdd(&opts.Filters, "until", generalRetention.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerPrune(ctx, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ContainerUpdateOption func(*client.ContainerUpdateOptions)

func (m *manager) ContainerUpdate(
	ctx context.Context,
	containerID string,
	options ...ContainerUpdateOption,
) (*client.ContainerUpdateResult, error) {
	opts := client.ContainerUpdateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerUpdate(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ContainerCopyFrom(
	ctx context.Context,
	containerID string,
	srcPath string,
) (*client.CopyFromContainerResult, error) {
	resp, err := m.client.CopyFromContainer(ctx, containerID, client.CopyFromContainerOptions{
		SourcePath: srcPath,
	})
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}

type ContainerCopyToOption func(*client.CopyToContainerOptions)

func ContainerCopyToWithAllowOverwriteDirWithFile(allow bool) ContainerCopyToOption {
	return func(o *client.CopyToContainerOptions) {
		o.AllowOverwriteDirWithFile = allow
	}
}

func ContainerCopyToWithCopyUIDGID(copyUIDGID bool) ContainerCopyToOption {
	return func(o *client.CopyToContainerOptions) {
		o.CopyUIDGID = copyUIDGID
	}
}

func (m *manager) ContainerCopyTo(
	ctx context.Context,
	containerID string,
	dstPath string,
	content io.Reader,
	options ...ContainerCopyToOption,
) (*client.CopyToContainerResult, error) {
	opts := client.CopyToContainerOptions{
		DestinationPath: dstPath,
		Content:         content,
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.CopyToContainer(ctx, containerID, opts)
	if err != nil {
		return nil, hperrors.NewInfra(err)
	}
	return &resp, nil
}
