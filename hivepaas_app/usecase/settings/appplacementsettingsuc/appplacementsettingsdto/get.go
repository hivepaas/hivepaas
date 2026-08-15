package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetAppPlacementSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetAppPlacementSettingsReq() *GetAppPlacementSettingsReq {
	return &GetAppPlacementSettingsReq{}
}

func (req *GetAppPlacementSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppPlacementSettingsResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *AppPlacementSettingsResp `json:"data"`
}

type AppPlacementSettingsResp struct {
	*settings.BaseSettingResp
	ExcludeManagerNodes bool `json:"excludeManagerNodes,omitempty"`
	ExcludeBuildNodes   bool `json:"excludeBuildNodes,omitempty"`
}

func TransformImageBuild(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *AppPlacementSettingsResp, err error) {
	config := setting.MustAsAppPlacementSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
