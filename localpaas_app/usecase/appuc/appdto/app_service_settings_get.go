package appdto

import (
	"strings"

	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/services/docker"
)

type GetAppServiceSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppServiceSettingsReq() *GetAppServiceSettingsReq {
	return &GetAppServiceSettingsReq{}
}

func (req *GetAppServiceSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppServiceSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *ServiceSettingsResp `json:"data"`
}

type ServiceSettingsResp struct {
	ModeSpec  *ServiceModeSpec `json:"modeSpec"`
	Placement *Placement       `json:"placement,omitempty"`

	UpdateVer int `json:"updateVer"`
}

type ServiceModeSpec struct {
	Mode                docker.ServiceMode `json:"mode,omitempty"`
	ServiceReplicas     *uint64            `json:"serviceReplicas,omitempty"`
	JobMaxConcurrent    *uint64            `json:"jobMaxConcurrent,omitempty"`
	JobTotalCompletions *uint64            `json:"jobTotalCompletions,omitempty"`
}

type Placement struct {
	Constraints []*PlacementConstraint `json:"constraints,omitempty"`
	Preferences []*PlacementPreference `json:"preferences,omitempty"`
}

type PlacementConstraint struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Op    string `json:"op"`
}

type PlacementPreference struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TransformServiceSettings(
	service *swarm.Service,
) (resp *ServiceSettingsResp, err error) {
	spec := &service.Spec
	resp = &ServiceSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.ModeSpec = TransformServiceMode(spec)
	resp.Placement = TransformServicePlacement(spec.TaskTemplate.Placement)

	return resp, nil
}

func TransformServiceMode(spec *swarm.ServiceSpec) *ServiceModeSpec {
	if spec == nil {
		return nil
	}
	res := &ServiceModeSpec{}
	switch {
	case spec.Mode.Replicated != nil:
		res.Mode = docker.ServiceModeReplicated
		res.ServiceReplicas = spec.Mode.Replicated.Replicas
	case spec.Mode.ReplicatedJob != nil:
		res.Mode = docker.ServiceModeReplicatedJob
		res.JobMaxConcurrent = spec.Mode.ReplicatedJob.MaxConcurrent
		res.JobTotalCompletions = spec.Mode.ReplicatedJob.TotalCompletions
	case spec.Mode.Global != nil:
		res.Mode = docker.ServiceModeGlobal
	case spec.Mode.GlobalJob != nil:
		res.Mode = docker.ServiceModeGlobalJob
	}
	return res
}

func TransformServicePlacement(placement *swarm.Placement) *Placement {
	if placement == nil {
		return nil
	}
	res := &Placement{
		Constraints: make([]*PlacementConstraint, 0, len(placement.Constraints)),
		Preferences: make([]*PlacementPreference, 0, len(placement.Preferences)),
	}
	for _, constraint := range placement.Constraints {
		op := "=="
		name, value, found := strings.Cut(constraint, op)
		if !found {
			op = "!="
			name, value, found = strings.Cut(constraint, op)
			if !found {
				name = constraint
				value = ""
				op = ""
			}
		}
		res.Constraints = append(res.Constraints, &PlacementConstraint{
			Name:  name,
			Value: value,
			Op:    op,
		})
	}
	for _, pref := range placement.Preferences {
		if pref.Spread != nil {
			res.Preferences = append(res.Preferences, &PlacementPreference{
				Name:  "spread",
				Value: pref.Spread.SpreadDescriptor,
			})
		}
	}
	return res
}
