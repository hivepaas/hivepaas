package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type DeleteUniqueStorageSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteUniqueStorageSettingsReq() *DeleteUniqueStorageSettingsReq {
	return &DeleteUniqueStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteUniqueStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteUniqueStorageSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
