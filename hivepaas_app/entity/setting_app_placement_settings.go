package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppPlacementSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppPlacementSettings, &appPlacementSettingsParser{})

type appPlacementSettingsParser struct {
}

func (s *appPlacementSettingsParser) New() SettingData {
	return &AppPlacementSettings{}
}

type AppPlacementSettings struct {
	ExcludeManagerNodes bool `json:"excludeManagerNodes,omitempty"`
	ExcludeBuildNodes   bool `json:"excludeBuildNodes,omitempty"`
}

func (s *AppPlacementSettings) GetType() base.SettingType {
	return base.SettingTypeAppPlacementSettings
}

func (s *AppPlacementSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppPlacementSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppPlacementSettings() (*AppPlacementSettings, error) {
	return parseSettingAs[*AppPlacementSettings](s)
}

func (s *Setting) MustAsAppPlacementSettings() *AppPlacementSettings {
	return gofn.Must(s.AsAppPlacementSettings())
}
