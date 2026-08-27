package volumeuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) DeleteVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.DeleteVolumeReq,
) (*volumedto.DeleteVolumeResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
		) error {
			if data.Setting.ObjectID == req.Scope.ScopeObjectID() {
				volEnt := data.Setting.MustAsClusterVolume()
				_, err := uc.dockerManager.VolumeRemove(ctx, volEnt.RefID, true)
				if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
					return hperrors.Wrap(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.DeleteVolumeResp{}, nil
}
