package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) DeleteCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.DeleteCloudStorageReq,
) (*cloudstoragedto.DeleteCloudStorageResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &cloudstoragedto.DeleteCloudStorageResp{}, nil
}
