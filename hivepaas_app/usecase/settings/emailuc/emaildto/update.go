package emaildto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateEmailReq struct {
	settings.UpdateSettingReq
	*EmailBaseReq
}

func NewUpdateEmailReq() *UpdateEmailReq {
	return &UpdateEmailReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateEmailReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateEmailResp struct {
	Meta *basedto.Meta `json:"meta"`
}
