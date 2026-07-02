package volumeuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
			if data.Setting.ObjectID == req.Scope.MainObjectID() {
				volEntity, err := data.Setting.AsClusterVolume()
				if err != nil {
					return apperrors.New(err)
				}
				_, err = uc.dockerManager.VolumeRemove(ctx, volEntity.VolumeID, true)
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

	return &volumedto.DeleteVolumeResp{}, nil
}
