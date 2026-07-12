package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
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
