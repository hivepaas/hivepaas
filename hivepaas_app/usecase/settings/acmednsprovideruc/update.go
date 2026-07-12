package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

func (uc *UC) UpdateAcmeDnsProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.UpdateAcmeDnsProviderReq,
) (*acmednsproviderdto.UpdateAcmeDnsProviderResp, error) {
	req.Type = currentSettingType
	acmeDnsProvider := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: acmeDnsProvider.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			if err := pData.Setting.SetData(acmeDnsProvider); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &acmednsproviderdto.UpdateAcmeDnsProviderResp{}, nil
}
