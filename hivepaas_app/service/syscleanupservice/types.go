package syscleanupservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SysCleanupReq struct {
	*queue.TaskExecData
	SysCleanupSettings *entity.SystemCleanup

	CleanupClusterContainers base.CleanupFlag
	CleanupClusterImages     base.CleanupFlag
	CleanupClusterVolumes    base.CleanupFlag
	CleanupClusterNetworks   base.CleanupFlag
	CleanupClusterBuildCache base.CleanupFlag

	CleanupBackupInLocal base.CleanupFlag
	CleanupBackupInCloud base.CleanupFlag

	CleanupCacheRepo base.CleanupFlag

	CleanupFilesTemp base.CleanupFlag
}

func (req *SysCleanupReq) SetCleanupFlagsDefault() {
	req.CleanupClusterContainers = base.CleanupFlagTrue
	req.CleanupClusterImages = base.CleanupFlagTrue
	req.CleanupClusterVolumes = base.CleanupFlagTrue
	req.CleanupClusterNetworks = base.CleanupFlagTrue
	req.CleanupClusterBuildCache = base.CleanupFlagTrue

	req.CleanupBackupInLocal = base.CleanupFlagTrue
	req.CleanupBackupInCloud = base.CleanupFlagTrue

	req.CleanupCacheRepo = base.CleanupFlagTrue

	req.CleanupFilesTemp = base.CleanupFlagTrue
}

type SysCleanupResp struct {
	TaskOutput             *entity.TaskSystemCleanupOutput
	SkipResultNotification bool
}
