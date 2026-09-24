package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppDockerAPISettings(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetAppDockerAPISettingsReq,
) (*appsettingsdto.GetAppDockerAPISettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	setting, err := uc.appDockerAPISetting(ctx, uc.db, app.ID)
	if err != nil {
		return nil, err
	}
	resp, err := appsettingsdto.TransformAppDockerAPISettings(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.GetAppDockerAPISettingsResp{Data: resp}, nil
}

// appDockerAPISetting is an app's Docker API setting whatever its status, nil
// when it has none.
func (uc *UC) appDockerAPISetting(ctx context.Context, db database.IDB, appID string) (*entity.Setting, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDockerAPI),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}
	return settings[0], nil
}
