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

	CleanupDB bool

	CleanupCluster           bool
	CleanupClusterContainers CleanupFlag
	CleanupClusterImages     CleanupFlag
	CleanupClusterVolumes    CleanupFlag
	CleanupClusterNetworks   CleanupFlag

	CleanupBackup        bool
	CleanupBackupInLocal CleanupFlag
	CleanupBackupInCloud CleanupFlag

	CleanupCache     bool
	CleanupCacheRepo CleanupFlag

	CleanupFiles     bool
	CleanupFilesTemp CleanupFlag
}

func (req *SysCleanupReq) SetCleanupAll() {
	req.CleanupDB = true

	req.CleanupCluster = true
	req.CleanupClusterContainers = CleanupFlagTrue
	req.CleanupClusterImages = CleanupFlagTrue
	req.CleanupClusterVolumes = CleanupFlagTrue
	req.CleanupClusterNetworks = CleanupFlagTrue

	req.CleanupBackup = true
	req.CleanupBackupInLocal = CleanupFlagTrue
	req.CleanupBackupInCloud = CleanupFlagTrue

	req.CleanupCache = true
	req.CleanupCacheRepo = CleanupFlagTrue

	req.CleanupFiles = true
	req.CleanupFilesTemp = CleanupFlagTrue
}

type SysCleanupResp struct {
	TaskOutput             *entity.TaskSystemCleanupOutput
	SkipResultNotification bool
}
