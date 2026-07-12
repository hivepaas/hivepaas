package imserviceuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
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

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		if err := setting.MustAsIMService().Decrypt(); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	respData, err := imservicedto.TransformIMService(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imservicedto.GetIMServiceResp{
		Data: respData,
	}, nil
}
