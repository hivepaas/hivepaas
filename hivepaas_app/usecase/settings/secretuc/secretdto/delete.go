package secretdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSecretReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSecretReq() *DeleteSecretReq {
	return &DeleteSecretReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSecretReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSecretResp struct {
	Meta *basedto.Meta `json:"meta"`
}
