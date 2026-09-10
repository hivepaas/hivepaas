package sslprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) UpdateSSLProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.UpdateSSLProviderReq,
) (*sslproviderdto.UpdateSSLProviderResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	sslProvider := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: sslProvider.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			current, err := data.Setting.AsSSLProvider()
			if err != nil {
				return hperrors.Wrap(err)
			}
			req.KeepMaskedSecrets(sslProvider, current)

			err = pData.Setting.SetData(sslProvider)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sslproviderdto.UpdateSSLProviderResp{}, nil
}
