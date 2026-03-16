package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
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

func (s *ProjectSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentProjectSettingsVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentProjectSettingsVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsProjectSettings() (*ProjectSettings, error) {
	return parseSettingAs[*ProjectSettings](s)
}

func (s *Setting) MustAsProjectSettings() *ProjectSettings {
	return gofn.Must(s.AsProjectSettings())
}
