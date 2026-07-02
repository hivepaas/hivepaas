package accesstokenuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) UpdateAccessToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.UpdateAccessTokenReq,
) (*accesstokendto.UpdateAccessTokenResp, error) {
	req.Type = currentSettingType
	accessToken := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: accessToken.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = gofn.Coalesce(string(req.Kind), pData.Setting.Kind)
			pData.Setting.ExpireAt = req.ExpireAt
			err := pData.Setting.SetData(accessToken)
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &accesstokendto.UpdateAccessTokenResp{}, nil
}
