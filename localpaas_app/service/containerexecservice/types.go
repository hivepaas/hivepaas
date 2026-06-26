package containerexecservice

import (
	"context"
	"time"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/services/docker"
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

type SchedJobExecReq struct {
	*queue.TaskExecData
	SchedJobSetting        *entity.Setting
	Project                *entity.Project
	App                    *entity.App
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
}

type SchedJobExecResp struct {
	SkipResultNotification bool
}
