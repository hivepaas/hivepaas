package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAppFeatureSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteAppFeatureSettingsReq() *DeleteAppFeatureSettingsReq {
	return &DeleteAppFeatureSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppFeatureSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppFeatureSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
