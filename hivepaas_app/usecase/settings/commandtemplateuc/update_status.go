package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) UpdateCommandTemplateStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.UpdateCommandTemplateStatusReq,
) (*commandtemplatedto.UpdateCommandTemplateStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &commandtemplatedto.UpdateCommandTemplateStatusResp{}, nil
}
