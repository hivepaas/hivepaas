package sslprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

func (uc *UC) CreateSSLProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslproviderdto.CreateSSLProviderReq,
) (*sslproviderdto.CreateSSLProviderResp, error) {
	req.Type = currentSettingType
	sslProvider := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: sslProvider.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			err := pData.Setting.SetData(sslProvider)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslproviderdto.CreateSSLProviderResp{
		Data: resp.Data,
	}, nil
}
