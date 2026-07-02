package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

func (uc *UC) GetAcmeDnsProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.GetAcmeDnsProviderReq,
) (*acmednsproviderdto.GetAcmeDnsProviderResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsAcmeDnsProvider().Decrypt(); err != nil {
			return nil, apperrors.New(err)
		}
	}

	respData, err := acmednsproviderdto.TransformAcmeDnsProvider(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &acmednsproviderdto.GetAcmeDnsProviderResp{
		Data: respData,
	}, nil
}
