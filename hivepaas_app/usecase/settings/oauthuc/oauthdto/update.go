package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateOAuthReq struct {
	settings.UpdateSettingReq
	*OAuthBaseReq
}

func NewUpdateOAuthReq() *UpdateOAuthReq {
	return &UpdateOAuthReq{}
}

func (req *UpdateOAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateOAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateOAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
