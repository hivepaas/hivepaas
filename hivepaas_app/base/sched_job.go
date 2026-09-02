package base

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"

type SchedJobType string

const (
	SchedJobTypeContainerCommand  SchedJobType = "container-command"
	SchedJobTypeSystemCleanup     SchedJobType = "system-cleanup"
	SchedJobTypeSystemBackup      SchedJobType = "system-backup"
	SchedJobTypeSSLRenewal        SchedJobType = "ssl-renewal"
	SchedJobTypeBackupRepoCleanup SchedJobType = "backup-repo-cleanup"
)

var (
	AllSchedJobTypes = []SchedJobType{SchedJobTypeContainerCommand, SchedJobTypeSystemCleanup,
		SchedJobTypeSystemBackup, SchedJobTypeSSLRenewal, SchedJobTypeBackupRepoCleanup}
)

const (
	ExecCommandMaxSize = 300 * unit.KB
)
