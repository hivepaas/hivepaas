package commandtemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetCommandTemplateReq struct {
	settings.GetSettingReq
}

func NewGetCommandTemplateReq() *GetCommandTemplateReq {
	return &GetCommandTemplateReq{}
}

func (req *GetCommandTemplateReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCommandTemplateResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *CommandTemplateResp `json:"data"`
}

type CommandTemplateResp struct {
	*settings.BaseSettingResp
	Command     string                          `json:"command"`
	Script      string                          `json:"script"`
	WorkingDir  string                          `json:"workingDir,omitempty"`
	EnvVars     []*basedto.EnvVarResp           `json:"envVars,omitempty"`
	ArgGroups   []*CommandTemplateArgGroupResp  `json:"argGroups,omitempty"`
	ConsoleSize *CommandTemplateConsoleSizeResp `json:"consoleSize,omitempty"`
	TTY         bool                            `json:"tty,omitempty"`
}

type CommandTemplateArgGroupResp struct {
	Enabled   bool                      `json:"enabled"`
	ExportEnv string                    `json:"exportEnv"`
	Separator string                    `json:"separator"`
	Args      []*CommandTemplateArgResp `json:"args"`
}

type CommandTemplateArgResp struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CommandTemplateConsoleSizeResp struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

func TransformCommandTemplate(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *CommandTemplateResp, err error) {
	config := setting.MustAsCommandTemplate()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resp, nil
}
