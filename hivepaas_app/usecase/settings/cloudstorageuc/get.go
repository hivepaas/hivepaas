package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) GetCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.GetCloudStorageReq,
) (*cloudstoragedto.GetCloudStorageResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := cloudstoragedto.TransformCloudStorage(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &cloudstoragedto.GetCloudStorageResp{
		Data: respData,
	}, nil
}
