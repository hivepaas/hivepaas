package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetImageBuildSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetImageBuildSettingsReq() *GetImageBuildSettingsReq {
	return &GetImageBuildSettingsReq{}
}

func (req *GetImageBuildSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetImageBuildSettingsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *ImageBuildSettingsResp `json:"data"`
}

type ImageBuildSettingsResp struct {
	*settings.BaseSettingResp
	Resources *ImageBuildResourceSettingsResp `json:"resources"`
	Sources   *ImageBuildSourceSettingsResp   `json:"sources"`
	NoCache   bool                            `json:"noCache"`
	NoVerbose bool                            `json:"noVerbose"`
}

type ImageBuildResourceSettingsResp struct {
	CPUs    uint          `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap"`
	ShmSize unit.DataSize `json:"shmSize"`
}

type ImageBuildSourceSettingsResp struct {
	RepoCache bool `json:"repoCache"`
}

func TransformImageBuild(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *ImageBuildSettingsResp, err error) {
	config := setting.MustAsImageBuildSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
