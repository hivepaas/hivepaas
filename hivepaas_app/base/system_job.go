package base

type SystemJobName string

const (
	SystemJobNameDataBackup  SystemJobName = "data-backup"
	SystemJobNameDataCleanup SystemJobName = "data-cleanup"
	SystemJobNameSslRenewal  SystemJobName = "ssl-renewal"
)

var (
	AllSystemJobNames = []SystemJobName{SystemJobNameDataBackup, SystemJobNameDataCleanup, SystemJobNameSslRenewal}
)
