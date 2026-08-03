package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
	var validators []vld.Validator
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

	if config.SourceCommand.ID != "" && refObjects != nil {
		sourceSetting := refObjects.RefSettings[config.SourceCommand.ID]
		resp.SourceCommand, err = settings.TransformSettingBase(sourceSetting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	if config.TargetCommand.ID != "" && refObjects != nil {
		targetSetting := refObjects.RefSettings[config.TargetCommand.ID]
		resp.TargetCommand, err = settings.TransformSettingBase(targetSetting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return resp, nil
}
