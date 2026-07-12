package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) GetHpAppReleaseInfo(
	ctx context.Context,
	_ *basedto.Auth,
	_ *hpappdto.GetHpAppReleaseInfoReq,
) (*hpappdto.GetHpAppReleaseInfoResp, error) {
	info, err := uc.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &hpappdto.GetHpAppReleaseInfoResp{
		Data: &hpappdto.HpAppReleaseInfoResp{
			AppReleaseInfo: info,
		},
	}, nil
}
