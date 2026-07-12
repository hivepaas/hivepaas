package schedjobuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

const (
	taskLogBatchThresholdPeriod = time.Millisecond * 1000
	taskLogBatchMaxFrame        = 20
	taskLogSessionTimeout       = 10 * time.Minute
)

func (uc *UC) GetSchedJobTaskLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.GetSchedJobTaskLogsReq,
) (*schedjobdto.GetSchedJobTaskLogsResp, error) {
	req.Type = currentSettingType
	jobSetting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.JobID, false,
		bunex.SelectRelation("Tasks", bunex.SelectWhere("task.id = ?", req.TaskID)),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task := gofn.FirstOr(jobSetting.Tasks, nil)
	if task == nil {
		return nil, apperrors.NewNotFound("Task")
	}

	resp, err := uc.taskService.GetTaskLogs(ctx, uc.DB, &taskservice.GetTaskLogsReq{
		TaskID:                  task.ID,
		Follow:                  req.Follow,
		Since:                   req.Since,
		Duration:                req.Duration.ToDuration(),
		Tail:                    req.Tail,
		LogBatchThresholdPeriod: taskLogBatchThresholdPeriod,
		LogBatchMaxFrame:        taskLogBatchMaxFrame,
		LogSessionTimeout:       taskLogSessionTimeout,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.GetSchedJobTaskLogsResp{
		Data: &schedjobdto.SchedJobTaskLogsDataResp{
			StaticLogs:       resp.StaticLogs,
			LogsStream:       resp.LogsStream,
			LogsStreamCloser: resp.LogsStreamCloser,
		},
	}, nil
}
