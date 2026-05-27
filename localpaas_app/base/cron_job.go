package base

type SchedJobType string

const (
	SchedJobTypeContainerCommand SchedJobType = "container-command"
	SchedJobTypeSystemCleanup    SchedJobType = "system-cleanup"
	SchedJobTypeSystemBackup     SchedJobType = "system-backup"
	SchedJobTypeSSLRenewal       SchedJobType = "ssl-renewal"
)

var (
	AllSchedJobTypes = []SchedJobType{SchedJobTypeContainerCommand, SchedJobTypeSystemCleanup,
		SchedJobTypeSystemBackup, SchedJobTypeSSLRenewal}
)
