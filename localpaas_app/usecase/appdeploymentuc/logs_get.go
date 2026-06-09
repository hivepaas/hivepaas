package appdeploymentuc

import (
	"context"
	"strings"
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
	auth *basedto.Auth, // BE CAREFUL: If req.Token presents, auth is nil
	req *appdeploymentdto.GetDeploymentLogsReq,
) (*appdeploymentdto.GetDeploymentLogsResp, error) {
	if auth == nil {
		key := strings.ReplaceAll(req.Token, "-", ":")
		ticketInfo, err := uc.consoleTicketRepo.Get(ctx, key)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
		}
		if req.DeploymentID != ticketInfo.TargetID || req.AppID != ticketInfo.AppID {
			return nil, apperrors.New(apperrors.ErrTokenInvalid)
		}
		// Remove the ticket from redis as this ticket is one-time object
		_ = uc.consoleTicketRepo.Del(ctx, key)
	}

	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID,
		bunex.SelectRelation("Tasks",
			bunex.SelectColumns("id"),
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
			StaticLogs:         resp.StaticLogs,
			RealtimeLogsStream: resp.RealtimeLogsStream,
			LogsStreamCloser:   resp.LogsStreamCloser,
		},
	}, nil
}
