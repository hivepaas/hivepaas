package dockerhelper

import (
	"strings"

	"github.com/moby/moby/api/types/volume"
)

func WrapNodeID(id string) string {
	if strings.HasPrefix(id, "dkr:") {
		return id
	}
	return "dkr:node:" + id
}

func WrapNetworkID(id string) string {
	if strings.HasPrefix(id, "dkr:") {
		return id
	}
	return "dkr:net:" + id
}

func WrapVolumeID(id string) string {
	if strings.HasPrefix(id, "dkr:") {
		return id
	}
	return "dkr:vol:" + id
}

func ParseID(wrapID string) string {
	wrapID = strings.TrimPrefix(wrapID, "dkr:")
	before, after, found := strings.Cut(wrapID, ":")
	if !found {
		return wrapID
	}
	switch before {
	case "node", "net", "vol":
		return after
	default:
		return wrapID
	}
}

func GetVolumeID(vol *volume.Volume) string {
	if vol == nil {
		return ""
	}
	if vol.ClusterVolume == nil {
		return vol.Name
	}
	return vol.ClusterVolume.ID
}
