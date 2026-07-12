package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentProjectEnvsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeProjectEnvs, &projectEnvsParser{})

type projectEnvsParser struct {
}

func (s *projectEnvsParser) New() SettingData {
	return &ProjectEnvs{}
}

type ProjectEnvs struct {
	Envs []*Env `json:"envs"`
}

type Env struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (s *ProjectEnvs) GetType() base.SettingType {
	return base.SettingTypeProjectEnvs
}

func (s *ProjectEnvs) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ProjectEnvs) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ProjectEnvs) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentProjectEnvsVersion {
		return false, nil
	}
	if setting.Version > CurrentProjectEnvsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentProjectEnvsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsProjectEnvs() (*ProjectEnvs, error) {
	return parseSettingAs[*ProjectEnvs](s)
}

func (s *Setting) MustAsProjectEnvs() *ProjectEnvs {
	return gofn.Must(s.AsProjectEnvs())
}
