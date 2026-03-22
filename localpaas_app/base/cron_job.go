package base

type CronJobType string

const (
	CronJobTypeContainerCommand CronJobType = "container-command"
	CronJobTypeSystemCleanup    CronJobType = "system-cleanup"
	CronJobTypeSSLRenewal       CronJobType = "ssl-renewal"
)

var (
	AllCronJobTypes = []CronJobType{CronJobTypeContainerCommand, CronJobTypeSystemCleanup,
		CronJobTypeSSLRenewal}
)
