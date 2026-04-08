package imserviceuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) GetIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.GetIMServiceReq,
) (*imservicedto.GetIMServiceResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsIMService().MustDecrypt()
	respData, err := imservicedto.TransformIMService(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imservicedto.GetIMServiceResp{
		Data: respData,
	}, nil
}
