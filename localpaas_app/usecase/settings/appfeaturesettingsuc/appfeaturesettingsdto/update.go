package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateAppFeatureSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*AppFeatureSettingsBaseReq
}

type AppFeatureSettingsBaseReq struct {
	LoggingSettings  *AppFeatureLoggingSettingsReq  `json:"loggingSettings"`
	SchedJobSettings *AppFeatureSchedJobSettingsReq `json:"schedJobSettings"`
	TerminalSettings *AppFeatureTerminalSettingsReq `json:"terminalSettings"`
}

func (req *AppFeatureSettingsBaseReq) ToEntity() *entity.AppFeatureSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeatureSettings{
		LoggingSettings:  req.LoggingSettings.ToEntity(),
		SchedJobSettings: req.SchedJobSettings.ToEntity(),
		TerminalSettings: req.TerminalSettings.ToEntity(),
	}
}

func (req *AppFeatureSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, req.TerminalSettings.validate(field+"terminalSettings")...)
	res = append(res, req.LoggingSettings.validate(field+"loggingSettings")...)
	res = append(res, req.SchedJobSettings.validate(field+"schedJobSettings")...)
	return res
}

type AppFeatureTerminalSettingsReq struct {
	Enabled bool `json:"enabled"`
}

func (req *AppFeatureTerminalSettingsReq) ToEntity() *entity.AppFeatureTerminalSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeatureTerminalSettings{
		Enabled: req.Enabled,
	}
}

func (req *AppFeatureTerminalSettingsReq) validate(_ string) (res []vld.Validator) {
	return res
}

type AppFeatureLoggingSettingsReq struct {
	Enabled bool `json:"enabled"`
}

func (req *AppFeatureLoggingSettingsReq) ToEntity() *entity.AppFeatureLoggingSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeatureLoggingSettings{
		Enabled: req.Enabled,
	}
}

func (req *AppFeatureLoggingSettingsReq) validate(_ string) (res []vld.Validator) {
	return res
}

type AppFeatureSchedJobSettingsReq struct {
	Enabled bool `json:"enabled"`
}

func (req *AppFeatureSchedJobSettingsReq) ToEntity() *entity.AppFeatureSchedJobSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeatureSchedJobSettings{
		Enabled: req.Enabled,
	}
}

func (req *AppFeatureSchedJobSettingsReq) validate(_ string) (res []vld.Validator) {
	return res
}

func NewUpdateAppFeatureSettingsReq() *UpdateAppFeatureSettingsReq {
	return &UpdateAppFeatureSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppFeatureSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppFeatureSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
