package containerexecserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	taskFindRetryMax           = 3
	taskFindRetryDelay         = time.Second * 5
	taskFindMinRunningDuration = time.Second * 10
)

func (s *service) ContainerExec(
	ctx context.Context,
	req *containerexecservice.ContainerExecReq,
) (resp *containerexecservice.ContainerExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	logStore := req.LogStore
	if logStore == nil {
		logStore = tasklog.NewNullStore()
	}

	task, _, err := s.dockerManager.ServiceTaskGetRunning(ctx, req.App.ServiceID,
		gofn.Coalesce(req.TaskMinRunningDuration, taskFindMinRunningDuration),
		gofn.Coalesce(req.TaskFindRetryMax, taskFindRetryMax),
		gofn.Coalesce(req.TaskFindRetryDelay, taskFindRetryDelay))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if task == nil {
		_ = logStore.Add(ctx, tasklog.NewWarnFrame(
			"No running task found, execution aborted", tasklog.TsNow))
		return nil, apperrors.NewNotFound("Running task of service")
	}

	dockerClient := s.dockerManager
	if task.NodeID != "" {
		dockerClient, err = s.dockerManager.NewClientForNode(ctx, task.NodeID)
		if err != nil {
			_ = logStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to connect to remote node agent: %v", err), tasklog.TsNow))
			return nil, apperrors.Wrap(err)
		}
	}
	resp = &containerexecservice.ContainerExecResp{
		DockerManager:     dockerClient,
		IsRemoteExecution: dockerClient != s.dockerManager,
	}
	resp.ExecResizeFunc = func(ctx context.Context, w, h uint) error {
		if resp.DockerManager == nil || resp.ExecCreateResult == nil || resp.ExecCreateResult.ID == "" {
			return nil
		}
		_, err := resp.DockerManager.ContainerExecResize(ctx, resp.ExecCreateResult.ID, w, h)
		return apperrors.Wrap(err)
	}

	defer func() {
		if err != nil || !req.TerminalMode {
			resp.Close()
		}
	}()

	containerID := task.Status.ContainerStatus.ContainerID
	createResp, attachResp, startResp, err := dockerClient.ContainerExec(ctx, containerID, req.ExecOptions)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.ExecCreateResult = createResp
	resp.ExecAttachResult = attachResp
	resp.ExecStartResult = startResp

	if req.TerminalMode {
		return resp, nil
	}

	logChan, _ := docker.StartScanningLog(ctx, io.NopCloser(attachResp.Reader), docker.WithParseLogHeader(false))

	for msgs := range logChan {
		_ = logStore.AddRedacted(ctx, msgs...)
	}

	execInfo, err := dockerClient.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if execInfo.ExitCode != 0 {
		_ = logStore.AddRedacted(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", execInfo.ExitCode), tasklog.TsNow))
		return nil, apperrors.Wrap(apperrors.ErrInfraActionFailed)
	}

	return resp, nil
}
