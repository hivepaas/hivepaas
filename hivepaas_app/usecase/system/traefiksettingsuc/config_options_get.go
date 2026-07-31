package traefiksettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

func (uc *UC) GetConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.GetConfigOptionsReq,
) (*traefiksettingsdto.GetConfigOptionsResp, error) {
	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData := traefiksettingsdto.TransformConfigOptions(traefikSvc)

	return &traefiksettingsdto.GetConfigOptionsResp{
		Data: respData,
	}, nil
}
