package secretuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) GetSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.GetSecretReq,
) (*secretdto.GetSecretResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// NOTE: we never return decrypted data to users

	resp.Data.MustAsSecret()
	respData, err := secretdto.TransformSecret(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &secretdto.GetSecretResp{
		Data: respData,
	}, nil
}
