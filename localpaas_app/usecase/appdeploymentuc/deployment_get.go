package appdeploymentuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) GetDeployment(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentReq,
) (*appdeploymentdto.GetDeploymentResp, error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var deploymentInfo *cacheentity.DeploymentInfo
	if deployment.Status != base.DeploymentStatusDone && deployment.Status != base.DeploymentStatusFailed &&
		deployment.Status != base.DeploymentStatusCanceled {
		deploymentInfo, err = uc.deploymentInfoRepo.Get(ctx, deployment.ID)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.Wrap(err)
		}
	}

	resp, err := appdeploymentdto.TransformDeployment(deployment, deploymentInfo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdeploymentdto.GetDeploymentResp{
		Data: resp,
	}, nil
}
