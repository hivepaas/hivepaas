package appuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) GetAppLogsInfo(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppLogsInfoReq,
) (*appdto.GetAppLogsInfoResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.New(apperrors.ErrUnavailable).
			WithMsgLog("service not exist for app")
	}

	taskList, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, docker.TaskStateRunning)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	dataResp := &appdto.AppLogsInfoDataResp{}
	for _, item := range taskList.Items {
		dataResp.Tasks = append(dataResp.Tasks, &appdto.TaskLogsInfoResp{
			ID: item.ID,
		})
	}

	return &appdto.GetAppLogsInfoResp{
		Data: dataResp,
	}, nil
}
