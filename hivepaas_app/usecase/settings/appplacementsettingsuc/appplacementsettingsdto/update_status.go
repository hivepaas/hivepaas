package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAppPlacementSettingsStatusReq struct {
	settings.UpdateUniqueSettingStatusReq
}

func NewUpdateAppPlacementSettingsStatusReq() *UpdateAppPlacementSettingsStatusReq {
	return &UpdateAppPlacementSettingsStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppPlacementSettingsStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingStatusReq.Validate()...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppPlacementSettingsStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
