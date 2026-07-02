package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteImageBuildSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteImageBuildSettingsReq() *DeleteImageBuildSettingsReq {
	return &DeleteImageBuildSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteImageBuildSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteImageBuildSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
