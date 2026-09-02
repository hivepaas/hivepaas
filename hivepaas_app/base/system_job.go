package base

type SystemJobName string

const (
	SystemJobNameDataBackup        SystemJobName = "data-backup"
	SystemJobNameDataCleanup       SystemJobName = "data-cleanup"
	SystemJobNameSslRenewal        SystemJobName = "ssl-renewal"
	SystemJobNameBackupRepoCleanup SystemJobName = "backup-repo-cleanup"
)

var (
	AllSystemJobNames = []SystemJobName{SystemJobNameDataBackup, SystemJobNameDataCleanup, SystemJobNameSslRenewal,
		SystemJobNameBackupRepoCleanup}
)
