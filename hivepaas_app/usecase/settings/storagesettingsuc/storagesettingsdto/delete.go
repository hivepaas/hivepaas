package storagesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteStorageSettingsReq struct {
	settings.DeleteUniqueSettingReq
}

func NewDeleteStorageSettingsReq() *DeleteStorageSettingsReq {
	return &DeleteStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteStorageSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteStorageSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
