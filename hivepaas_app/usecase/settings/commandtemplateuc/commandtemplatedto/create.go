package commandtemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CreateCommandTemplateReq struct {
	settings.CreateSettingReq
	*CommandTemplateBaseReq
}

type CommandTemplateBaseReq struct {
	Name        string                         `json:"name"`
	Kind        base.CommandTemplateKind       `json:"kind"`
	Command     string                         `json:"command"`
	Script      string                         `json:"script"`
	WorkingDir  string                         `json:"workingDir"`
	EnvVars     []*basedto.EnvVarReq           `json:"envVars"`
	ArgGroups   []*CommandTemplateArgGroupReq  `json:"argGroups"`
	ConsoleSize *CommandTemplateConsoleSizeReq `json:"consoleSize"`
	TTY         bool                           `json:"tty"`
}

func (req *CommandTemplateBaseReq) ToEntity() *entity.CommandTemplate {
	if req == nil {
		return nil
	}
	return &entity.CommandTemplate{
		Command:    req.Command,
		Script:     req.Script,
		WorkingDir: req.WorkingDir,
		EnvVars: gofn.MapSlice(req.EnvVars, func(item *basedto.EnvVarReq) *entity.EnvVar {
			return item.ToEntity(base.EnvVarKindRuntime)
		}),
		ArgGroups: gofn.MapSlice(req.ArgGroups, func(item *CommandTemplateArgGroupReq) *entity.CommandTemplateArgGroup {
			return item.ToEntity()
		}),
		ConsoleSize: req.ConsoleSize.ToEntity(),
		TTY:         req.TTY,
	}
}

func (req *CommandTemplateBaseReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.WorkingDir = strings.TrimSpace(req.WorkingDir)
	req.Script = strings.ReplaceAll(req.Script, "\r\n", "\n")
	return nil
}

func (req *CommandTemplateBaseReq) Validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	cmdValid := (req.Command != "" && req.Script == "") || (req.Command == "" && req.Script != "")
	res = append(res, basedto.ValidateCond(cmdValid, field+"command|script")...)
	res = append(res, basedto.ValidateStr(&req.Command, false, 1, int(base.ExecCommandMaxSize), field+"command")...)
	res = append(res, basedto.ValidateStr(&req.Script, false, 1, int(base.ExecCommandMaxSize), field+"script")...)
	return res
}

type CommandTemplateArgGroupReq struct {
	Enabled   bool                     `json:"enabled"`
	ExportEnv string                   `json:"exportEnv"`
	Separator string                   `json:"separator"`
	Args      []*CommandTemplateArgReq `json:"args"`
}

func (req *CommandTemplateArgGroupReq) ToEntity() *entity.CommandTemplateArgGroup {
	if req == nil {
		return nil
	}
	return &entity.CommandTemplateArgGroup{
		Enabled:   req.Enabled,
		ExportEnv: req.ExportEnv,
		Separator: req.Separator,
		Args: gofn.MapSlice(req.Args, func(item *CommandTemplateArgReq) *entity.CommandTemplateArg {
			return item.ToEntity()
		}),
	}
}

type CommandTemplateArgReq struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (req *CommandTemplateArgReq) ToEntity() *entity.CommandTemplateArg {
	if req == nil {
		return nil
	}
	return &entity.CommandTemplateArg{
		Use:   req.Use,
		Name:  req.Name,
		Value: req.Value,
	}
}

type CommandTemplateConsoleSizeReq struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

func (req *CommandTemplateConsoleSizeReq) ToEntity() entity.CommandTemplateConsoleSize {
	if req == nil {
		return entity.CommandTemplateConsoleSize{}
	}
	return entity.CommandTemplateConsoleSize{
		Width:  req.Width,
		Height: req.Height,
	}
}

func NewCreateCommandTemplateReq() *CreateCommandTemplateReq {
	return &CreateCommandTemplateReq{}
}

func (req *CreateCommandTemplateReq) ModifyRequest() error {
	return req.CommandTemplateBaseReq.ModifyRequest()
}

func (req *CreateCommandTemplateReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.CommandTemplateBaseReq.Validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCommandTemplateResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
