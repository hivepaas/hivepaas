package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) GetCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.GetCommandTemplateReq,
) (*commandtemplatedto.GetCommandTemplateResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	respData, err := commandtemplatedto.TransformCommandTemplate(setting, resp.RefObjects, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &commandtemplatedto.GetCommandTemplateResp{
		Data: respData,
	}, nil
}
