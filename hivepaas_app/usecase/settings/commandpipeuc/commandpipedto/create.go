package commandpipedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CreateCommandPipeReq struct {
	settings.CreateSettingReq
	*CommandPipeBaseReq
}

type CommandPipeBaseReq struct {
	Name          string              `json:"name"`
	SourceCommand basedto.ObjectIDReq `json:"sourceCommand"`
	TargetCommand basedto.ObjectIDReq `json:"targetCommand"`
}

func (req *CommandPipeBaseReq) ToEntity() *entity.CommandPipe {
	return &entity.CommandPipe{
		SourceCommand: *req.SourceCommand.ToEntity(),
		TargetCommand: *req.TargetCommand.ToEntity(),
	}
}

func (req *CommandPipeBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	return nil
}

func (req *CommandPipeBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.SourceCommand, true, field+"sourceCommand")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.TargetCommand, true, field+"targetCommand")...)
	return res
}

func NewCreateCommandPipeReq() *CreateCommandPipeReq {
	return &CreateCommandPipeReq{}
}

func (req *CreateCommandPipeReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCommandPipeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCommandPipeResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
