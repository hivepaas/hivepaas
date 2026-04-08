package imserviceuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) UpdateIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.UpdateIMServiceReq,
) (*imservicedto.UpdateIMServiceResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: req.Name,
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(req.Kind)
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

	return &imservicedto.UpdateIMServiceResp{}, nil
}
