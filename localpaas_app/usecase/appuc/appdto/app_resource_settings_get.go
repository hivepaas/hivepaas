package appdto

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppResourceSettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *ResourceSettingsResp `json:"data"`
}

type ResourceSettingsResp struct {
	Reservations *ResourceReservations `json:"reservations"`
	Limits       *ResourceLimits       `json:"limits"`
	Ulimits      []*Ulimit             `json:"ulimits"`
	Capabilities *Capabilities         `json:"capabilities"`

	UpdateVer int `json:"updateVer"`
}

type ResourceReservations struct {
	CPUs             float64            `json:"cpus,omitempty"`
	MemoryMB         int64              `json:"memoryMB,omitempty"`
	GenericResources []*GenericResource `json:"genericResources,omitempty"`
}

type GenericResource struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ResourceLimits struct {
	CPUs     float64 `json:"cpus,omitempty"`
	MemoryMB int64   `json:"memoryMB,omitempty"`
	Pids     int64   `json:"pids,omitempty"`
}

type Ulimit struct {
	Name string
	Hard int64
	Soft int64
}

type Capabilities struct {
	CapabilityAdd  []string          `json:"capabilityAdd,omitempty"`
	CapabilityDrop []string          `json:"capabilityDrop,omitempty"`
	EnableGPU      bool              `json:"enableGPU,omitempty"`
	OomScoreAdj    int64             `json:"oomScoreAdj,omitempty"`
	Sysctls        map[string]string `json:"sysctls,omitempty"`
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
	resp.Ulimits = TransformUlimits(spec.TaskTemplate.ContainerSpec.Ulimits)
	resp.Capabilities = TransformCapabilities(spec.TaskTemplate.ContainerSpec)

	return resp, nil
}

func TransformResourceReservations(res *swarm.ResourceRequirements) *ResourceReservations {
	if res == nil || res.Reservations == nil {
		return nil
	}
	resp := &ResourceReservations{
		CPUs:             float64(res.Reservations.NanoCPUs / docker.UnitCPUNano),
		MemoryMB:         res.Reservations.MemoryBytes / docker.UnitMemMB,
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
		CPUs:     float64(res.Limits.NanoCPUs / docker.UnitCPUNano),
		MemoryMB: res.Limits.MemoryBytes / docker.UnitMemMB,
		Pids:     res.Limits.Pids,
	}
}

func TransformUlimits(ulimits []*container.Ulimit) []*Ulimit {
	resp := make([]*Ulimit, 0, len(ulimits))
	for _, ulimit := range ulimits {
		resp = append(resp, &Ulimit{
			Name: ulimit.Name,
			Hard: ulimit.Hard,
			Soft: ulimit.Soft,
		})
	}
	return resp
}

func TransformCapabilities(containerSpec *swarm.ContainerSpec) *Capabilities {
	return &Capabilities{
		CapabilityAdd:  containerSpec.CapabilityAdd,
		CapabilityDrop: containerSpec.CapabilityDrop,
		EnableGPU:      gofn.Contain(containerSpec.CapabilityAdd, "[gpu]"),
		OomScoreAdj:    containerSpec.OomScoreAdj,
		Sysctls:        containerSpec.Sysctls,
	}
}
