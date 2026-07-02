package containerexecservice

import (
	"context"
	"io"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/docker"
)

type ContainerExecReq struct {
	Project                *entity.Project
	App                    *entity.App
	ExecOptions            docker.ExecCreateOption
	TerminalMode           bool
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
	LogStore               *tasklog.Store
	StdoutWriter           io.Writer
}

type ContainerExecResp struct {
	ExecStarted      bool
	IsRemoteExec     bool
	ExecCreateResult *client.ExecCreateResult
	ExecAttachResult *client.ExecAttachResult
	ExecStartResult  *client.ExecStartResult

	CloseFunc      func() // NOTE: need to call this when done
	ExecResizeFunc func(ctx context.Context, w, h uint) error
}
