package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAppPlacementSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteAppPlacementSettingsReq() *DeleteAppPlacementSettingsReq {
	return &DeleteAppPlacementSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppPlacementSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppPlacementSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
