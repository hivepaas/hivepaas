package commandpipedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (req *CreateCommandPipeFromTemplateReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.CommandType = strings.ToLower(strings.TrimSpace(req.CommandType))
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCommandPipeFromTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 1,
		base.SettingNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStr(&req.CommandType, true, 1,
		base.SettingNameMaxLen, "commandType")...)
	validators = append(validators, basedto.ValidateStr(&req.CommandKind, true, 1,
		base.SettingNameMaxLen, "commandKind")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCommandPipeFromTemplateResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
