package commandtemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateCommandTemplateReq struct {
	settings.UpdateSettingReq
	*CommandTemplateBaseReq
}

func NewUpdateCommandTemplateReq() *UpdateCommandTemplateReq {
	return &UpdateCommandTemplateReq{}
}

func (req *UpdateCommandTemplateReq) ModifyRequest() error {
	return req.CommandTemplateBaseReq.ModifyRequest()
}

func (req *UpdateCommandTemplateReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.CommandTemplateBaseReq.Validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateCommandTemplateResp struct {
	Meta *basedto.Meta `json:"meta"`
}
