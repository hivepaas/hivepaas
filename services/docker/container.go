package docker

import (
	"context"
	"sync"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

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
		return nil, apperrors.NewInfra(err)
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

func (m *manager) ServiceContainerGetActive(
	ctx context.Context,
	serviceID string,
	maxRetry int,
	retryDelay time.Duration,
) (active *container.Summary, all *client.ContainerListResult, err error) {
	return m.serviceContainerGetActive(ctx, serviceID, -1, maxRetry, retryDelay)
}

func (m *manager) serviceContainerGetActive(
	ctx context.Context,
	serviceID string,
	retry int,
	maxRetry int,
	retryDelay time.Duration,
) (active *container.Summary, all *client.ContainerListResult, err error) {
	if retry >= maxRetry {
		return nil, nil, nil
	}
	listResp, err := m.ServiceContainerList(ctx, serviceID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	for i := range listResp.Items {
		c := &listResp.Items[i]
		if c.State == container.StateRunning {
			return c, listResp, nil
		}
	}

	time.Sleep(retryDelay)
	return m.serviceContainerGetActive(ctx, serviceID, retry+1, maxRetry, retryDelay)
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
		return nil, apperrors.NewInfra(err)
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
			return nil, map[string]error{containerIDs[0]: apperrors.Wrap(err)}
		}
		return map[string]*client.ContainerInspectResult{containerIDs[0]: resp}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := make(map[string]*client.ContainerInspectResult, len(containerIDs))
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			resp, err := m.ContainerInspect(ctx, containerID, options...)
			mu.Lock()
			if err != nil {
				allErrors[containerID] = apperrors.Wrap(err)
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
		return nil, apperrors.NewInfra(err)
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
		return nil, apperrors.NewInfra(err)
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
			return map[string]error{containerIDs[0]: apperrors.Wrap(err)}
		}
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			_, err := m.ContainerRestart(ctx, containerID, options...)
			if err != nil {
				mu.Lock()
				allErrors[containerID] = apperrors.Wrap(err)
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
		return nil, apperrors.NewInfra(err)
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
			return map[string]error{containerIDs[0]: apperrors.Wrap(err)}
		}
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allErrors := map[string]error{}
	for _, containerID := range containerIDs {
		wg.Go(func() {
			_, err := m.ContainerKill(ctx, containerID, signal, options...)
			if err != nil {
				mu.Lock()
				allErrors[containerID] = apperrors.Wrap(err)
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
	onlyObjectsOlderThan time.Duration,
	options ...ContainerPruneOption,
) (*client.ContainerPruneResult, error) {
	opts := client.ContainerPruneOptions{}
	if onlyObjectsOlderThan > 0 {
		FilterAdd(&opts.Filters, "until", onlyObjectsOlderThan.String())
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ContainerPrune(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}
