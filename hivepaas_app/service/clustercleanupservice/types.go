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
