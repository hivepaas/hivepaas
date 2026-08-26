package dockerhelper

import (
	"github.com/moby/moby/api/types/volume"
)

func GetVolumeID(vol *volume.Volume) string {
	if vol == nil {
		return ""
	}
	if vol.ClusterVolume == nil {
		return vol.Name
	}
	return vol.ClusterVolume.ID
}
