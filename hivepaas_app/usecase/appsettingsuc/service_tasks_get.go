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
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	listResp, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, req.States)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	nodeListResp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformServiceTasks(listResp.Items, nodeListResp.Items)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppServiceTasksResp{
		Data: resp,
	}, nil
}
