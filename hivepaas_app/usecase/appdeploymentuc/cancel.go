package appdeploymentuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) CancelDeployment(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.CancelDeploymentReq,
) (*appdeploymentdto.CancelDeploymentResp, error) {
	canceled := false
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		deployment, err := uc.deploymentRepo.GetByID(ctx, db, req.AppID, req.DeploymentID,
			bunex.SelectFor("UPDATE OF deployment SKIP LOCKED"),
		)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}

		if deployment != nil {
			if !deployment.CanCancel() {
				return hperrors.Wrap(hperrors.ErrActionNotAllowedByStatus)
			}
			deployment.Status = base.DeploymentStatusCanceled
			deployment.UpdatedAt = timeutil.NowUTC()
			err = uc.deploymentRepo.Update(ctx, db, deployment,
				bunex.UpdateColumns("status", "updated_at"),
			)
			if err != nil {
				return hperrors.Wrap(err)
			}
			canceled = true
			return nil
		}

		// Deployment is in-progress, send `cancel` command to the executor of the deployment task
		deploymentInfo, err := uc.deploymentInfoRepo.Get(ctx, req.DeploymentID)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				return hperrors.Wrap(hperrors.ErrUnavailable).
					WithMsgLog("deployment info not found, please try again later")
			}
			return hperrors.Wrap(err)
		}

		err = uc.taskControlRepo.Push(ctx, deploymentInfo.TaskID, &cacheentity.TaskControl{
			ID:  deploymentInfo.TaskID,
			Cmd: base.TaskCommandCancel,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdeploymentdto.CancelDeploymentResp{
		Data: &appdeploymentdto.CancelDeploymentDataResp{Canceled: canceled},
	}, nil
}
