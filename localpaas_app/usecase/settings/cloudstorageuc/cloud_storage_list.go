package cloudstorageuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) ListCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.ListCloudStorageReq,
) (*cloudstoragedto.ListCloudStorageResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := cloudstoragedto.TransformCloudStorages(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.ListCloudStorageResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
