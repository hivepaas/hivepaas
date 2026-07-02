package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateImageBuildSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*ImageBuildSettingsBaseReq
}

type ImageBuildSettingsBaseReq struct {
	Resources *ImageBuildResourceSettingsReq `json:"resources"`
	Sources   *ImageBuildSourceSettingsReq   `json:"sources"`
	NoCache   bool                           `json:"noCache"`
	NoVerbose bool                           `json:"noVerbose"`
}

func (req *ImageBuildSettingsBaseReq) ToEntity() *entity.ImageBuildSettings {
	return &entity.ImageBuildSettings{
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
	res = append(res, req.Resources.validate(field+"resources")...)
	res = append(res, req.Sources.validate(field+"sources")...)
	return res
}

type ImageBuildResourceSettingsReq struct {
	CPUs    uint          `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap"`
	ShmSize unit.DataSize `json:"shmSize"`
}

func (req *ImageBuildResourceSettingsReq) ToEntity() entity.ImageBuildResourceSettings {
	return entity.ImageBuildResourceSettings{
		CPUs:    req.CPUs,
		Mem:     req.Mem,
		MemSwap: req.MemSwap,
		ShmSize: req.ShmSize,
	}
}

// nolint
func (req *ImageBuildResourceSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type ImageBuildSourceSettingsReq struct {
	RepoCache bool `json:"repoCache"`
}

func (req *ImageBuildSourceSettingsReq) ToEntity() entity.ImageBuildSourceSettings {
	return entity.ImageBuildSourceSettings{
		RepoCache: req.RepoCache,
	}
}

// nolint
func (req *ImageBuildSourceSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

func NewUpdateImageBuildSettingsReq() *UpdateImageBuildSettingsReq {
	return &UpdateImageBuildSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateImageBuildSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateImageBuildSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
