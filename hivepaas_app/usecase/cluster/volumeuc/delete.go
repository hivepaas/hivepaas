package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

func (uc *UC) DeleteVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.DeleteVolumeReq,
) (*volumedto.DeleteVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
		) error {
			if data.Setting.ObjectID == req.Scope.ScopeObjectID() {
				// Retrying rather than removing once: deleting a volume
				// normally follows removing whatever was using it, and swarm
				// tears task containers down in the background, so the first
				// attempt can lose that race by a second or two. RemoveVolume
				// treats an already-missing volume as success, which is the
				// common case now that volumes materialize lazily on whichever
				// node runs the task rather than here.
				err := uc.volumeService.RemoveVolume(ctx, data.Setting.RefID, true,
					volumeservice.VolumeRemovalRetryMax, 0)
				if err != nil {
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
