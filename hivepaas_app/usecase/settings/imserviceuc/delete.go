package imserviceuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) DeleteIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.DeleteIMServiceReq,
) (*imservicedto.DeleteIMServiceResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.DeleteIMServiceResp{}, nil
}
