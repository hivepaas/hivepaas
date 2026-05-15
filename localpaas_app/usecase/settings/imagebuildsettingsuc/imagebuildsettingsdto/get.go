package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	Resources *ImageBuildSettingResourcesResp `json:"resources"`
	NoCache   bool                            `json:"noCache"`
	NoVerbose bool                            `json:"noVerbose"`
}

type ImageBuildSettingResourcesResp struct {
	CPUs      int32 `json:"cpus"`
	MemMB     int64 `json:"memMB"`
	MemSwapMB int64 `json:"memSwapMB"`
	ShmSizeMB int64 `json:"shmSizeMB"`
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
