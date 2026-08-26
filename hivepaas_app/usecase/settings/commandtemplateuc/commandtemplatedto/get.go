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
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
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
	Script      string                          `json:"script" copy:"-"` // use manual copy
	WorkingDir  string                          `json:"workingDir,omitempty"`
	EnvVars     []*basedto.EnvVarResp           `json:"envVars,omitempty"`
	ArgGroups   []*CommandTemplateArgGroupResp  `json:"argGroups,omitempty"`
	ConsoleSize *CommandTemplateConsoleSizeResp `json:"consoleSize,omitempty"`
	TTY         bool                            `json:"tty,omitempty"`
	Link        string                          `json:"link,omitempty"`
	Desc        string                          `json:"desc,omitempty"`
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
	refObjects *entity.RefObjects,
	isListAPI bool,
) (resp *CommandTemplateResp, err error) {
	config := setting.MustAsCommandTemplate()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if resp.Command == "" && !isListAPI {
		TransformScript(&config.Script, refObjects, resp)
	}

	return resp, nil
}

func TransformScript(
	script *entity.ObjectValue,
	refObjects *entity.RefObjects,
	resp *CommandTemplateResp,
) {
	if script == nil {
		return
	}
	if script.Value != "" {
		resp.Script = script.Value
	} else if script.ID != "" && refObjects != nil {
		scriptSetting := refObjects.RefSettings[script.ID]
		if scriptSetting != nil {
			resp.Script = scriptSetting.MustAsScript().Data
		}
	}
}
