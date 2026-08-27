package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateOAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateOAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
