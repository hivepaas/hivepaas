package storagesettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

func (uc *UC) GetUniqueStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *storagesettingsdto.GetUniqueStorageSettingsReq,
) (*storagesettingsdto.GetUniqueStorageSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := storagesettingsdto.TransformStorageSettings(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &storagesettingsdto.GetUniqueStorageSettingsResp{
		Data: respData,
	}, nil
}
