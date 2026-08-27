package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAppFeatureSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteAppFeatureSettingsReq() *DeleteAppFeatureSettingsReq {
	return &DeleteAppFeatureSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppFeatureSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppFeatureSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
