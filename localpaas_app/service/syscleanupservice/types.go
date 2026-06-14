package syscleanupservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type CleanupFlag int8

const (
	CleanupFlagFalse = iota
	CleanupFlagTrue
	CleanupFlagForce
)

type SysCleanupReq struct {
	*queue.TaskExecData
	SysCleanupSettings *entity.SystemCleanup

	CleanupClusterContainers CleanupFlag
	CleanupClusterImages     CleanupFlag
	CleanupClusterVolumes    CleanupFlag
	CleanupClusterNetworks   CleanupFlag
	CleanupClusterBuildCache CleanupFlag

	CleanupBackupInLocal CleanupFlag
	CleanupBackupInCloud CleanupFlag

	CleanupCacheRepo CleanupFlag

	CleanupFilesTemp CleanupFlag
}

func (req *SysCleanupReq) SetCleanupFlagsDefault() {
	req.CleanupClusterContainers = CleanupFlagTrue
	req.CleanupClusterImages = CleanupFlagTrue
	req.CleanupClusterVolumes = CleanupFlagTrue
	req.CleanupClusterNetworks = CleanupFlagTrue
	req.CleanupClusterBuildCache = CleanupFlagFalse

	req.CleanupBackupInLocal = CleanupFlagTrue
	req.CleanupBackupInCloud = CleanupFlagTrue

	req.CleanupCacheRepo = CleanupFlagTrue

	req.CleanupFilesTemp = CleanupFlagTrue
}

type SysCleanupResp struct {
	TaskOutput             *entity.TaskSystemCleanupOutput
	SkipResultNotification bool
}
