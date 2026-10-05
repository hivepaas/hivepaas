package dockerhelper

import (
	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/services/docker"
)

// GetVolumeID is a volume's id: its cluster id for a cluster volume, its name
// for any other. It is docker.VolumeID, which VolumeListByIDs matches on.
func GetVolumeID(vol *volume.Volume) string {
	return docker.VolumeID(vol)
}
