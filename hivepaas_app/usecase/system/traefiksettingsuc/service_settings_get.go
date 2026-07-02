package traefiksettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

func (uc *UC) GetServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.GetServiceSettingsReq,
) (*traefiksettingsdto.GetServiceSettingsResp, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, nil, base.SettingTypeTraefikService, true)
	if err != nil {
		return nil, apperrors.New(err)
	}

	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := traefiksettingsdto.TransformServiceSettings(&traefiksettingsdto.ServiceSettingsTransformInput{
		Setting:        setting,
		TraefikService: traefikSvc,
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &traefiksettingsdto.GetServiceSettingsResp{
		Data: respData,
	}, nil
}
