package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAppPlacementSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*AppPlacementSettingsBaseReq
}

type AppPlacementSettingsBaseReq struct {
	ExcludeManagerNodes bool `json:"excludeManagerNodes"`
	ExcludeBuildNodes   bool `json:"excludeBuildNodes"`
}

func (req *AppPlacementSettingsBaseReq) ToEntity() *entity.AppPlacementSettings {
	if req == nil {
		return nil
	}
	return &entity.AppPlacementSettings{
		ExcludeManagerNodes: req.ExcludeManagerNodes,
		ExcludeBuildNodes:   req.ExcludeBuildNodes,
	}
}

func (req *AppPlacementSettingsBaseReq) validate(_ string) []vld.Validator {
	return nil
}

func NewUpdateAppPlacementSettingsReq() *UpdateAppPlacementSettingsReq {
	return &UpdateAppPlacementSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppPlacementSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppPlacementSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
