package accesstokenuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

func (uc *UC) UpdateAccessTokenStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.UpdateAccessTokenStatusReq,
) (*accesstokendto.UpdateAccessTokenStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &accesstokendto.UpdateAccessTokenStatusResp{}, nil
}
