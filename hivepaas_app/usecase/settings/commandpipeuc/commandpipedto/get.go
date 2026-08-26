package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetCommandPipeReq struct {
	settings.GetSettingReq
}

func NewGetCommandPipeReq() *GetCommandPipeReq {
	return &GetCommandPipeReq{}
}

func (req *GetCommandPipeReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetCommandPipeResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *CommandPipeResp `json:"data"`
}

type CommandPipeResp struct {
	*settings.BaseSettingResp
	SourceCommand *settings.BaseSettingResp `json:"sourceCommand,omitempty"`
	TargetCommand *settings.BaseSettingResp `json:"targetCommand,omitempty"`
}

func TransformCommandPipe(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *CommandPipeResp, err error) {
	config := setting.MustAsCommandPipe()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if refObjects == nil {
		refObjects = &entity.RefObjects{}
	}

	if config.SourceCommand.ID != "" {
		sourceSetting := refObjects.RefSettings[config.SourceCommand.ID]
		cmdResp, _ := settings.TransformSettingBase(sourceSetting)
		if cmdResp == nil {
			cmdResp = settings.NewMissingSetting(config.SourceCommand.ID, base.SettingTypeCommandTemplate)
		}
		resp.SourceCommand = cmdResp
	}

	if config.TargetCommand.ID != "" {
		targetSetting := refObjects.RefSettings[config.TargetCommand.ID]
		cmdResp, _ := settings.TransformSettingBase(targetSetting)
		if cmdResp == nil {
			cmdResp = settings.NewMissingSetting(config.TargetCommand.ID, base.SettingTypeCommandTemplate)
		}
		resp.TargetCommand = cmdResp
	}

	return resp, nil
}
