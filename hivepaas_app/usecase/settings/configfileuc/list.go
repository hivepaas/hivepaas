package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) ListConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.ListConfigFileReq,
) (*configfiledto.ListConfigFileResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := configfiledto.TransformConfigFiles(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &configfiledto.ListConfigFileResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
