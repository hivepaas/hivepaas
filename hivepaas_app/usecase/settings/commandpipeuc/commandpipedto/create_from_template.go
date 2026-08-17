package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CreateCommandPipeFromTemplateReq struct {
	settings.CreateSettingReq
	Name        string `json:"name"`
	CommandType string `json:"commandType"`
	CommandKind string `json:"commandKind"`
}

func NewCreateCommandPipeFromTemplateReq() *CreateCommandPipeFromTemplateReq {
	return &CreateCommandPipeFromTemplateReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCommandPipeFromTemplateReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1,
		base.SettingNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStr(&req.CommandType, true, 1,
		base.SettingNameMaxLen, "commandType")...)
	validators = append(validators, basedto.ValidateStr(&req.CommandKind, true, 1,
		base.SettingNameMaxLen, "commandKind")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCommandPipeFromTemplateResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
