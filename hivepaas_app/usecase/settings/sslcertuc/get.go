package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) GetSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.GetSSLCertReq,
) (*sslcertdto.GetSSLCertResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsSSLCert().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	respData, err := sslcertdto.TransformSSLCert(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sslcertdto.GetSSLCertResp{
		Data: respData,
	}, nil
}
