package docker

const (
	UnitCPUNano    = 1000 * 1000 * 1000
	MinCPUFraction = 0.25
	// GenericResourceGPU is the kind of generic resource a node advertises its
	// GPUs as, and an app reserves one by: swarm places the app on a node with
	// one free and names it to the container in DOCKER_RESOURCE_NVIDIA-GPU,
	// which NVIDIA's container runtime reads to hand the container that GPU.
	GenericResourceGPU = "NVIDIA-GPU"
	// CapabilityGPU is how Enable GPU used to ask for the GPU: in the
	// capability list, where docker refuses it ("unknown capability"). It is
	// read to know such an app, and taken away when the app is saved.
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
