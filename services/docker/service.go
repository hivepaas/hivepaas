package docker

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type ServiceListOption func(options *client.ServiceListOptions)

func (m *manager) ServiceList(
	ctx context.Context,
	options ...ServiceListOption,
) (*client.ServiceListResult, error) {
	opts := client.ServiceListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceList(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceListByStack(
	ctx context.Context,
	namespace string,
	options ...ServiceListOption,
) (*client.ServiceListResult, error) {
	options = append(options, func(opts *client.ServiceListOptions) {
		FilterAdd(&opts.Filters, "label", StackLabelNamespace+"="+namespace)
	})
	resp, err := m.ServiceList(ctx, options...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (m *manager) ServiceGetByName(
	ctx context.Context,
	serviceName string,
	status bool,
) (*swarm.Service, error) {
	option := func(opts *client.ServiceListOptions) {
		FilterAdd(&opts.Filters, "name", serviceName)
		opts.Status = status
	}
	resp, err := m.ServiceList(ctx, option)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(resp.Items) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrInfraNotFound).
			WithMsgLog("service '%s' not found", serviceName)
	}
	return &resp.Items[0], nil
}

type ServiceInspectOption func(*client.ServiceInspectOptions)

func (m *manager) ServiceInspect(
	ctx context.Context,
	serviceID string,
	options ...ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	if serviceID == "" {
		return nil, nil
	}

	opts := client.ServiceInspectOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceInspect(ctx, serviceID, opts)
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

type ServiceCreateOption func(options *client.ServiceCreateOptions)

func (m *manager) ServiceCreate(
	ctx context.Context,
	spec *swarm.ServiceSpec,
	options ...ServiceCreateOption,
) (*client.ServiceCreateResult, error) {
	if spec == nil {
		return nil, nil
	}
	opts := client.ServiceCreateOptions{
		Spec: *spec,
	}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceCreate(ctx, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ServiceUpdateOption func(options *client.ServiceUpdateOptions)

func (m *manager) ServiceUpdate(
	ctx context.Context,
	serviceID string,
	version *swarm.Version,
	spec *swarm.ServiceSpec,
	options ...ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	if serviceID == "" || spec == nil {
		return nil, nil
	}
	opts := client.ServiceUpdateOptions{
		Spec: *spec,
	}
	for _, opt := range options {
		opt(&opts)
	}

	if version == nil {
		inspectResp, err := m.ServiceInspect(ctx, serviceID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		version = &inspectResp.Service.Version
	}
	opts.Version = *version

	resp, err := m.client.ServiceUpdate(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceRollback(
	ctx context.Context,
	serviceID string,
	options ...ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	if serviceID == "" {
		return nil, nil
	}
	opts := client.ServiceUpdateOptions{
		Rollback: "previous",
	}
	for _, opt := range options {
		opt(&opts)
	}

	inspectResp, err := m.ServiceInspect(ctx, serviceID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	opts.Version = inspectResp.Service.Version

	resp, err := m.client.ServiceUpdate(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

func (m *manager) ServiceForceUpdate(ctx context.Context, serviceID string) error {
	if serviceID == "" {
		return nil
	}
	resp, err := m.client.ServiceInspect(ctx, serviceID, client.ServiceInspectOptions{})
	if err != nil {
		return apperrors.NewInfra(err)
	}

	resp.Service.Spec.TaskTemplate.ForceUpdate++
	_, err = m.ServiceUpdate(ctx, serviceID, &resp.Service.Version, &resp.Service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

type ServiceRemoveOption func(options *client.ServiceRemoveOptions)

func (m *manager) ServiceRemove(
	ctx context.Context,
	serviceID string,
	options ...ServiceRemoveOption,
) (*client.ServiceRemoveResult, error) {
	if serviceID == "" {
		return nil, nil
	}
	opts := client.ServiceRemoveOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	resp, err := m.client.ServiceRemove(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return &resp, nil
}

type ServiceLogsOption func(*client.ServiceLogsOptions)

func (m *manager) ServiceLogs(
	ctx context.Context,
	serviceID string,
	options ...ServiceLogsOption,
) (client.ServiceLogsResult, error) {
	if serviceID == "" {
		return nil, nil
	}

	opts := client.ServiceLogsOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	resp, err := m.client.ServiceLogs(ctx, serviceID, opts)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	return resp, nil
}

func (m *manager) ServiceUpdateWait(
	ctx context.Context,
	serviceID string,
	inspectInterval time.Duration,
) (*swarm.Service, error) {
	if serviceID == "" {
		return nil, nil
	}
	for {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return nil, apperrors.NewInfra(err)
		}

		inspectResp, err := gofn.ExecRetryCtx2(ctx, func() (*client.ServiceInspectResult, error) {
			return m.ServiceInspect(ctx, serviceID)
		}, 2, time.Second*3) //nolint:mnd
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		service := &inspectResp.Service
		if service.UpdateStatus == nil ||
			service.UpdateStatus.State == swarm.UpdateStateCompleted ||
			service.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
			return service, nil
		}

		select {
		case <-ctx.Done():
			return nil, apperrors.Wrap(ctx.Err())
		case <-time.After(inspectInterval):
		}
	}
}

func (m *manager) ServiceWaitUntilRunning(
	ctx context.Context,
	serviceID string,
	requireAllReplicas bool,
	requireRunningDuration time.Duration,
	checkInterval time.Duration,
) (bool, error) {
	if serviceID == "" {
		return false, nil
	}

	inspectResp, err := gofn.ExecRetry2(func() (*client.ServiceInspectResult, error) {
		return m.ServiceInspect(ctx, serviceID)
	}, 2, time.Second*3) //nolint:mnd
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	// Service must be a replicated one
	service := &inspectResp.Service
	if service.Spec.Mode.Replicated == nil {
		return false, nil
	}
	desiredTasks := int(*service.Spec.Mode.Replicated.Replicas) //nolint:gosec
	if desiredTasks == 0 {
		return false, nil
	}

	for {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return false, apperrors.NewInfra(err)
		}

		taskListResp, err := gofn.ExecRetry2(func() (*client.TaskListResult, error) {
			return m.ServiceTaskList(ctx, serviceID, []swarm.TaskState{swarm.TaskStateRunning})
		}, 2, time.Second*3) //nolint:mnd
		if err != nil {
			return false, apperrors.Wrap(err)
		}

		satisfiedTasks := 0
		timeNow := time.Now()
		for i := range taskListResp.Items {
			t := &taskListResp.Items[i]
			if t.Status.State == swarm.TaskStateRunning && timeNow.Sub(t.Status.Timestamp) > requireRunningDuration {
				satisfiedTasks++
			}
		}

		if (requireAllReplicas && satisfiedTasks < desiredTasks) || (!requireAllReplicas && satisfiedTasks == 0) {
			select {
			case <-ctx.Done():
				return false, apperrors.Wrap(ctx.Err())
			case <-time.After(checkInterval):
			}
			continue
		}
		return true, nil
	}
}
