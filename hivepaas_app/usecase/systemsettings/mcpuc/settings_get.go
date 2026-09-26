package mcpuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/mcpuc/mcpdto"
)

func (uc *UC) GetMCPSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *mcpdto.GetMCPSettingsReq,
) (*mcpdto.GetMCPSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSettingOrEmpty(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := mcpdto.TransformMCPSettings(resp.Data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &mcpdto.GetMCPSettingsResp{Data: respData}, nil
}
