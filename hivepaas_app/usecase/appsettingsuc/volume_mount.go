package appsettingsuc

import (
	"path/filepath"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
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

// applyVolumeDriverConfigUnlessOverridden fills in the volume's driver config
// unless the request already carries one. An explicit DriverConfig on the
// mount came from the caller overriding, and filling it in from the setting
// anyway would silently discard that override.
func applyVolumeDriverConfigUnlessOverridden(dockerMnt *mount.Mount, vol *entity.ClusterVolume) {
	if dockerMnt.VolumeOptions == nil || dockerMnt.VolumeOptions.DriverConfig == nil {
		applyVolumeDriverConfig(dockerMnt, vol)
	}
}

// refuseConflictingVolumePins reports whether the final mount set pins the
// service to more than one node.
//
// A pin becomes a required swarm constraint (node.id==X), unlike every other
// constraint HivePaaS emits, which only excludes candidates. Two volumes
// pinned to different nodes therefore describe a placement no node can
// satisfy - and placementserviceimpl responds to that by emitting no
// constraint at all, silently turning a pinned service into an unconstrained
// one. This is the last point before the mounts are saved where that
// contradiction can still be caught, so it is refused here instead.
func refuseConflictingVolumePins(mounts []mount.Mount, volumes []*entity.Setting) error {
	pins, err := placementservice.VolumePinsForMounts(mounts, volumes)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, conflict := placementservice.VolumePinConstraint(pins); conflict != nil {
		// conflict.Error() names both volumes and the nodes they are pinned to;
		// WithExtraDetail is what actually carries that to the caller - WithMsgLog
		// only reaches the server log (see BaseHandler.RenderError, which strips
		// DebugLog outside dev but always sends Detail/extraDetail).
		return hperrors.NewArgumentInvalid("Mounts").WithExtraDetail("%s", conflict.Error())
	}
	return nil
}
