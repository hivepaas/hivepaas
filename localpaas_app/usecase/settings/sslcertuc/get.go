package sslcertuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) GetSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.GetSSLCertReq,
) (*sslcertdto.GetSSLCertResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsSSLCert().MustDecrypt()
	}

	respData, err := sslcertdto.TransformSSLCert(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslcertdto.GetSSLCertResp{
		Data: respData,
	}, nil
}
