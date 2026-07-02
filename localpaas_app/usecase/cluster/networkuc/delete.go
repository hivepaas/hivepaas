package networkuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

func (uc *UC) DeleteNetwork(
	ctx context.Context,
	auth *basedto.Auth,
	req *networkdto.DeleteNetworkReq,
) (*networkdto.DeleteNetworkResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
		) error {
			if data.Setting.ObjectID == req.Scope.MainObjectID() {
				netEntity, err := data.Setting.AsClusterNetwork()
				if err != nil {
					return apperrors.New(err)
				}
				_, err = uc.dockerManager.NetworkRemove(ctx, netEntity.NetworkID)
				if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
					return apperrors.New(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &networkdto.DeleteNetworkResp{}, nil
}
