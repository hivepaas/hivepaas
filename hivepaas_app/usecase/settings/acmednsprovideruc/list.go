package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

func (uc *UC) ListAcmeDnsProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.ListAcmeDnsProviderReq,
) (*acmednsproviderdto.ListAcmeDnsProviderResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := acmednsproviderdto.TransformAcmeDnsProviders(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &acmednsproviderdto.ListAcmeDnsProviderResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
