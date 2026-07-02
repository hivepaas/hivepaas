package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
func (req *UpdateSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate(false, "")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
}
