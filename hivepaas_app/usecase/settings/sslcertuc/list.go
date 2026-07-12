package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/domainhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) ListSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.ListSSLCertReq,
) (*sslcertdto.ListSSLCertResp, error) {
	req.Type = currentSettingType
	var extraLoadOpts []bunex.SelectQueryOption
	if req.Domain != "" {
		extraLoadOpts = append(extraLoadOpts,
			bunex.SelectWhereIn("setting.name IN (?)", domainhelper.CalcMatchingDomains(req.Domain)...))
	}

	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{
		ExtraLoadOpts: extraLoadOpts,
	})
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
