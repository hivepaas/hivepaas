package oauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) UpdateOAuthStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.UpdateOAuthStatusReq,
) (*oauthdto.UpdateOAuthStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &oauthdto.UpdateOAuthStatusResp{}, nil
}
