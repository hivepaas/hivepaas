package lpappsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappsettingsuc/lpappsettingsdto"
)

func (uc *UC) GetServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *lpappsettingsdto.GetServiceSettingsReq,
) (*lpappsettingsdto.GetServiceSettingsResp, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeLocalPaaSService, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	mainSvc, err := uc.lpAppService.GetLpAppSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	workerSvc, err := uc.lpAppService.GetLpWorkerSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := lpappsettingsdto.TransformServiceSettings(&lpappsettingsdto.ServiceSettingsTransformInput{
		Setting:       setting,
		MainService:   mainSvc,
		WorkerService: workerSvc,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &lpappsettingsdto.GetServiceSettingsResp{
		Data: respData,
	}, nil
}
