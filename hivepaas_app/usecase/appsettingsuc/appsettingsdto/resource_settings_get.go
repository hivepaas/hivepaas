package appsettingsdto

import (
	"strconv"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

type GetAppResourceSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppResourceSettingsReq() *GetAppResourceSettingsReq {
	return &GetAppResourceSettingsReq{}
}

func (req *GetAppResourceSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppResourceSettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *ResourceSettingsResp `json:"data"`
}

type ResourceSettingsResp struct {
	Reservations *ResourceReservations `json:"reservations"`
	Limits       *ResourceLimits       `json:"limits"`
	Memory       *Memory               `json:"memory"`
	Capabilities *Capabilities         `json:"capabilities"`

	UpdateVer int `json:"updateVer"`
}

type ResourceReservations struct {
	CPUs             float64            `json:"cpus,omitempty"`
	Memory           unit.DataSize      `json:"memory,omitempty"`
	GenericResources []*GenericResource `json:"genericResources,omitempty"`
}

type GenericResource struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ResourceLimits struct {
	CPUs   float64       `json:"cpus,omitempty"`
	Memory unit.DataSize `json:"memory,omitempty"`
	Pids   int64         `json:"pids,omitempty"`
}

type Memory struct {
	Swap       unit.DataSize `json:"swap"`
	Swappiness *int64        `json:"swappiness"`
	ShmSize    unit.DataSize `json:"shmSize"`
}

type Capabilities struct {
	Ulimits        []*Ulimit         `json:"ulimits"`
	CapabilityAdd  []string          `json:"capabilityAdd,omitempty"`
	CapabilityDrop []string          `json:"capabilityDrop,omitempty"`
	EnableGPU      bool              `json:"enableGPU,omitempty"`
	OomScoreAdj    int64             `json:"oomScoreAdj,omitempty"`
	Sysctls        map[string]string `json:"sysctls,omitempty"`
}

func (c *Capabilities) Equal(other *Capabilities) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if c.EnableGPU != other.EnableGPU {
		return false
	}
	if c.OomScoreAdj != other.OomScoreAdj {
		return false
	}
	if !gofn.ContentEqualPtr(c.Ulimits, other.Ulimits) {
		return false
	}
	if !gofn.ContentEqual(c.CapabilityAdd, other.CapabilityAdd) {
		return false
	}
	if !gofn.ContentEqual(c.CapabilityDrop, other.CapabilityDrop) {
		return false
	}
	if !gofn.MapEqual(c.Sysctls, other.Sysctls) {
		return false
	}
	return true
}

type Ulimit struct {
	Name string
	Hard int64
	Soft int64
}

func TransformResourceSettings(
	service *swarm.Service,
) (resp *ResourceSettingsResp, err error) {
	spec := &service.Spec
	resp = &ResourceSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.Reservations = TransformResourceReservations(spec.TaskTemplate.Resources)
	resp.Limits = TransformResourceLimits(spec.TaskTemplate.Resources)
	resp.Memory = TransformMemory(&spec.TaskTemplate)
	resp.Capabilities = TransformCapabilities(spec.TaskTemplate.ContainerSpec)

	return resp, nil
}

func TransformResourceReservations(res *swarm.ResourceRequirements) *ResourceReservations {
	if res == nil || res.Reservations == nil {
		return nil
	}
	resp := &ResourceReservations{
		CPUs:             float64(res.Reservations.NanoCPUs) / docker.UnitCPUNano,
		Memory:           unit.DataSize(res.Reservations.MemoryBytes),
		GenericResources: make([]*GenericResource, 0, len(res.Reservations.GenericResources)),
	}
	for _, r := range res.Reservations.GenericResources {
		if r.NamedResourceSpec != nil {
			resp.GenericResources = append(resp.GenericResources, &GenericResource{
				Kind:  r.NamedResourceSpec.Kind,
				Value: r.NamedResourceSpec.Value,
			})
		}
		if r.DiscreteResourceSpec != nil {
			resp.GenericResources = append(resp.GenericResources, &GenericResource{
				Kind:  r.DiscreteResourceSpec.Kind,
				Value: strconv.FormatInt(r.DiscreteResourceSpec.Value, 10),
			})
		}
	}
	return resp
}

func TransformResourceLimits(res *swarm.ResourceRequirements) *ResourceLimits {
	if res == nil || res.Limits == nil {
		return nil
	}
	return &ResourceLimits{
		CPUs:   float64(res.Limits.NanoCPUs) / docker.UnitCPUNano,
		Memory: unit.DataSize(res.Limits.MemoryBytes),
		Pids:   res.Limits.Pids,
	}
}

func TransformMemory(taskSpec *swarm.TaskSpec) *Memory {
	if taskSpec == nil {
		return nil
	}
	resp := &Memory{}
	// Resource requirements
	if taskSpec.Resources != nil {
		if taskSpec.Resources.SwapBytes != nil {
			resp.Swappiness = taskSpec.Resources.MemorySwappiness
		}
		if taskSpec.Resources.SwapBytes != nil {
			resp.Swap = unit.DataSize(*taskSpec.Resources.SwapBytes)
		}
	}
	// Shm size (extract from Tmpfs mount)
	shmMount := dockerhelper.GetShmMount(taskSpec)
	if shmMount != nil && shmMount.TmpfsOptions != nil {
		resp.ShmSize = unit.DataSize(shmMount.TmpfsOptions.SizeBytes)
	}
	return resp
}

func TransformCapabilities(containerSpec *swarm.ContainerSpec) *Capabilities {
	return &Capabilities{
		Ulimits: gofn.MapSlice(containerSpec.Ulimits, func(ulimit *container.Ulimit) *Ulimit {
			return &Ulimit{
				Name: ulimit.Name,
				Hard: ulimit.Hard,
				Soft: ulimit.Soft,
			}
		}),
		CapabilityAdd:  containerSpec.CapabilityAdd,
		CapabilityDrop: containerSpec.CapabilityDrop,
		EnableGPU:      gofn.Contain(containerSpec.CapabilityAdd, "[gpu]"),
		OomScoreAdj:    containerSpec.OomScoreAdj,
		Sysctls:        containerSpec.Sysctls,
	}
}
