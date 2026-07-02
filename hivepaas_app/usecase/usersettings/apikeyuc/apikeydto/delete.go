package apikeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAPIKeyReq struct {
	settings.DeleteSettingReq
}

func NewDeleteAPIKeyReq() *DeleteAPIKeyReq {
	return &DeleteAPIKeyReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAPIKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAPIKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
}
