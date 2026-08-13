package base

type SystemEventType string

const (
	SystemEventPeriodicSettingsReload SystemEventType = "periodic-settings:reload"
	SystemEventHivepaasDomainReload   SystemEventType = "hivepaas-domain:reload"
)
