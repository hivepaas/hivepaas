package appdeploymentuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

const (
	deploymentLogBatchThresholdPeriod = time.Millisecond * 1000
	deploymentLogBatchMaxFrame        = 20
	deploymentLogSessionTimeout       = 10 * time.Minute
)

func (uc *UC) GetDeploymentLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentLogsReq,
) (_ *appdeploymentdto.GetDeploymentLogsResp, err error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID,
		bunex.SelectRelation("Tasks",
			bunex.SelectColumns("id", "target_id"), // Must select target_id, otherwise bun will report error
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task := gofn.FirstOr(deployment.Tasks, nil)
	if task == nil {
		return nil, apperrors.NewNotFound("Deployment task")
	}

	resp, err := uc.taskService.GetTaskLogs(ctx, uc.db, &taskservice.GetTaskLogsReq{
		TaskID:                  task.ID,
		Follow:                  req.Follow,
		Since:                   req.Since,
		Duration:                req.Duration.ToDuration(),
		Tail:                    req.Tail,
		LogBatchThresholdPeriod: deploymentLogBatchThresholdPeriod,
		LogBatchMaxFrame:        deploymentLogBatchMaxFrame,
		LogSessionTimeout:       deploymentLogSessionTimeout,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdeploymentdto.GetDeploymentLogsResp{
		Data: &appdeploymentdto.DeploymentLogsDataResp{
			StaticLogs:       resp.StaticLogs,
			LogsStream:       resp.LogsStream,
			LogsStreamCloser: resp.LogsStreamCloser,
		},
	}, nil
}
