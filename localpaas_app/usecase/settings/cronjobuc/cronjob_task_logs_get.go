package cronjobuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

const (
	taskLogBatchThresholdPeriod = time.Millisecond * 1000
	taskLogBatchMaxFrame        = 20
	taskLogSessionTimeout       = 10 * time.Minute
)

func (uc *UC) GetCronJobTaskLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *cronjobdto.GetCronJobTaskLogsReq,
) (*cronjobdto.GetCronJobTaskLogsResp, error) {
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
		Duration:                req.Duration,
		Tail:                    req.Tail,
		LogBatchThresholdPeriod: taskLogBatchThresholdPeriod,
		LogBatchMaxFrame:        taskLogBatchMaxFrame,
		LogSessionTimeout:       taskLogSessionTimeout,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cronjobdto.GetCronJobTaskLogsResp{
		Data: &cronjobdto.CronJobTaskLogsDataResp{
			Logs:          resp.Logs,
			LogChan:       resp.LogChan,
			LogChanCloser: resp.LogChanCloser,
		},
	}, nil
}
