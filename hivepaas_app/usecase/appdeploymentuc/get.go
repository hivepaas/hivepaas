package appdeploymentuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) GetDeployment(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentReq,
) (*appdeploymentdto.GetDeploymentResp, error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var deploymentInfo *cacheentity.DeploymentInfo
	if deployment.IsNotStarted() || deployment.IsInProgress() {
		deploymentInfo, err = uc.deploymentInfoRepo.Get(ctx, deployment.ID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
	}

	triggerUserMap, err := uc.loadDeploymentTriggerUsers(ctx, uc.db, []*entity.Deployment{deployment})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appdeploymentdto.DeploymentTransformInput{
		DeploymentInfoMap: map[string]*cacheentity.DeploymentInfo{
			req.DeploymentID: deploymentInfo,
		},
		TriggerUserMap: triggerUserMap,
	}

	resp, err := appdeploymentdto.TransformDeployment(deployment, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdeploymentdto.GetDeploymentResp{
		Data: resp,
	}, nil
}
