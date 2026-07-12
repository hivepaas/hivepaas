package appuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetAppLogsInfo(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppLogsInfoReq,
) (*appdto.GetAppLogsInfoResp, error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	resp := &appdto.GetAppLogsInfoResp{
		Data: &appdto.AppLogsInfoDataResp{Enabled: true},
	}
	if featureSettings.LoggingSettings != nil && !featureSettings.LoggingSettings.Enabled {
		resp.Data.Enabled = false
		return resp, nil
	}

	taskList, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, item := range taskList.Items {
		resp.Data.Tasks = append(resp.Data.Tasks, &appdto.TaskLogsInfoResp{
			ID: item.ID,
		})
	}

	return resp, nil
}
