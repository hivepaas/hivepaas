package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateIMServiceReq struct {
	settings.UpdateSettingReq
	*IMServiceBaseReq
}

func NewUpdateIMServiceReq() *UpdateIMServiceReq {
	return &UpdateIMServiceReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateIMServiceReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateIMServiceResp struct {
	Meta *basedto.Meta `json:"meta"`
}
