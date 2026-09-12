package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppPlacementVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppPlacement, &appPlacementSettingsParser{})

type appPlacementSettingsParser struct {
}

func (s *appPlacementSettingsParser) New() SettingData {
	return &AppPlacementSettings{}
}

type AppPlacementSettings struct {
	ExcludeManagerNodes bool `json:"excludeManagerNodes,omitempty"`
	ExcludeBuildNodes   bool `json:"excludeBuildNodes,omitempty"`

	// RequireNodeLabels keeps apps on nodes carrying these labels, as
	// `key=value` (a bare key means `key=true`).
	//
	// Swarm ANDs placement constraints, so several entries mean a node must
	// carry all of them - there is no "any of these".
	RequireNodeLabels []string `json:"requireNodeLabels,omitempty"`

	// ExcludeNodeLabels keeps apps off nodes carrying these labels, in the same
	// form. A node is excluded when it matches any entry.
	ExcludeNodeLabels []string `json:"excludeNodeLabels,omitempty"`
}

func (s *AppPlacementSettings) GetType() base.SettingType {
	return base.SettingTypeAppPlacement
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
