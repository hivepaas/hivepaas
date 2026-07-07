package commandtemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteCommandTemplateReq struct {
	settings.DeleteSettingReq
}

func NewDeleteCommandTemplateReq() *DeleteCommandTemplateReq {
	return &DeleteCommandTemplateReq{}
}

func (req *DeleteCommandTemplateReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteCommandTemplateResp struct {
	Meta *basedto.Meta `json:"meta"`
}
