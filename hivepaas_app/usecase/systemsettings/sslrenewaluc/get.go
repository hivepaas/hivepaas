package sslrenewaluc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

const (
	getSettingRetryMax = 1
)

func (uc *UC) GetSSLRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslrenewaldto.GetSSLRenewalReq,
) (_ *sslrenewaldto.GetSSLRenewalResp, err error) {
	req.Type = currentSettingType
	var resp *settings.GetUniqueSettingResp
	for i := range getSettingRetryMax + 1 {
		resp, err = uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
		if err == nil {
			break
		}
		if i < getSettingRetryMax && errors.Is(err, hperrors.ErrNotFound) {
			if e := uc.SettingInitService.InitDefaults(ctx, uc.DB); e != nil {
				return nil, hperrors.Wrap(err)
			}
			continue
		}
		return nil, hperrors.Wrap(err)
	}

	respData, err := sslrenewaldto.TransformSSLRenewal(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sslrenewaldto.GetSSLRenewalResp{
		Data: respData,
	}, nil
}
