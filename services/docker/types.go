package docker

import "strings"

const (
	UnitCPUNano    = 1000 * 1000 * 1000
	MinCPUFraction = 0.25
	// GenericResourceGPU is the kind NVIDIA's GPUs are listed as by a node, and
	// reserved as by an app: swarm places the app on a node with one free and
	// names it to the container in DOCKER_RESOURCE_NVIDIA-GPU, which NVIDIA's
	// container runtime reads to hand the container that GPU. It is the kind
	// enableGPU, and a Compose file's `gpus`, reserve. AMD's are AMD_GPU.
	GenericResourceGPU = "NVIDIA-GPU"
	// CapabilityGPU is how Enable GPU used to ask for the GPU: in the
	// capability list, where docker refuses it ("unknown capability"). It is
	// read as one NVIDIA-GPU reserved, and taken away when the app is saved.
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

// IsGPUKind says whether a kind of generic resource is a GPU: one with GPU in
// its name - NVIDIA-GPU, AMD_GPU, or whatever a node lists its GPUs as.
func IsGPUKind(kind string) bool {
	return strings.Contains(strings.ToUpper(kind), "GPU")
}
