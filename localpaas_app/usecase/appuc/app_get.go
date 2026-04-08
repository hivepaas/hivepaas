package appuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppReq,
) (*appdto.GetAppResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("display_order")),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ProjectID != req.ProjectID {
		return nil, apperrors.New(apperrors.ErrUnauthorized)
	}

	transformationInput := &appdto.AppTransformationInput{}

	if req.GetStats {
		serviceMap, err := uc.loadAppsSwarmService(ctx, app.Project.Key, []*entity.App{app})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		transformationInput.SwarmServiceMap = serviceMap
	}

	resp, err := appdto.TransformApp(app, transformationInput)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.GetAppResp{
		Data: resp,
	}, nil
}
