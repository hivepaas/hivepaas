package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetRoutingSettingsReq,
) (*hpappsettingsdto.GetRoutingSettingsResp, error) {
	app, err := uc.hpAppService.LoadAppByKey(ctx, uc.db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &hpappsettingsdto.RoutingSettingsTransformInput{
		App:            app,
		RoutingSetting: settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting),
	}

	err = uc.settingService.LoadRefObjectsSkipMissing(ctx, uc.db, &input.RefObjects,
		app.GetObjectScope(), false, input.RoutingSetting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := hpappsettingsdto.TransformRoutingSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	pending, err := uc.findPendingProbation(ctx, uc.db, app.ID, base.SettingTypeAppRouting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.PendingChange = hpappsettingsdto.TransformPendingChange(pending)

	return &hpappsettingsdto.GetRoutingSettingsResp{
		Data: resp,
	}, nil
}
