package sslcertuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) ListSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.ListSSLCertReq,
) (*sslcertdto.ListSSLCertResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := sslcertdto.TransformSSLCerts(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslcertdto.ListSSLCertResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
