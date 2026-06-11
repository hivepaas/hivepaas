package sslprovideruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) ListSSLProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.ListSSLProviderReq,
) (*sslproviderdto.ListSSLProviderResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := sslproviderdto.TransformSSLProviders(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslproviderdto.ListSSLProviderResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
