package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) GetSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.GetSecretReq,
) (*secretdto.GetSecretResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// NOTE: we never return decrypted data to users

	resp.Data.MustAsSecret()
	respData, err := secretdto.TransformSecret(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.GetSecretResp{
		Data: respData,
	}, nil
}
