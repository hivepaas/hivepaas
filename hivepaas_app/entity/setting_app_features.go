package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	PreviewSettings  *AppFeaturePreviewSettings  `json:"previewSettings"`
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

type AppFeaturePreviewSettings struct {
	Enabled       bool              `json:"enabled,omitempty"`
	CreationDelay timeutil.Duration `json:"creationDelay,omitempty"`
	AppsToClone   ObjectIDSlice     `json:"appsToClone,omitempty"`
	AutoCloneApps bool              `json:"autoCloneApps,omitempty"`
	Commands      ObjectIDSlice     `json:"commands,omitempty"`
}

func (s *AppFeatureSettings) GetType() base.SettingType {
	return base.SettingTypeAppFeatures
}

func (s *AppFeatureSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.PreviewSettings != nil {
		for _, app := range s.PreviewSettings.AppsToClone {
			if app.ID != "" {
				refIDs.RefAppIDs = append(refIDs.RefAppIDs, app.ID)
			}
		}
		for _, cmd := range s.PreviewSettings.Commands {
			if cmd.ID != "" {
				refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, cmd.ID)
			}
		}
	}
	return refIDs
}

func (s *AppFeatureSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
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
	if settings.PreviewSettings == nil {
		settings.PreviewSettings = &AppFeaturePreviewSettings{Enabled: true}
	}
}
