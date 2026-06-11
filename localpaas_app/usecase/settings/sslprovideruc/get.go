package sslprovideruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) GetSSLProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.GetSSLProviderReq,
) (*sslproviderdto.GetSSLProviderResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsSSLProvider().MustDecrypt()
	}

	respData, err := sslproviderdto.TransformSSLProvider(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslproviderdto.GetSSLProviderResp{
		Data: respData,
	}, nil
}
