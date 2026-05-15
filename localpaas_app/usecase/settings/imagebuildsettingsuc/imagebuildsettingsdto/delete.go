package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
