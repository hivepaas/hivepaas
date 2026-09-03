package imagebuildsettingsdto

import (
	"math"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	nodeLabelMaxLen = 100

	cpuMax = 1000
	memMax = math.MaxInt64
)

type UpdateImageBuildSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*ImageBuildSettingsBaseReq
}

type ImageBuildSettingsBaseReq struct {
	Workers   *ImageBuildWorkerSettingsReq   `json:"workers"`
	Resources *ImageBuildResourceSettingsReq `json:"resources"`
	Sources   *ImageBuildSourceSettingsReq   `json:"sources"`
	NoCache   bool                           `json:"noCache"`
	NoVerbose bool                           `json:"noVerbose"`
}

func (req *ImageBuildSettingsBaseReq) ToEntity() *entity.ImageBuildSettings {
	return &entity.ImageBuildSettings{
		Workers:   req.Workers.ToEntity(),
		Resources: req.Resources.ToEntity(),
		Sources:   req.Sources.ToEntity(),
		NoCache:   req.NoCache,
		NoVerbose: req.NoVerbose,
	}
}

func (req *ImageBuildSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.Workers.validate(field+"workers")...)
	res = append(res, req.Resources.validate(field+"resources")...)
	res = append(res, req.Sources.validate(field+"sources")...)
	return res
}

type ImageBuildWorkerSettingsReq struct {
	Nodes          basedto.ObjectIDSliceReq `json:"nodes"`
	NodeLabels     []string                 `json:"nodeLabels"`
	MaxParallelism int                      `json:"maxParallelism"`
}

func (req *ImageBuildWorkerSettingsReq) ToEntity() entity.ImageBuildWorkerSettings {
	if req == nil {
		return entity.ImageBuildWorkerSettings{}
	}
	return entity.ImageBuildWorkerSettings{
		NodeIDs:        req.Nodes.ToIDStringSlice(),
		NodeLabels:     req.NodeLabels,
		MaxParallelism: req.MaxParallelism,
	}
}

func (req *ImageBuildWorkerSettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateObjectIDSliceReq(req.Nodes, true, 0, field+"nodes")...)
	res = append(res, basedto.ValidateSliceEx(req.NodeLabels, true, 0, nodeLabelMaxLen,
		nil, field+"nodeLabels")...)
	return res
}

type ImageBuildResourceSettingsReq struct {
	CPUs    uint          `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap"`
	ShmSize unit.DataSize `json:"shmSize"`
}

func (req *ImageBuildResourceSettingsReq) ToEntity() entity.ImageBuildResourceSettings {
	if req == nil {
		return entity.ImageBuildResourceSettings{}
	}
	return entity.ImageBuildResourceSettings{
		CPUs:    req.CPUs,
		Mem:     req.Mem,
		MemSwap: req.MemSwap,
		ShmSize: req.ShmSize,
	}
}

func (req *ImageBuildResourceSettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.CPUs, false, 0, cpuMax, field+"cpus")...)
	res = append(res, basedto.ValidateNumber(&req.Mem, false, 0, memMax, field+"mem")...)
	res = append(res, basedto.ValidateNumber(&req.MemSwap, false, 0, memMax, field+"memSwap")...)
	res = append(res, basedto.ValidateNumber(&req.ShmSize, false, 0, memMax, field+"shmSize")...)
	return res
}

type ImageBuildSourceSettingsReq struct {
	RepoCache bool `json:"repoCache"`
}

func (req *ImageBuildSourceSettingsReq) ToEntity() entity.ImageBuildSourceSettings {
	if req == nil {
		return entity.ImageBuildSourceSettings{}
	}
	return entity.ImageBuildSourceSettings{
		RepoCache: req.RepoCache,
	}
}

func (req *ImageBuildSourceSettingsReq) validate(_ string) []vld.Validator {
	return nil
}

func NewUpdateImageBuildSettingsReq() *UpdateImageBuildSettingsReq {
	return &UpdateImageBuildSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateImageBuildSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateImageBuildSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
