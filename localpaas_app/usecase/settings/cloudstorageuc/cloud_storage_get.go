package cloudstorageuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) GetCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.GetCloudStorageReq,
) (*cloudstoragedto.GetCloudStorageResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsCloudStorage().MustDecrypt()
	respData, err := cloudstoragedto.TransformCloudStorage(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.GetCloudStorageResp{
		Data: respData,
	}, nil
}
