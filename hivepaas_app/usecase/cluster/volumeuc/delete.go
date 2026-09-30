package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
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
			if err := refuseVolumeWithFiles(ctx, db, uc.FileService, data.Setting); err != nil {
				return hperrors.Wrap(err)
			}
			if data.Setting.ObjectID == req.Scope.ScopeObjectID() {
				// Retrying rather than removing once: deleting a volume
				// normally follows removing whatever was using it, and swarm
				// tears task containers down in the background, so the first
				// attempt can lose that race by a second or two. RemoveVolume
				// treats an already-missing volume as success, which is the
				// common case now that volumes materialize lazily on whichever
				// node runs the task rather than here.
				// The volume goes on whichever node holds it. What it was made
				// of stays: a volume made of a host directory keeps its data
				// when it is removed, and nothing here asked for that data to
				// be deleted.
				err := uc.volumeService.RemoveVolumeInCluster(ctx, data.Setting, false, true,
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

// refuseVolumeWithFiles refuses to delete a volume that still holds HivePaaS's
// files. Deleting them with it would remove each through the file layer.
func refuseVolumeWithFiles(
	ctx context.Context,
	db database.IDB,
	files fileservice.Service,
	volume *entity.Setting,
) error {
	count, err := files.CountOnVolume(ctx, db, volume.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count > 0 {
		return hperrors.Wrap(hperrors.ErrVolumeHasFiles).WithParam("Name", volume.Name).WithParam("Count", count)
	}
	return nil
}
