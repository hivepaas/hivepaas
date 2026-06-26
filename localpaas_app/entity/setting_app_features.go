package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentAppFeatureSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppFeatures, &appFeatureSettingsParser{})

type appFeatureSettingsParser struct {
}

func (s *appFeatureSettingsParser) New() SettingData {
	return &AppFeatureSettings{}
}

type AppFeatureSettings struct {
	TerminalSettings *AppFeatureTerminalSettings `json:"terminalSettings"`
	LoggingSettings  *AppFeatureLoggingSettings  `json:"loggingSettings"`
	SchedJobSettings *AppFeatureSchedJobSettings `json:"schedJobSettings"`
}

type AppFeatureTerminalSettings struct {
	Enabled bool `json:"enabled,omitempty"`
}

type AppFeatureLoggingSettings struct {
	Enabled bool `json:"enabled,omitempty"`
}

type AppFeatureSchedJobSettings struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *AppFeatureSettings) GetType() base.SettingType {
	return base.SettingTypeAppFeatures
}

func (s *AppFeatureSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *AppFeatureSettings) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AppFeatureSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppFeatureSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppFeatureSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppFeatureSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAppFeatureSettings() (*AppFeatureSettings, error) {
	return parseSettingAs[*AppFeatureSettings](s)
}

func (s *Setting) MustAsAppFeatureSettings() *AppFeatureSettings {
	return gofn.Must(s.AsAppFeatureSettings())
}

func InitAppFeatureSettingsDefault(settings *AppFeatureSettings) {
	if settings == nil {
		return
	}
	if settings.LoggingSettings == nil {
		settings.LoggingSettings = &AppFeatureLoggingSettings{Enabled: true}
	}
	if settings.SchedJobSettings == nil {
		settings.SchedJobSettings = &AppFeatureSchedJobSettings{Enabled: true}
	}
	if settings.TerminalSettings == nil {
		settings.TerminalSettings = &AppFeatureTerminalSettings{Enabled: true}
	}
}
