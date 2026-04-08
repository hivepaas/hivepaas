package appdeploymentuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) ListDeployment(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.ListDeploymentReq,
) (*appdeploymentdto.ListDeploymentResp, error) {
	deploymentInfoMap, err := uc.deploymentInfoRepo.GetAll(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	inprogressDeploymentIDs := make([]string, 0, len(deploymentInfoMap))
	for id, info := range deploymentInfoMap {
		if info.Status == base.DeploymentStatusInProgress {
			inprogressDeploymentIDs = append(inprogressDeploymentIDs, id)
		}
	}

	var listOpts []bunex.SelectQueryOption
	if len(req.Status) > 0 { //nolint:nestif
		statuses := req.Status
		if gofn.Contain(statuses, base.DeploymentStatusInProgress) {
			cond := bunex.SelectWhereIn("deployment.id IN (?)", inprogressDeploymentIDs...)
			statuses = gofn.Drop(statuses, base.DeploymentStatusInProgress)
			if len(statuses) == 0 {
				listOpts = append(listOpts, cond)
			} else {
				listOpts = append(listOpts, cond,
					bunex.SelectWhereOrGroup(
						bunex.SelectWhereNotIn("deployment.id NOT IN (?)", inprogressDeploymentIDs...),
						bunex.SelectWhereIn("deployment.status IN (?)", statuses),
					),
				)
			}
		} else {
			listOpts = append(listOpts,
				bunex.SelectWhereNotIn("deployment.id NOT IN (?)", inprogressDeploymentIDs...),
				bunex.SelectWhereIn("deployment.status IN (?)", statuses...))
		}
	}
	if req.Search != "" { //nolint
		// TODO: add implementation
	}
	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhereIn("deployment.id IN (?)", auth.AllowObjectIDs...),
		)
	}

	deployments, paging, err := uc.deploymentRepo.List(ctx, uc.db, req.AppID, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appdeploymentdto.TransformDeployments(deployments, deploymentInfoMap)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdeploymentdto.ListDeploymentResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}
