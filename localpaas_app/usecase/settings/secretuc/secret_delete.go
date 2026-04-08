package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) DeleteSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.DeleteSecretReq,
) (*secretdto.DeleteSecretResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.DeleteSecretResp{}, nil
}
