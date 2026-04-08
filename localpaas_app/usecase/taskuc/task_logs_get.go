package taskuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

const (
	taskLogBatchThresholdPeriod = time.Millisecond * 1000
	taskLogBatchMaxFrame        = 20
	taskLogSessionTimeout       = 10 * time.Minute
)

func (uc *UC) GetTaskLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskLogsReq,
) (*taskdto.GetTaskLogsResp, error) {
	resp, err := uc.taskService.GetTaskLogs(ctx, uc.db, &taskservice.GetTaskLogsReq{
		TaskID:                  req.TaskID,
		Follow:                  req.Follow,
		Since:                   req.Since,
		Duration:                req.Duration,
		Tail:                    req.Tail,
		LogBatchThresholdPeriod: taskLogBatchThresholdPeriod,
		LogBatchMaxFrame:        taskLogBatchMaxFrame,
		LogSessionTimeout:       taskLogSessionTimeout,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.GetTaskLogsResp{
		Data: &taskdto.TaskLogsDataResp{
			Logs:          resp.Logs,
			LogChan:       resp.LogChan,
			LogChanCloser: resp.LogChanCloser,
		},
	}, nil
}
