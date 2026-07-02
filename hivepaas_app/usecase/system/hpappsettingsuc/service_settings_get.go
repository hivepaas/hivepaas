package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetServiceSettingsReq,
) (*hpappsettingsdto.GetServiceSettingsResp, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeHivePaaSService, true)
	if err != nil {
		return nil, apperrors.New(err)
	}

	mainSvc, err := uc.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	workerSvc, err := uc.hpAppService.GetHpWorkerSwarmService(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := hpappsettingsdto.TransformServiceSettings(&hpappsettingsdto.ServiceSettingsTransformInput{
		Setting:       setting,
		MainService:   mainSvc,
		WorkerService: workerSvc,
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &hpappsettingsdto.GetServiceSettingsResp{
		Data: respData,
	}, nil
}
