package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteStorageSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteStorageSettingsReq() *DeleteStorageSettingsReq {
	return &DeleteStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteStorageSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
