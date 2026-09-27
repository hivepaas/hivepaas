package keyauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateKeyAuthReq struct {
	settings.UpdateSettingReq
	*KeyAuthBaseReq
}

func NewUpdateKeyAuthReq() *UpdateKeyAuthReq {
	return &UpdateKeyAuthReq{}
}

func (req *UpdateKeyAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateKeyAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateKeyAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
