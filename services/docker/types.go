package docker

const (
	UnitCPUNano    = 1000 * 1000 * 1000
	MinCPUFraction = 0.25
	// CapabilityGPU is how a service spec asks for the GPU: docker carries the
	// request in the capability list, under a name no real capability has.
	CapabilityGPU = "[gpu]"
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

const (
	LabelTempResource    = "hivepaas.system.temp"
	LabelTempResourceVal = "true"
	LabelTempCreatedAt   = "hivepaas.system.createdAt"

	TempContainerPrefix = "hivepaas-cont-"
	TempServicePrefix   = "hivepaas-svc-"
)
