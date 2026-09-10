package imserviceuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) UpdateIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.UpdateIMServiceReq,
) (*imservicedto.UpdateIMServiceResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	imPlatform := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: imPlatform.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			current, err := data.Setting.AsIMService()
			if err != nil {
				return hperrors.Wrap(err)
			}
			req.KeepMaskedSecrets(imPlatform, current)

			err = pData.Setting.SetData(imPlatform)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.UpdateIMServiceResp{}, nil
}
