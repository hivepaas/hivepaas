package entity

import (
	"bytes"
	"encoding/json"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

// UnmarshalJSON reads a value that was not written as text.
//
// What reaches a container is text either way: the environment has no other
// type. But a value is not always typed by the person who wrote it - a template
// puts one of its own parameters here, and a parameter that asks a yes-or-no
// question arrives as true, not "true" - and somebody writing the same thing by
// hand has no reason to think about the quotes. So a scalar is taken as what it
// would print as, and only what could not become a value at all is refused.
func (e *EnvVar) UnmarshalJSON(data []byte) error {
	type envVar EnvVar // Without this type's methods, so this one does not call itself.
	aux := &struct {
		Value json.RawMessage `json:"v"`
		*envVar
	}{envVar: (*envVar)(e)}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(aux); err != nil {
		return hperrors.Wrap(err)
	}
	value, err := envValueAsString(aux.Value)
	if err != nil {
		return err
	}
	e.Value = value
	return nil
}

func envValueAsString(raw json.RawMessage) (string, error) {
	text := string(bytes.TrimSpace(raw))
	switch {
	case text == "" || text == "null":
		return "", nil
	case text[0] == '"':
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", hperrors.Wrap(err)
		}
		return value, nil
	case text[0] == '{' || text[0] == '[':
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithExtraDetail(
			"an environment variable is text, a number or true/false, not %s", text)
	}
	return text, nil
}

func (e *EnvVar) Equal(e2 *EnvVar) bool {
	if e == nil || e2 == nil {
		return e == nil && e2 == nil
	}
	// NOTE: skip IsSystem comparing
	return e.Key == e2.Key && e.Value == e2.Value && e.IsBuild == e2.IsBuild &&
		e.IsShared == e2.IsShared && e.IsLiteral == e2.IsLiteral
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

func (s *Setting) AsEnvVars() (*EnvVars, error) {
	return parseSettingAs[*EnvVars](s)
}

func (s *Setting) MustAsEnvVars() *EnvVars {
	return gofn.Must(s.AsEnvVars())
}
