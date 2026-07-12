package configfileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) GetConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.GetConfigFileReq,
) (*configfiledto.GetConfigFileResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Data.MustAsConfigFile()
	respData, err := configfiledto.TransformConfigFile(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &configfiledto.GetConfigFileResp{
		Data: respData,
	}, nil
}
