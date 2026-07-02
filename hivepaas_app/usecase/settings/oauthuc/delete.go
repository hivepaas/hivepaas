package oauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) DeleteOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.DeleteOAuthReq,
) (*oauthdto.DeleteOAuthResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &oauthdto.DeleteOAuthResp{}, nil
}
