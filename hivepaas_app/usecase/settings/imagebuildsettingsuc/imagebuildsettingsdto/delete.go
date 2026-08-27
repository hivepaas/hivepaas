package imagebuildsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteImageBuildSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteImageBuildSettingsReq() *DeleteImageBuildSettingsReq {
	return &DeleteImageBuildSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteImageBuildSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteImageBuildSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
