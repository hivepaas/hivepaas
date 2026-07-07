package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) ListCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.ListCommandTemplateReq,
) (*commandtemplatedto.ListCommandTemplateResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	respData, err := commandtemplatedto.TransformCommandTemplates(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &commandtemplatedto.ListCommandTemplateResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
