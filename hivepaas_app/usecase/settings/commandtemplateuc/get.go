package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) GetCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.GetCommandTemplateReq,
) (*commandtemplatedto.GetCommandTemplateResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	setting := resp.Data
	respData, err := commandtemplatedto.TransformCommandTemplate(setting, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &commandtemplatedto.GetCommandTemplateResp{
		Data: respData,
	}, nil
}
