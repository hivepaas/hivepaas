package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) UpdateVulnIgnoreStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.UpdateVulnIgnoreStatusReq,
) (*vulnignoredto.UpdateVulnIgnoreStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.UpdateVulnIgnoreStatusResp{}, nil
}
