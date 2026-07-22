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
	Key       string `json:"k"`
	Value     string `json:"v"`
	IsBuild   bool   `json:"build,omitempty"`
	IsShared  bool   `json:"shared,omitempty"`
	IsLiteral bool   `json:"literal,omitempty"`
	IsSystem  bool   `json:"system,omitempty"`
}

func (s *EnvVars) GetType() base.SettingType {
	return base.SettingTypeEnvVar
}

func (s *EnvVars) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *EnvVars) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *EnvVars) GetEnv(key string) *EnvVar {
	for _, env := range s.Data {
		if env.Key == key {
			return env
		}
	}
	return nil
}

func (s *EnvVars) GetEnvs(kind base.EnvVarKind) []*EnvVar {
	if kind == "" {
		return s.Data
	}
	res := make([]*EnvVar, 0, 10) //nolint:mnd
	for _, env := range s.Data {
		switch kind {
		case base.EnvVarKindRuntime:
			if !env.IsBuild {
				res = append(res, env)
			}
		case base.EnvVarKindShared:
			if env.IsShared {
				res = append(res, env)
			}
		case base.EnvVarKindBuild:
			if env.IsBuild {
				res = append(res, env)
			}
		}
	}
	return res
}

func (s *EnvVars) GetSystemEnvs(kind base.EnvVarKind) []*EnvVar {
	res := make([]*EnvVar, 0, 10) //nolint:mnd
	for _, env := range s.Data {
		if !env.IsSystem {
			continue
		}
		switch kind {
		case "":
			res = append(res, env)
		case base.EnvVarKindRuntime:
			if !env.IsBuild {
				res = append(res, env)
			}
		case base.EnvVarKindShared:
			if env.IsShared {
				res = append(res, env)
			}
		case base.EnvVarKindBuild:
			if env.IsBuild {
				res = append(res, env)
			}
		}
	}
	return res
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
