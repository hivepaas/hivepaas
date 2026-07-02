package accesstokenuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) DeleteAccessToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.DeleteAccessTokenReq,
) (*accesstokendto.DeleteAccessTokenResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &accesstokendto.DeleteAccessTokenResp{}, nil
}
