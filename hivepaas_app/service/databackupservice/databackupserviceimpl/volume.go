package databackupserviceimpl

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) CheckAppVolume(ctx context.Context, db database.IDB, app *entity.App, volumeID string) error {
	_, err := s.appVolumeMount(ctx, db, app, volumeID)
	return hperrors.Wrap(err)
}

// appVolumeMount is the app's mount of its own directory in the volume, as its
// service has it.
func (s *service) appVolumeMount(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	volumeID string,
) (*mount.Mount, error) {
	swarmService, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	mounts := swarmService.Spec.TaskTemplate.ContainerSpec.Mounts
	descs, err := s.volumeService.DescribeAppMounts(ctx, db, app, mounts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	picked := pickAppVolumeMount(mounts, descs, volumeID)
	if picked == nil {
		return nil, hperrors.NewArgumentInvalid("dataBackup.sourceVolume").
			WithExtraDetail("the app does not mount this volume as its own directory")
	}
	return picked, nil
}

func (s *service) FindAppVolume(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	volumeID string,
) (*databackupservice.AppVolume, error) {
	picked, err := s.appVolumeMount(ctx, db, app, volumeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	err = s.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, nil, false,
		&entity.RefObjectIDs{RefSettingIDs: []string{volumeID}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	volume := refObjects.RefSettings[volumeID]
	if volume == nil {
		return nil, hperrors.NewNotFound("Volume")
	}
	volumeDir, err := s.backupRepoService.VolumeHostDir(ctx, volume)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &databackupservice.AppVolume{
		HostDir:   appVolumeDir(picked, volumeDir.Dir),
		NodeID:    volumeDir.NodeID,
		NodeLabel: volumeDir.NodeLabel,
	}, nil
}

// pickAppVolumeMount is the app's mount of its own directory in the volume; nil
// when it has none. A directory of another app it was given is not its data.
func pickAppVolumeMount(
	mounts []mount.Mount,
	descs []*volumeservice.AppMountDesc,
	volumeID string,
) *mount.Mount {
	for i, desc := range descs {
		if i < len(mounts) && desc != nil && desc.VolumeID == volumeID && desc.Own {
			return &mounts[i]
		}
	}
	return nil
}

// appVolumeDir is the app's part of the volume on the host: a bind mount names
// the directory itself, a volume mount a subpath of the volume's directory.
func appVolumeDir(mnt *mount.Mount, volumeDir string) string {
	if mnt.Type == mount.TypeBind {
		return mnt.Source
	}
	if mnt.VolumeOptions != nil && mnt.VolumeOptions.Subpath != "" {
		return filepath.Join(volumeDir, mnt.VolumeOptions.Subpath)
	}
	return volumeDir
}

// joinSubpath is subpath inside dir; an error for one that would leave it.
func joinSubpath(dir, subpath string) (string, error) {
	if subpath == "" {
		return dir, nil
	}
	cleaned := path.Clean(subpath)
	if strings.HasPrefix(subpath, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", hperrors.NewArgumentInvalid("dataBackup.sourceVolumeSubpath").
			WithExtraDetail("the subpath has to stay inside the app's part of the volume")
	}
	return filepath.Join(dir, cleaned), nil
}
