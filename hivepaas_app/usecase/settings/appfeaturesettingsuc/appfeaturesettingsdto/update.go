package appfeaturesettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateAppFeatureSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*AppFeatureSettingsBaseReq
}

type AppFeatureSettingsBaseReq struct {
	LoggingSettings  *AppFeatureLoggingSettingsReq  `json:"loggingSettings"`
	SchedJobSettings *AppFeatureSchedJobSettingsReq `json:"schedJobSettings"`
	TerminalSettings *AppFeatureTerminalSettingsReq `json:"terminalSettings"`
	PreviewSettings  *AppFeaturePreviewSettingsReq  `json:"previewSettings"`
}

func (req *AppFeatureSettingsBaseReq) ToEntity() *entity.AppFeatureSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeatureSettings{
		LoggingSettings:  req.LoggingSettings.ToEntity(),
		SchedJobSettings: req.SchedJobSettings.ToEntity(),
		TerminalSettings: req.TerminalSettings.ToEntity(),
		PreviewSettings:  req.PreviewSettings.ToEntity(),
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
	res = append(res, req.PreviewSettings.validate(field+"previewSettings")...)
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

type AppFeaturePreviewSettingsReq struct {
	Enabled       bool                     `json:"enabled"`
	CreationDelay timeutil.Duration        `json:"creationDelay"`
	AppsToClone   basedto.ObjectIDSliceReq `json:"appsToClone"`
	AutoCloneApps bool                     `json:"autoCloneApps"`
	Commands      basedto.ObjectIDSliceReq `json:"commands"`
}

func (req *AppFeaturePreviewSettingsReq) ToEntity() *entity.AppFeaturePreviewSettings {
	if req == nil {
		return nil
	}
	return &entity.AppFeaturePreviewSettings{
		Enabled:       req.Enabled,
		CreationDelay: req.CreationDelay,
		AppsToClone:   req.AppsToClone.ToEntity(),
		AutoCloneApps: req.AutoCloneApps,
		Commands:      req.Commands.ToEntity(),
	}
}

func (req *AppFeaturePreviewSettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateObjectIDSliceReq(req.AppsToClone, true, 0, field+"appsToClone")...)
	res = append(res, basedto.ValidateObjectIDSliceReq(req.Commands, true, 0, field+"commands")...)
	return res
}

func NewUpdateAppFeatureSettingsReq() *UpdateAppFeatureSettingsReq {
	return &UpdateAppFeatureSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppFeatureSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppFeatureSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
