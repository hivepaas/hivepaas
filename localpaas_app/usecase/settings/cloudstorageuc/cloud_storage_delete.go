package cloudstorageuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) DeleteCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.DeleteCloudStorageReq,
) (*cloudstoragedto.DeleteCloudStorageResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.DeleteCloudStorageResp{}, nil
}
