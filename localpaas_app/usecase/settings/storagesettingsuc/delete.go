package storagesettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

func (uc *UC) DeleteStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *storagesettingsdto.DeleteStorageSettingsReq,
) (*storagesettingsdto.DeleteStorageSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &storagesettingsdto.DeleteStorageSettingsResp{}, nil
}
