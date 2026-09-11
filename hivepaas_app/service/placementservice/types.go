package placementservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ApplyPlacementSettingsReq struct {
	App                *entity.App
	Service            *swarm.Service               // If nil, it will be loaded
	PlacementSettings  *entity.AppPlacementSettings // If nil, it will be loaded
	BuildSettings      *entity.ImageBuildSettings   // If nil, it will be loaded
	SkipSavingToDocker bool

	// VolumePins are the pins of the volumes this app mounts. Nil means they are
	// loaded from the app's mounts.
	VolumePins []VolumePin
}

type ApplyPlacementSettingsResp struct {
	Service    *swarm.Service
	HasChanges bool
}
