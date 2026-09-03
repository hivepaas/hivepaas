package containerexecserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/containerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	taskFindRetryMax           = 5
	taskFindRetryDelay         = time.Second * 2
	taskFindMinRunningDuration = time.Second * 10

	containerExecRetryMax   = 2
	containerExecRetryDelay = 2 * time.Second
)

func (s *service) ContainerExec(
	ctx context.Context,
	req *containerexecservice.ContainerExecReq,
) (resp *containerexecservice.ContainerExecResp, lastErr error) {
	for i := range containerExecRetryMax + 1 {
		var retryable bool
		resp, retryable, lastErr = s.containerExec(ctx, req)
		if lastErr == nil {
			return resp, nil
		}
		if i >= containerExecRetryMax || !retryable {
			break
		}
		if req.LogStore != nil {
			_ = req.LogStore.Add(ctx, tasklog.NewOutFrame("Retrying command execution for service "+
				req.App.ServiceID, tasklog.TsNow))
		}
		delay := containerExecRetryDelay + time.Duration(i)*time.Second
		if sleepErr := timeutil.SleepCtx(ctx, delay); sleepErr != nil {
			return nil, hperrors.Wrap(sleepErr)
		}
	}
	return nil, hperrors.Wrap(lastErr)
}

func (s *service) containerExec(
	ctx context.Context,
	req *containerexecservice.ContainerExecReq,
) (resp *containerexecservice.ContainerExecResp, retryable bool, err error) {
	defer safego.RecoverTo(&err)

	logStore := req.LogStore
	if logStore == nil {
		logStore = tasklog.NewNullStore()
	}

	serviceID := req.App.ServiceID
	if serviceID == "" {
		return nil, false, hperrors.NewNotFound("Swarm service")
	}

	inspectResp, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return nil, true, hperrors.Wrap(err)
	}
	svcMode := &inspectResp.Service.Spec.Mode
	if svcMode.Replicated != nil && (svcMode.Replicated.Replicas == nil || *svcMode.Replicated.Replicas == 0) {
		return &containerexecservice.ContainerExecResp{ExecStarted: false}, false, nil
	}

	task, _, err := s.dockerManager.ServiceTaskGetRunning(ctx, serviceID,
		gofn.Coalesce(req.TaskMinRunningDuration, taskFindMinRunningDuration),
		gofn.Coalesce(req.TaskFindRetryMax, taskFindRetryMax),
		gofn.Coalesce(req.TaskFindRetryDelay, taskFindRetryDelay),
		nil)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	if task == nil {
		_ = logStore.Add(ctx, tasklog.NewWarnFrame("No running task found for service: "+serviceID,
			tasklog.TsNow))
		return nil, false, hperrors.NewNotFound("Running task of service")
	}

	currNodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, true, hperrors.Wrap(err)
	}

	isRemote := task.NodeID != "" && task.NodeID != currNodeID
	if config.Current.DevMode.Enabled && config.Current.DevMode.ForceAgentLocal {
		isRemote = true
	}

	execHelper := &containerExecHelper{
		logStore:     logStore,
		agentService: s.agentService,
		dockerClient: gofn.If(isRemote, nil, s.dockerManager),
		targetNodeID: task.NodeID,
		retryable:    true,
	}

	containerID := task.Status.ContainerStatus.ContainerID
	resp = &containerexecservice.ContainerExecResp{
		ContainerID:    containerID,
		NodeID:         task.NodeID,
		IsRemoteExec:   isRemote,
		CloseFunc:      execHelper.Close,
		ExecResizeFunc: execHelper.ExecResize,
	}
	defer func() {
		if err != nil || !req.TerminalMode {
			resp.CloseFunc()
		}
	}()

	err = execHelper.ExecCreateAndStart(ctx, containerID, req)
	if err != nil {
		_ = logStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf("Execution failed to start in node %s",
			execHelper.targetNodeID), tasklog.TsNow))
		return nil, execHelper.retryable, hperrors.Wrap(err)
	}

	resp.ExecStarted = true
	resp.ExecCreateResult = execHelper.createResult
	resp.ExecAttachResult = execHelper.attachResult
	resp.ExecStartResult = execHelper.startResult

	if req.StdinReader != nil {
		go func() {
			defer safego.Recover("containerexec.copyStdin")
			_, _ = io.Copy(execHelper.attachResult.Conn, req.StdinReader)
			_ = execHelper.attachResult.CloseWrite()
		}()
	}

	if req.TerminalMode {
		return resp, false, nil
	}

	logChan, _ := docker.StartScanningLog(ctx, io.NopCloser(execHelper.attachResult.Reader),
		docker.WithParseLogHeader(!execHelper.isTTY), docker.WithStdoutWriter(req.StdoutWriter))
	for msgs := range logChan {
		_ = logStore.AddRedacted(ctx, msgs...)
	}

	exitCode, retryable, err := execHelper.GetExecExitCode(ctx)
	if err != nil {
		return nil, retryable, hperrors.Wrap(err)
	}
	if exitCode != 0 {
		_ = logStore.AddRedacted(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", exitCode), tasklog.TsNow))
		return nil, false, hperrors.Wrap(hperrors.ErrInfraActionFailed)
	}

	return resp, false, nil
}

type containerExecHelper struct {
	dockerClient docker.Manager                        // for local container exec
	remoteStream *containerservice.ContainerExecStream // for remote container exec
	agentClient  containerservice.ContainerServiceClient

	targetNodeID string

	createResult *client.ExecCreateResult
	attachResult *client.ExecAttachResult
	startResult  *client.ExecStartResult
	isTTY        bool

	retryable    bool
	agentService agentservice.Service
	logStore     *tasklog.Store
}

func (h *containerExecHelper) ExecCreateAndStart(
	ctx context.Context,
	containerID string,
	req *containerexecservice.ContainerExecReq,
) (err error) {
	defer func() {
		if err != nil {
			_ = h.calcIsRetryable(ctx)
			h.Close()
		} else {
			h.retryable = false // Exec created, not allow to retry when a subsequence step fails
		}
	}()

	// Local exec
	if h.dockerClient != nil {
		h.createResult, h.attachResult, h.startResult, err = h.dockerClient.ContainerExec(ctx, containerID,
			func(opts *client.ExecCreateOptions) {
				req.ExecOptions(opts)
				if req.StdinReader != nil {
					opts.AttachStdin = true
				}
				h.isTTY = opts.TTY
			})
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	}

	// Remote exec
	if h.remoteStream == nil {
		agentAddr, err := h.agentService.GetAgentAddrForNode(ctx, h.targetNodeID)
		if err != nil {
			_ = h.logStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to get IP of agent for node %s: %v", h.targetNodeID, err), tasklog.TsNow))
			return hperrors.Wrap(err)
		}

		h.agentClient, err = containerservice.NewContainerServiceClient(agentAddr)
		if err != nil {
			_ = h.logStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to connect to agent at %s: %v", agentAddr, err), tasklog.TsNow))
			return hperrors.Wrap(err)
		}

		h.remoteStream, err = h.agentClient.ContainerExec(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	opts := &client.ExecCreateOptions{}
	req.ExecOptions(opts)
	h.isTTY = opts.TTY

	err = h.remoteStream.SendExecCreate(containerID, req.ExecOptions)
	if err != nil {
		return hperrors.Wrap(err)
	}

	h.createResult = &client.ExecCreateResult{ID: "remote"}
	h.attachResult = h.remoteStream.ToExecAttachResult()
	h.startResult = &client.ExecStartResult{}
	return nil
}

func (h *containerExecHelper) ExecResize(
	ctx context.Context,
	width, height uint,
) error {
	// Local exec
	if h.dockerClient != nil {
		_, err := h.dockerClient.ContainerExecResize(ctx, h.createResult.ID, width, height)
		return hperrors.Wrap(err)
	}

	// Remote exec
	return hperrors.Wrap(h.remoteStream.SendResize(width, height))
}

func (h *containerExecHelper) GetExecExitCode(
	ctx context.Context,
) (code int, retryable bool, err error) {
	// Local exec
	if h.dockerClient != nil {
		execInfo, err := h.dockerClient.ContainerExecInspect(ctx, h.createResult.ID)
		if err != nil {
			return 0, false, hperrors.Wrap(err)
		}
		return execInfo.ExitCode, false, nil
	}

	// Remote exec
	exitCode, ok := h.remoteStream.GetExitCode()
	if !ok {
		_ = h.calcIsRetryable(ctx)
		return 0, h.retryable, hperrors.Wrap(hperrors.ErrGRPCRequestFailed).
			WithParam("Error", "stream closed without exit code")
	}
	return int(exitCode), false, nil
}

func (h *containerExecHelper) Close() {
	if h.attachResult != nil {
		h.attachResult.Close()
	}
	if h.remoteStream != nil {
		_ = h.remoteStream.Close()
	}
	if h.agentClient != nil {
		_ = h.agentClient.Close()
	}
}

func (h *containerExecHelper) calcIsRetryable(
	ctx context.Context,
) error {
	if h.dockerClient != nil {
		var execID string
		if h.createResult != nil {
			execID = h.createResult.ID
		}
		canRetry, err := h.dockerClient.CanRetryExec(ctx, execID)
		if err != nil {
			h.retryable = false
			return hperrors.Wrap(err)
		}
		h.retryable = canRetry
		return nil
	}

	if h.remoteStream != nil {
		if canRetry, ok := h.remoteStream.GetCanRetry(); ok {
			h.retryable = canRetry
			return nil
		}
	}
	return nil
}
