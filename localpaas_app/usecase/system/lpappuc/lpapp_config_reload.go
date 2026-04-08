package lpappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc/lpappdto"
)

func (uc *UC) ReloadLpAppConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *lpappdto.ReloadLpAppConfigReq,
) (*lpappdto.ReloadLpAppConfigResp, error) {
	err := uc.lpAppService.ReloadLpAppConfig(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &lpappdto.ReloadLpAppConfigResp{}, nil
}
