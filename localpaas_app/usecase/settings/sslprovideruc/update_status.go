package sslprovideruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) UpdateSSLProviderStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.UpdateSSLProviderStatusReq,
) (*sslproviderdto.UpdateSSLProviderStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslproviderdto.UpdateSSLProviderStatusResp{}, nil
}
