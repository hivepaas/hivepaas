package imserviceuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) ListIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.ListIMServiceReq,
) (*imservicedto.ListIMServiceResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := imservicedto.TransformIMServices(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imservicedto.ListIMServiceResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
