package containerexecserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/interface/agent/client/containerservice"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/service/agentservice"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	taskFindRetryMax           = 3
	taskFindRetryDelay         = time.Second * 5
	taskFindMinRunningDuration = time.Second * 10

	containerExecRetryMax = 1
)

func (s *service) ContainerExec(
	ctx context.Context,
	req *containerexecservice.ContainerExecReq,
) (resp *containerexecservice.ContainerExecResp, err error) {
	return s.containerExec(ctx, req, 0)
}

func (s *service) containerExec(
	ctx context.Context,
	req *containerexecservice.ContainerExecReq,
	retry int,
) (resp *containerexecservice.ContainerExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	logStore := req.LogStore
	if logStore == nil {
		logStore = tasklog.NewNullStore()
	}

	serviceID := req.App.ServiceID
	if serviceID == "" {
		return nil, apperrors.NewNotFound("Swarm service")
	}

	inspectResp, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return nil, apperrors.New(err)
	}
	svcMode := &inspectResp.Service.Spec.Mode
	if svcMode.Replicated != nil && (svcMode.Replicated.Replicas == nil || *svcMode.Replicated.Replicas == 0) {
		return &containerexecservice.ContainerExecResp{ExecStarted: false}, nil
	}

	task, _, err := s.dockerManager.ServiceTaskGetRunning(ctx, serviceID,
		gofn.Coalesce(req.TaskMinRunningDuration, taskFindMinRunningDuration),
		gofn.Coalesce(req.TaskFindRetryMax, taskFindRetryMax),
		gofn.Coalesce(req.TaskFindRetryDelay, taskFindRetryDelay),
		nil)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if task == nil {
		_ = logStore.Add(ctx, tasklog.NewWarnFrame(
			"No running task found, execution aborted", tasklog.TsNow))
		return nil, apperrors.NewNotFound("Running task of service")
	}

	currNodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, apperrors.New(err)
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

	resp = &containerexecservice.ContainerExecResp{
		IsRemoteExec:   isRemote,
		CloseFunc:      execHelper.Close,
		ExecResizeFunc: execHelper.ExecResize,
	}
	defer func() {
		if err != nil || !req.TerminalMode {
			resp.CloseFunc()
		}
	}()

	containerID := task.Status.ContainerStatus.ContainerID
	createResp, attachResp, startResp, err := execHelper.ExecCreate(ctx, containerID, req)
	if err != nil {
		// retry one more time with excluding the current node from the list
		if execHelper.retryable && retry < containerExecRetryMax {
			_ = logStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf(
				"Execution failed to start in node %s, retrying...", execHelper.targetNodeID), tasklog.TsNow))
			return s.containerExec(ctx, req, retry+1)
		}
		return nil, apperrors.New(err)
	}

	resp.ExecStarted = true
	resp.ExecCreateResult = createResp
	resp.ExecAttachResult = attachResp
	resp.ExecStartResult = startResp

	if req.TerminalMode {
		return resp, nil
	}

	logChan, _ := docker.StartScanningLog(ctx, io.NopCloser(attachResp.Reader),
		docker.WithParseLogHeader(!execHelper.isTTY), docker.WithStdoutWriter(req.StdoutWriter))
	for msgs := range logChan {
		_ = logStore.AddRedacted(ctx, msgs...)
	}

	exitCode, err := execHelper.GetExecExitCode(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if exitCode != 0 {
		_ = logStore.AddRedacted(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", exitCode), tasklog.TsNow))
		return nil, apperrors.New(apperrors.ErrInfraActionFailed)
	}

	return resp, nil
}

type containerExecHelper struct {
	dockerClient docker.Manager                        // for local container exec
	remoteStream *containerservice.ContainerExecStream // for remote container exec
	agentClient  containerservice.ContainerServiceClient

	targetNodeID string

	createResult *client.ExecCreateResult
	attachResult *client.ExecAttachResult
	isTTY        bool

	retryable    bool
	agentService agentservice.Service
	logStore     *tasklog.Store
}

func (h *containerExecHelper) ExecCreate(
	ctx context.Context,
	containerID string,
	req *containerexecservice.ContainerExecReq,
) (_ *client.ExecCreateResult, _ *client.ExecAttachResult, _ *client.ExecStartResult, err error) {
	defer func() {
		if err != nil {
			h.Close()
		} else {
			h.retryable = false // Exec created, not allow to retry when a subsequence step fails
		}
	}()

	// Local exec
	if h.dockerClient != nil {
		createRes, attachRes, startRes, err := h.dockerClient.ContainerExec(ctx, containerID,
			func(opts *client.ExecCreateOptions) {
				req.ExecOptions(opts)
				h.isTTY = opts.TTY || opts.ConsoleSize.Width > 0 && opts.ConsoleSize.Height > 0
			})
		if err != nil {
			return nil, nil, nil, apperrors.New(err)
		}
		h.createResult = createRes
		h.attachResult = attachRes
		return createRes, attachRes, startRes, nil
	}

	// Remote exec
	if h.remoteStream == nil {
		agentAddr, err := h.agentService.GetAgentAddrForNode(ctx, h.targetNodeID)
		if err != nil {
			_ = h.logStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to get IP of agent for node %s: %v", h.targetNodeID, err), tasklog.TsNow))
			return nil, nil, nil, apperrors.New(err)
		}

		h.agentClient, err = containerservice.NewContainerServiceClient(agentAddr)
		if err != nil {
			_ = h.logStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to connect to agent at %s: %v", agentAddr, err), tasklog.TsNow))
			return nil, nil, nil, apperrors.New(err)
		}

		h.remoteStream, err = h.agentClient.ContainerExec(ctx)
		if err != nil {
			return nil, nil, nil, apperrors.New(err)
		}
	}

	err = h.remoteStream.SendExecCreate(containerID, req.ExecOptions)
	if err != nil {
		return nil, nil, nil, apperrors.New(err)
	}

	h.createResult = &client.ExecCreateResult{ID: "remote"}
	h.attachResult = h.remoteStream.ToExecAttachResult()
	return h.createResult, h.attachResult, &client.ExecStartResult{}, nil
}

func (h *containerExecHelper) ExecResize(
	ctx context.Context,
	width, height uint,
) error {
	// Local exec
	if h.dockerClient != nil {
		_, err := h.dockerClient.ContainerExecResize(ctx, h.createResult.ID, width, height)
		return apperrors.New(err)
	}

	// Remote exec
	return apperrors.New(h.remoteStream.SendResize(width, height))
}

func (h *containerExecHelper) GetExecExitCode(
	ctx context.Context,
) (int, error) {
	// Local exec
	if h.dockerClient != nil {
		execInfo, err := h.dockerClient.ContainerExecInspect(ctx, h.createResult.ID)
		if err != nil {
			return 0, apperrors.New(err)
		}
		return execInfo.ExitCode, nil
	}

	// Remote exec
	exitCode, ok := h.remoteStream.GetExitCode()
	if !ok {
		return 0, apperrors.New(apperrors.ErrGRPCRequestFailed).
			WithParam("Error", "stream closed without exit code")
	}
	return int(exitCode), nil
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
