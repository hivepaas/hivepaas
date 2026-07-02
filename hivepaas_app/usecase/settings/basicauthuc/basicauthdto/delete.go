package basicauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteBasicAuthReq struct {
	settings.DeleteSettingReq
}

func NewDeleteBasicAuthReq() *DeleteBasicAuthReq {
	return &DeleteBasicAuthReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteBasicAuthReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteBasicAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
