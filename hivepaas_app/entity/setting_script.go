package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentScriptVersion = 1
)

var _ = registerSettingParser(base.SettingTypeScript, &scriptParser{})

type scriptParser struct {
}

func (s *scriptParser) New() SettingData {
	return &Script{}
}

type Script struct {
	Data string `json:"data"`
}

func (s *Script) GetType() base.SettingType {
	return base.SettingTypeScript
}

func (s *Script) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *Script) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsScript() (*Script, error) {
	return parseSettingAs[*Script](s)
}

func (s *Setting) MustAsScript() *Script {
	return gofn.Must(s.AsScript())
}
