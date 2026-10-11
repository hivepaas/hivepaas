package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) GetVulnIgnore(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.GetVulnIgnoreReq,
) (*vulnignoredto.GetVulnIgnoreResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := vulnignoredto.TransformVulnIgnore(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.GetVulnIgnoreResp{
		Data: respData,
	}, nil
}
