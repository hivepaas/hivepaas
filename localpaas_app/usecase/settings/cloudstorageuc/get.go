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

	setting := resp.Data
	if setting.ObjectID == setting.CurrentObjectID { // not return sensitive data if setting is inherited
		setting.MustAsCloudStorage().MustDecrypt()
	}

	respData, err := cloudstoragedto.TransformCloudStorage(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.GetCloudStorageResp{
		Data: respData,
	}, nil
}
