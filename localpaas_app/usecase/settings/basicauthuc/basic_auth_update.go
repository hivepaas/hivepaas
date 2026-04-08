package basicauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) UpdateBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.UpdateBasicAuthReq,
) (*basicauthdto.UpdateBasicAuthResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: req.Name,
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := pData.Setting.SetData(req.ToEntity())
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &basicauthdto.UpdateBasicAuthResp{}, nil
}
