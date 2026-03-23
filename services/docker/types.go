package docker

const (
	UnitCPUNano = 1000 * 1000 * 1000
	UnitMemMB   = 1024 * 1024
)

type ServiceMode string

const (
	ServiceModeReplicated    ServiceMode = "replicated"
	ServiceModeReplicatedJob ServiceMode = "replicated-job"
	ServiceModeGlobal        ServiceMode = "global"
	ServiceModeGlobalJob     ServiceMode = "global-job"
)

type HealthcheckMode string

const (
	HealthcheckModeInherit  = HealthcheckMode("")
	HealthcheckModeNone     = HealthcheckMode("NONE")
	HealthcheckModeCmd      = HealthcheckMode("CMD")
	HealthcheckModeCmdShell = HealthcheckMode("CMD-SHELL")
)
