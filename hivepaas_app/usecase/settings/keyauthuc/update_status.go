package keyauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) UpdateKeyAuthStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.UpdateKeyAuthStatusReq,
) (*keyauthdto.UpdateKeyAuthStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.UpdateKeyAuthStatusResp{}, nil
}
