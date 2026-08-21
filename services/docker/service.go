package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

const (
	retry2Times       = 2
	defaultRetryDelay = time.Second * 2
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

func (m *manager) ServiceUpdateFunc(
	ctx context.Context,
	serviceID string,
	service *swarm.Service,
	fn func(int, *swarm.Service) (bool, error),
	retryMax int,
	retryDelay time.Duration,
	options ...ServiceUpdateOption,
) (err error) {
	if serviceID == "" {
		return nil
	}
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}

	for i := range retryMax + 1 {
		if i > 0 {
			if retryDelay > 0 {
				timer := time.NewTimer(retryDelay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return apperrors.Wrap(ctx.Err())
				case <-timer.C:
				}
			}
		}

		if i > 0 || service == nil {
			inspect, e := m.ServiceInspect(ctx, serviceID)
			if e != nil { // error, need to retry
				err = apperrors.Wrap(e)
				if errors.Is(e, apperrors.ErrNotFound) {
					return err
				}
				continue
			}
			service = &inspect.Service
		}

		success, e := fn(i, service)
		if e != nil { // error from user function, no retry
			err = apperrors.Wrap(e)
			return err
		}
		if !success { // the user doesn't want to continue the update
			return nil
		}

		_, e = m.ServiceUpdate(ctx, serviceID, &service.Version, &service.Spec, options...)
		if e != nil { // error, need to retry
			err = apperrors.Wrap(e)
			continue
		}

		return nil // successful, no need to retry
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
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
		}, retry2Times, defaultRetryDelay)
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
	}, retry2Times, defaultRetryDelay)
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
		}, retry2Times, defaultRetryDelay)
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

func (m *manager) ServiceWaitUntilStopped(
	ctx context.Context,
	serviceID string,
	checkInterval time.Duration,
) (bool, error) {
	if checkInterval <= 0 {
		checkInterval = time.Second * 2 //nolint:mnd
	}

	for {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return false, apperrors.NewInfra(err)
		}

		taskListResp, err := gofn.ExecRetry2(func() (*client.TaskListResult, error) {
			return m.ServiceTaskList(ctx, serviceID, nil)
		}, retry2Times, defaultRetryDelay)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return true, nil
			}
			return false, apperrors.Wrap(err)
		}

		activeTasks := 0
		for i := range taskListResp.Items {
			state := taskListResp.Items[i].Status.State
			if !isTaskTerminalState(state) {
				activeTasks++
			}
		}

		if activeTasks == 0 {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, apperrors.Wrap(ctx.Err())
		case <-time.After(checkInterval):
		}
	}
}

func isTaskTerminalState(state swarm.TaskState) bool {
	switch state { //nolint:exhaustive
	case swarm.TaskStateComplete,
		swarm.TaskStateFailed,
		swarm.TaskStateShutdown,
		swarm.TaskStateRejected,
		swarm.TaskStateRemove,
		swarm.TaskStateOrphaned:
		return true
	default:
		return false
	}
}

//nolint:gocognit
func (m *manager) ServiceCreateToExec(
	ctx context.Context,
	image string,
	cmd []string,
	timeout time.Duration,
	checkInterval time.Duration,
	options ...ServiceCreateOption,
) (createResp *client.ServiceCreateResult, statusCode int64, err error) {
	createOpts := client.ServiceCreateOptions{
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: TempServicePrefix + gofn.RandString(5), //nolint:mnd
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:   image,
					Command: cmd,
				},
				RestartPolicy: &swarm.RestartPolicy{
					Condition: swarm.RestartPolicyConditionNone,
				},
			},
		},
	}
	for _, option := range options {
		option(&createOpts)
	}

	// Try to resolve pinned RepoDigest to bypass remote registry network lookups and start in ~400ms
	if containerSpec := createOpts.Spec.TaskTemplate.ContainerSpec; containerSpec != nil &&
		containerSpec.Image != "" &&
		!strings.Contains(containerSpec.Image, "@sha256:") {
		inspectRes, inspectErr := m.ImageInspect(ctx, containerSpec.Image)
		if inspectErr == nil && len(inspectRes.RepoDigests) > 0 {
			containerSpec.Image = inspectRes.RepoDigests[0]
		}
	}

	if createOpts.Spec.Labels == nil {
		createOpts.Spec.Labels = make(map[string]string)
	}
	createOpts.Spec.Labels[LabelTempResource] = LabelTempResourceVal
	createOpts.Spec.Labels[LabelTempCreatedAt] = time.Now().UTC().Format(time.RFC3339)

	if createOpts.Spec.TaskTemplate.ContainerSpec != nil {
		if createOpts.Spec.TaskTemplate.ContainerSpec.Labels == nil {
			createOpts.Spec.TaskTemplate.ContainerSpec.Labels = make(map[string]string)
		}
		createOpts.Spec.TaskTemplate.ContainerSpec.Labels[LabelTempResource] = LabelTempResourceVal
		createOpts.Spec.TaskTemplate.ContainerSpec.Labels[LabelTempCreatedAt] = time.Now().UTC().Format(time.RFC3339)
	}

	if createOpts.Spec.TaskTemplate.RestartPolicy == nil {
		createOpts.Spec.TaskTemplate.RestartPolicy = &swarm.RestartPolicy{}
	}
	createOpts.Spec.TaskTemplate.RestartPolicy.Condition = swarm.RestartPolicyConditionNone

	createRes, err := m.ServiceCreate(ctx, &createOpts.Spec)
	if err != nil {
		return nil, 0, apperrors.Wrap(err)
	}
	svcID := createRes.ID

	defer func() { //nolint:contextcheck
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second) //nolint:mnd
		defer cancel()
		_, _ = m.ServiceRemove(cleanupCtx, svcID)
	}()

	if timeout <= 0 {
		timeout = 60 * time.Second //nolint:mnd
	}
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	if checkInterval <= 0 {
		checkInterval = 500 * time.Millisecond //nolint:mnd
	}
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return createRes, 0, apperrors.Wrap(timeoutCtx.Err())
		case <-ticker.C:
			tasksRes, err := m.ServiceTaskList(timeoutCtx, svcID, nil)
			if err != nil || len(tasksRes.Items) == 0 {
				continue
			}

			task := &tasksRes.Items[0]
			state := task.Status.State

			var taskExitCode int64
			if task.Status.ContainerStatus != nil {
				taskExitCode = int64(task.Status.ContainerStatus.ExitCode)
			}

			if state == swarm.TaskStateComplete {
				return createRes, taskExitCode, nil
			}
			if isTaskTerminalState(state) {
				errMsg := task.Status.Err
				if errMsg == "" && taskExitCode != 0 {
					errMsg = fmt.Sprintf("task exited with status code %d", taskExitCode)
				}
				if errMsg == "" {
					errMsg = fmt.Sprintf("task finished with state %s", state)
				}
				return createRes, taskExitCode, apperrors.Wrap(errors.New(errMsg)) //nolint:err113
			}
		}
	}
}
