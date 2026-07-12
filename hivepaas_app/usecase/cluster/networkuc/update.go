package networkuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) UpdateNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.UpdateNetworkReq,
) (*networkdto.UpdateNetworkResp, error) {
	req.Type = currentSettingType
	// NOTE: only allow updating `availInProjects` and `default`
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &networkdto.UpdateNetworkResp{}, nil
}
