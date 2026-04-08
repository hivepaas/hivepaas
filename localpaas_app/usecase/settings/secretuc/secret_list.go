package secretuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) ListSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.ListSecretReq,
) (*secretdto.ListSecretResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := secretdto.TransformSecrets(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &secretdto.ListSecretResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
