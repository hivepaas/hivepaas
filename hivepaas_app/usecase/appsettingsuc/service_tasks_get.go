package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppServiceTasks(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppServiceTasksReq,
) (*appsettingsdto.GetAppServiceTasksResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	listResp, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, req.States)
	if err != nil {
		return nil, apperrors.New(err)
	}

	nodeListResp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := appsettingsdto.TransformServiceTasks(listResp.Items, nodeListResp.Items)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.GetAppServiceTasksResp{
		Data: resp,
	}, nil
}
