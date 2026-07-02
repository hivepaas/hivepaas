package networkuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

func (uc *UC) UpdateNetworkStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.UpdateNetworkStatusReq,
) (*networkdto.UpdateNetworkStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.UpdateNetworkStatusResp{}, nil
}
