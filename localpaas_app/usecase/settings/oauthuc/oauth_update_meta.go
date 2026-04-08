package oauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) UpdateOAuthMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.UpdateOAuthMetaReq,
) (*oauthdto.UpdateOAuthMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingMeta(ctx, &req.UpdateSettingMetaReq, &settings.UpdateSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &oauthdto.UpdateOAuthMetaResp{}, nil
}
