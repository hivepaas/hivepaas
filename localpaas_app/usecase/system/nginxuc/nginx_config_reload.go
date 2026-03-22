package nginxuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/nginxuc/nginxdto"
)

func (uc *NginxUC) ReloadNginxConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *nginxdto.ReloadNginxConfigReq,
) (*nginxdto.ReloadNginxConfigResp, error) {
	err := uc.nginxService.ReloadNginxConfig(ctx, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nginxdto.ReloadNginxConfigResp{}, nil
}
