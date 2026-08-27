package commandtemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *UpdateCommandTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CommandTemplateBaseReq.Validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateCommandTemplateResp struct {
	Meta *basedto.Meta `json:"meta"`
}
