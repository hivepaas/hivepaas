package appdeploymentuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/taskservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
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
) (*appdeploymentdto.GetDeploymentLogsResp, error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID,
		bunex.SelectRelation("Tasks"),
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
		Duration:                req.Duration,
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
			Logs:          resp.Logs,
			LogChan:       resp.LogChan,
			LogChanCloser: resp.LogChanCloser,
		},
	}, nil
}
