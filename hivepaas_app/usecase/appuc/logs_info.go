package appuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, hperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	resp := &appdto.GetAppLogsInfoResp{
		Data: &appdto.AppLogsInfoDataResp{Enabled: true},
	}
	if featureSettings.LoggingSettings != nil && !featureSettings.LoggingSettings.Enabled {
		resp.Data.Enabled = false
		return resp, nil
	}

	history, err := uc.loggingService.AppHistory(ctx, uc.db, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Data.History = &appdto.AppLogHistoryInfoResp{
		Available: history.Available,
		Reason:    string(history.Reason),
		Retention: historyRetention(history.Retention),
	}

	taskList, err := uc.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, item := range taskList.Items {
		resp.Data.Tasks = append(resp.Data.Tasks, &appdto.TaskLogsInfoResp{
			ID: item.ID,
		})
	}

	return resp, nil
}

// historyRetention writes a retention the way the logging setting writes it,
// and writes nothing at all for the zero the service reports when it does not
// know - "0s" would read as "kept for no time".
func historyRetention(d timeutil.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.String()
}
