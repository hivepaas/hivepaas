package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetAppFeatureSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetAppFeatureSettingsReq() *GetAppFeatureSettingsReq {
	return &GetAppFeatureSettingsReq{}
}

func (req *GetAppFeatureSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppFeatureSettingsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppFeatureSettingsResp `json:"data"`
}

type AppFeatureSettingsResp struct {
	*settings.BaseSettingResp
	TerminalSettings *AppFeatureTerminalSettingsResp `json:"terminalSettings"`
	LoggingSettings  *AppFeatureLoggingSettingsResp  `json:"loggingSettings"`
	SchedJobSettings *AppFeatureSchedJobSettingsResp `json:"schedJobSettings"`
}

type AppFeatureTerminalSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeatureLoggingSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeatureSchedJobSettingsResp struct {
	Enabled bool `json:"enabled"`
}

type AppFeatureSettingsTransformInput struct {
	Setting *entity.Setting
}

func TransformAppFeatureSettings(
	input *AppFeatureSettingsTransformInput,
) (resp *AppFeatureSettingsResp, err error) {
	config := input.Setting.MustAsAppFeatureSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(input.Setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resp, nil
}
