package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) ReloadHpAppConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *hpappdto.ReloadHpAppConfigReq,
) (*hpappdto.ReloadHpAppConfigResp, error) {
	err := uc.hpAppService.ReloadHpAppConfig(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &hpappdto.ReloadHpAppConfigResp{}, nil
}
