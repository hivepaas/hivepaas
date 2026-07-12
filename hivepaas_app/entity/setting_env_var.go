package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentEnvVarsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeEnvVar, &envVarsParser{})

type envVarsParser struct {
}

func (s *envVarsParser) New() SettingData {
	return &EnvVars{}
}

type EnvVars struct {
	Data []*EnvVar `json:"data"`
}

type EnvVar struct {
	Key        string `json:"k"`
	Value      string `json:"v"`
	IsBuildEnv bool   `json:"isBuildEnv,omitempty"`
	IsLiteral  bool   `json:"isLiteral,omitempty"`
}

func (s *EnvVars) GetType() base.SettingType {
	return base.SettingTypeEnvVar
}

func (s *EnvVars) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *EnvVars) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *EnvVars) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentEnvVarsVersion {
		return false, nil
	}
	if setting.Version > CurrentEnvVarsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentEnvVarsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsEnvVars() (*EnvVars, error) {
	return parseSettingAs[*EnvVars](s)
}

func (s *Setting) MustAsEnvVars() *EnvVars {
	return gofn.Must(s.AsEnvVars())
}
