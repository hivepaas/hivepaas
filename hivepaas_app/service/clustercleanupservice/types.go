package clustercleanupservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type ClusterCleanupReq struct {
	*queue.TaskExecData
	CleanupSettings *entity.SystemClusterCleanup

	CleanupContainers base.CleanupFlag
	CleanupImages     base.CleanupFlag
	CleanupVolumes    base.CleanupFlag
	CleanupNetworks   base.CleanupFlag
	CleanupBuildCache base.CleanupFlag

	// KeepImages are the images every node keeps however long it has not used
	// them, from KeptImages. None, from a manager of an older version, prunes
	// only the images no tag names, as that version did.
	KeepImages []string
	// KeepArtifacts are the artifacts every node keeps, from KeptArtifacts, when
	// ArtifactsKnown. A manager of an older version knows nothing of them, and a
	// node it asks removes none.
	KeepArtifacts  []string
	ArtifactsKnown bool
}

func (req *ClusterCleanupReq) SetCleanupFlagsDefault() {
	req.CleanupContainers = base.CleanupFlagTrue
	req.CleanupImages = base.CleanupFlagTrue
	req.CleanupVolumes = base.CleanupFlagTrue
	req.CleanupNetworks = base.CleanupFlagTrue
	req.CleanupBuildCache = base.CleanupFlagTrue
}

type ClusterCleanupResp struct {
	Output *entity.ClusterNodeCleanupOutput
}
