package basicauthuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) ListBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.ListBasicAuthReq,
) (*basicauthdto.ListBasicAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := basicauthdto.TransformBasicAuths(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &basicauthdto.ListBasicAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
