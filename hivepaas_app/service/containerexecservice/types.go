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
	App                    *entity.App
	ExecOptions            docker.ExecCreateOption
	TerminalMode           bool
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
	LogStore               *tasklog.Store
	StdoutWriter           io.Writer
	StdinReader            io.Reader
	// ContainerID and NodeID run the command in that container rather than in
	// one the app's running tasks offer: a second command that has to meet the
	// files the first left, in an app with more than one replica.
	ContainerID string
	NodeID      string
}

type ContainerExecResp struct {
	ContainerID string
	NodeID      string
	// ExitCode is the command's. A command that exits non-zero is an error, and
	// the response comes with it, for its container and its exit code.
	ExitCode         int
	ExecStarted      bool
	IsRemoteExec     bool
	ExecCreateResult *client.ExecCreateResult
	ExecAttachResult *client.ExecAttachResult
	ExecStartResult  *client.ExecStartResult

	CloseFunc      func() // NOTE: need to call this when done
	ExecResizeFunc func(ctx context.Context, w, h uint) error
}
