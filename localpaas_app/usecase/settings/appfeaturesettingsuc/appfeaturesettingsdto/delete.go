package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
