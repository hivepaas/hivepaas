package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateImageBuildSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*ImageBuildSettingsBaseReq
}

type ImageBuildSettingsBaseReq struct {
	Resources *ImageBuildSettingResourcesReq `json:"resources"`
	NoCache   bool                           `json:"noCache"`
	NoVerbose bool                           `json:"noVerbose"`
}

func (req *ImageBuildSettingsBaseReq) ToEntity() *entity.ImageBuildSettings {
	return &entity.ImageBuildSettings{
		Resources: req.Resources.ToEntity(),
		NoCache:   req.NoCache,
		NoVerbose: req.NoVerbose,
	}
}

// nolint
func (req *ImageBuildSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

type ImageBuildSettingResourcesReq struct {
	CPUs    int32         `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap"`
	ShmSize unit.DataSize `json:"shmSize"`
}

func (req *ImageBuildSettingResourcesReq) ToEntity() *entity.ImageBuildSettingResources {
	return &entity.ImageBuildSettingResources{
		CPUs:    req.CPUs,
		Mem:     req.Mem,
		MemSwap: req.MemSwap,
		ShmSize: req.ShmSize,
	}
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
