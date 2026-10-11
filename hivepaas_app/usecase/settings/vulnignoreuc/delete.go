package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) DeleteVulnIgnore(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.DeleteVulnIgnoreReq,
) (*vulnignoredto.DeleteVulnIgnoreResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.DeleteVulnIgnoreResp{}, nil
}
