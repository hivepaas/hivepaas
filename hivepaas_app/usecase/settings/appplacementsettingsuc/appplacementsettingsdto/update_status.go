package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAppPlacementSettingsStatusReq struct {
	settings.UpdateUniqueSettingStatusReq
}

func NewUpdateAppPlacementSettingsStatusReq() *UpdateAppPlacementSettingsStatusReq {
	return &UpdateAppPlacementSettingsStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppPlacementSettingsStatusReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppPlacementSettingsStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
