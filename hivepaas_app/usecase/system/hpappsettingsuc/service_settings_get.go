package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetServiceSettingsReq,
) (*hpappsettingsdto.GetServiceSettingsResp, error) {
	db := uc.db
	setting, err := uc.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeHivePaaSService, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	mainSvc, err := uc.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	workerSvc, err := uc.hpAppService.GetHpWorkerSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := hpappsettingsdto.TransformServiceSettings(&hpappsettingsdto.ServiceSettingsTransformInput{
		Setting:       setting,
		MainService:   mainSvc,
		WorkerService: workerSvc,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	pending, err := uc.findPendingProbation(ctx, db, app.ID, base.SettingTypeHivePaaSService)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	respData.PendingChange = hpappsettingsdto.TransformPendingChange(pending)

	return &hpappsettingsdto.GetServiceSettingsResp{
		Data: respData,
	}, nil
}
