package cloudstorageuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) UpdateCloudStorageStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.UpdateCloudStorageStatusReq,
) (*cloudstoragedto.UpdateCloudStorageStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.UpdateCloudStorageStatusResp{}, nil
}
