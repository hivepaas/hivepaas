package basicauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateBasicAuthReq struct {
	settings.UpdateSettingReq
	*BasicAuthBaseReq
}

func NewUpdateBasicAuthReq() *UpdateBasicAuthReq {
	return &UpdateBasicAuthReq{}
}

func (req *UpdateBasicAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateBasicAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateBasicAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
