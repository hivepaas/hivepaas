package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentProjectSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeProject, &projectSettingsParser{})

type projectSettingsParser struct {
}

func (s *projectSettingsParser) New() SettingData {
	return &ProjectSettings{}
}

type ProjectSettings struct {
	Test string `json:"test"`
}

func (s *ProjectSettings) GetType() base.SettingType {
	return base.SettingTypeProject
}

func (s *ProjectSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ProjectSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsProjectSettings() (*ProjectSettings, error) {
	return parseSettingAs[*ProjectSettings](s)
}

func (s *Setting) MustAsProjectSettings() *ProjectSettings {
	return gofn.Must(s.AsProjectSettings())
}
