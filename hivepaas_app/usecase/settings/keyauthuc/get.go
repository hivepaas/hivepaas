package keyauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) GetKeyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.GetKeyAuthReq,
) (*keyauthdto.GetKeyAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := keyauthdto.TransformKeyAuth(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.GetKeyAuthResp{
		Data: respData,
	}, nil
}
