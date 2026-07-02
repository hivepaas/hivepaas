package basicauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
)

func (uc *UC) ListBasicAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *basicauthdto.ListBasicAuthReq,
) (*basicauthdto.ListBasicAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := basicauthdto.TransformBasicAuths(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &basicauthdto.ListBasicAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
