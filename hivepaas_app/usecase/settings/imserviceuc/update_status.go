package imserviceuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) UpdateIMServiceStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.UpdateIMServiceStatusReq,
) (*imservicedto.UpdateIMServiceStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &imservicedto.UpdateIMServiceStatusResp{}, nil
}
