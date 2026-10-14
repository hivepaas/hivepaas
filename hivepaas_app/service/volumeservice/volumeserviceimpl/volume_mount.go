package volumeserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) BuildVolumeMount(
	ctx context.Context,
	db database.IDB,
	volumeID string,
	target string,
) (*volumeservice.VolumeMount, error) {
	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeClusterVolume,
		volumeID, true)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.NewNotFound("Volume").WithMsgLog("volume %v not found", volumeID)
		}
		return nil, hperrors.Wrap(err)
	}
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if volumeservice.DriverOptsReachDockerSocket(vol.DriverOpts) {
		return nil, hperrors.Wrap(hperrors.ErrArgumentInvalid).
			WithExtraDetail("volume %s reaches the Docker socket of the node.", setting.Name)
	}
	return &volumeservice.VolumeMount{Setting: setting, Volume: vol, Mount: wholeVolumeMount(setting, vol, target)}, nil
}

// wholeVolumeMount is a mount of a whole volume: a bind of its directory for a
// local bind volume, which creates the directory where it is missing, as an
// app's mount does; the named volume otherwise, carrying its driver and
// options for a managed one, so the node builds it when it does not have it.
func wholeVolumeMount(setting *entity.Setting, vol *entity.ClusterVolume, target string) mount.Mount {
	if directory, propagation, ok := bindMountTarget(vol, ""); ok {
		mnt := mount.Mount{
			Type:        mount.TypeBind,
			Source:      directory,
			Target:      target,
			BindOptions: &mount.BindOptions{CreateMountpoint: true},
		}
		if propagation != "" {
			mnt.BindOptions.Propagation = propagation
		}
		return mnt
	}
	mnt := mount.Mount{Type: mount.TypeVolume, Source: setting.RefID, Target: target}
	applyVolumeDriverConfig(&mnt, vol)
	return mnt
}
