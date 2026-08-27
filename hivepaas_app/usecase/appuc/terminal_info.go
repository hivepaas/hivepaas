package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetTerminalInfo(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetTerminalInfoReq,
) (_ *appdto.GetTerminalInfoResp, err error) {
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

	resp := &appdto.GetTerminalInfoResp{
		Data: &appdto.TerminalInfoDataResp{Enabled: true},
	}
	if featureSettings.TerminalSettings != nil && !featureSettings.TerminalSettings.Enabled {
		resp.Data.Enabled = false
		return resp, nil
	}

	resp.Data.SupportedShells = appdto.SupportedShells
	return resp, nil
}
