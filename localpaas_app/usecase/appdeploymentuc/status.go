package appdeploymentuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) GetDeploymentStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentStatusReq,
) (*appdeploymentdto.GetDeploymentStatusResp, error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var deploymentInfo *cacheentity.DeploymentInfo
	if deployment.IsNotStarted() || deployment.IsInProgress() {
		deploymentInfo, err = uc.deploymentInfoRepo.Get(ctx, deployment.ID)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.Wrap(err)
		}
	}

	return &appdeploymentdto.GetDeploymentStatusResp{
		Data: appdeploymentdto.TransformDeploymentStatus(deployment, deploymentInfo),
	}, nil
}
