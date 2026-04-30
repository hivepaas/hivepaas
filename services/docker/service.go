package docker

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

type ServiceListOption func(options *swarm.ServiceListOptions)

func (m *manager) ServiceList(
	ctx context.Context,
	options ...ServiceListOption,
) ([]swarm.Service, error) {
	opts := swarm.ServiceListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ServiceListByStack(
	ctx context.Context,
	namespace string,
	options ...ServiceListOption,
) ([]swarm.Service, error) {
	options = append(options, func(opts *swarm.ServiceListOptions) {
		FilterAdd(&opts.Filters, "label", StackLabelNamespace+"="+namespace)
	})
	resp, err := m.ServiceList(ctx, options...)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ServiceGetByName(
	ctx context.Context,
	serviceName string,
	options ...ServiceListOption,
) (*swarm.Service, error) {
	options = append(options, func(opts *swarm.ServiceListOptions) {
		FilterAdd(&opts.Filters, "name", serviceName)
	})
	resp, err := m.ServiceList(ctx, options...)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	if len(resp) == 0 {
		return nil, apperrors.New(apperrors.ErrInfraNotFound).
			WithMsgLog("service '%s' not found", serviceName)
	}
	return &resp[0], nil
}

type ServiceCreateOption func(options *swarm.ServiceCreateOptions)

func (m *manager) ServiceCreate(
	ctx context.Context,
	service *swarm.ServiceSpec,
	options ...ServiceCreateOption,
) (*swarm.ServiceCreateResponse, error) {
	if service == nil {
		return nil, nil
	}
	opts := swarm.ServiceCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceCreate(ctx, *service, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ServiceUpdateOption func(options *swarm.ServiceUpdateOptions)

func (m *manager) ServiceUpdate(
	ctx context.Context,
	serviceID string,
	version *swarm.Version,
	service *swarm.ServiceSpec,
	options ...ServiceUpdateOption,
) (*swarm.ServiceUpdateResponse, error) {
	if serviceID == "" || service == nil {
		return nil, nil
	}
	opts := swarm.ServiceUpdateOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	if version == nil {
		resp, _, err := m.client.ServiceInspectWithRaw(ctx, serviceID, swarm.ServiceInspectOptions{})
		if err != nil {
			return nil, apperrors.NewInfra(err)
		}
		version = &resp.Version
	}

	resp, err := m.client.ServiceUpdate(ctx, serviceID, *version, *service, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceForceUpdate(ctx context.Context, serviceID string) error {
	service, _, err := m.client.ServiceInspectWithRaw(ctx, serviceID, swarm.ServiceInspectOptions{})
	if err != nil {
		return apperrors.NewInfra(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = m.client.ServiceUpdate(ctx, serviceID, service.Version, service.Spec, swarm.ServiceUpdateOptions{})
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

func (m *manager) ServiceRemove(ctx context.Context, serviceID string) error {
	if serviceID == "" {
		return nil
	}
	err := m.client.ServiceRemove(ctx, serviceID)
	if err != nil {
		return apperrors.NewInfra(err)
	}
	return nil
}

type ServiceInspectOption func(*swarm.ServiceInspectOptions)

func (m *manager) ServiceInspect(
	ctx context.Context,
	serviceID string,
	options ...ServiceInspectOption,
) (*swarm.Service, error) {
	if serviceID == "" {
		return nil, nil
	}

	opts := swarm.ServiceInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, _, err := m.client.ServiceInspectWithRaw(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceExists(ctx context.Context, serviceID string) bool {
	if serviceID == "" {
		return false
	}
	resp, err := m.ServiceInspect(ctx, serviceID)
	return err == nil && resp != nil
}

type ContainerLogsOption func(*container.LogsOptions)

func (m *manager) ServiceLogs(
	ctx context.Context,
	serviceID string,
	options ...ContainerLogsOption,
) (io.ReadCloser, error) {
	if serviceID == "" {
		return nil, nil
	}

	opts := container.LogsOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceLogs(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ServiceWaitUntilRunning(
	ctx context.Context,
	serviceID string,
	requireAllReplicas bool,
	requireRunningDuration time.Duration,
	checkInterval time.Duration,
	timeout time.Duration,
) (bool, error) {
	if serviceID == "" {
		return false, nil
	}
	start := time.Now()
	var isRunningFrom time.Time
	for time.Since(start) <= timeout {
		service, _, err := m.client.ServiceInspectWithRaw(ctx, serviceID, swarm.ServiceInspectOptions{})
		if err != nil {
			return false, apperrors.NewInfra(err)
		}
		if service.Spec.Mode.Replicated == nil || *service.Spec.Mode.Replicated.Replicas == 0 {
			return false, nil
		}
		if service.ServiceStatus == nil {
			return false, nil
		}
		if (requireAllReplicas && service.ServiceStatus.RunningTasks < service.ServiceStatus.DesiredTasks) ||
			(!requireAllReplicas && service.ServiceStatus.RunningTasks == 0) {
			isRunningFrom = time.Time{}
			time.Sleep(checkInterval)
			continue
		}
		if isRunningFrom.IsZero() {
			isRunningFrom = time.Now()
		}
		if time.Since(isRunningFrom) >= requireRunningDuration {
			return true, nil
		}
	}

	return false, nil
}
