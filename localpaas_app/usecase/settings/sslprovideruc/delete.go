package sslprovideruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) DeleteSSLProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.DeleteSSLProviderReq,
) (*sslproviderdto.DeleteSSLProviderResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslproviderdto.DeleteSSLProviderResp{}, nil
}
