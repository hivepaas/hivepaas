package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSecretReq struct {
	settings.UpdateSettingReq
	*SecretBaseReq
}

func NewUpdateSecretReq() *UpdateSecretReq {
	return &UpdateSecretReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSecretReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate(false, "")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
}
