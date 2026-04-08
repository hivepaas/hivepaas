package sslrenewaluc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

func (uc *UC) GetSSLRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslrenewaldto.GetSSLRenewalReq,
) (*sslrenewaldto.GetSSLRenewalResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := sslrenewaldto.TransformSSLRenewal(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslrenewaldto.GetSSLRenewalResp{
		Data: respData,
	}, nil
}
