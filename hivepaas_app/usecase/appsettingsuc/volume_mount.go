package appsettingsuc

import (
	"path/filepath"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// bindMountTarget reports the directory a bind volume points at, so it can be
// mounted by path rather than by name.
//
// It reads the setting rather than the docker volume because the docker volume
// only exists on the node that happened to create it, which is the whole reason
// a volume's description lives in its setting now.
func bindMountTarget(
	vol *entity.ClusterVolume,
	subpath string,
) (directory string, propagation mount.Propagation, ok bool) {
	if vol == nil || vol.Driver != string(docker.VolumeDriverLocal) {
		return "", "", false
	}
	device := vol.DriverOpts["device"]
	if vol.DriverOpts["type"] != "none" || device == "" {
		return "", "", false
	}
	return filepath.Join(device, subpath), getConfiguredPropagation(vol.DriverOpts["o"]), true
}

// applyVolumeDriverConfig puts the volume's description into the mount, so the
// node running the task builds the volume from it instead of creating an empty
// default when the name is unknown there.
func applyVolumeDriverConfig(dockerMnt *mount.Mount, vol *entity.ClusterVolume) {
	if vol == nil || !vol.Managed || vol.Driver == "" {
		return
	}
	if dockerMnt.VolumeOptions == nil {
		dockerMnt.VolumeOptions = &mount.VolumeOptions{}
	}
	dockerMnt.VolumeOptions.DriverConfig = &mount.Driver{
		Name:    vol.Driver,
		Options: vol.DriverOpts,
	}
	if len(vol.Labels) > 0 && dockerMnt.VolumeOptions.Labels == nil {
		dockerMnt.VolumeOptions.Labels = vol.Labels
	}
}
