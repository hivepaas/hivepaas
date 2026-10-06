package syscleanupservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SysCleanupReq struct {
	*queue.TaskExecData
	Scope              *entity.ObjectScope
	SysCleanupSettings *entity.SystemCleanup

	CleanupClusterContainers base.CleanupFlag
	CleanupClusterImages     base.CleanupFlag
	CleanupClusterVolumes    base.CleanupFlag
	CleanupClusterNetworks   base.CleanupFlag
	CleanupClusterBuildCache base.CleanupFlag

	CleanupCacheRepo base.CleanupFlag

	CleanupFilesTemp base.CleanupFlag

	SyncSystemApps base.CleanupFlag
}

func (req *SysCleanupReq) SetCleanupFlagsDefault() {
	req.CleanupClusterContainers = base.CleanupFlagTrue
	req.CleanupClusterImages = base.CleanupFlagTrue
	req.CleanupClusterVolumes = base.CleanupFlagTrue
	req.CleanupClusterNetworks = base.CleanupFlagTrue
	req.CleanupClusterBuildCache = base.CleanupFlagTrue

	req.CleanupCacheRepo = base.CleanupFlagTrue

	req.CleanupFilesTemp = base.CleanupFlagTrue

	req.SyncSystemApps = base.CleanupFlagTrue
}

type SysCleanupResp struct {
	TaskOutput             *entity.TaskSystemCleanupOutput
	SkipResultNotification bool
}
