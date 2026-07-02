package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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

func (s *ProjectSettings) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ProjectSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentProjectSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentProjectSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentProjectSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsProjectSettings() (*ProjectSettings, error) {
	return parseSettingAs[*ProjectSettings](s)
}

func (s *Setting) MustAsProjectSettings() *ProjectSettings {
	return gofn.Must(s.AsProjectSettings())
}
