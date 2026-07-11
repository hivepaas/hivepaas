package networkuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
				_, err := uc.dockerManager.NetworkRemove(ctx, dockerhelper.ParseID(data.Setting.ID))
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
