package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentCommandTplVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCommandTpl, &commandTplParser{})

type commandTplParser struct {
}

func (s *commandTplParser) New() SettingData {
	return &CommandTpl{}
}

type CommandTpl struct {
	Command     string                `json:"command"`
	Script      string                `json:"script,omitempty"`
	WorkingDir  string                `json:"workingDir,omitempty"`
	EnvVars     []*EnvVar             `json:"envVars,omitempty"`
	ArgGroups   []*CommandTplArgGroup `json:"argGroups,omitempty"`
	ConsoleSize CommandTplConsoleSize `json:"consoleSize"`
	TTY         bool                  `json:"tty,omitempty"`
}

type CommandTplArgGroup struct {
	Enabled   bool             `json:"enabled"`
	ExportEnv string           `json:"exportEnv"`
	Separator string           `json:"separator"`
	Args      []*CommandTplArg `json:"args,omitempty"`
}

type CommandTplArg struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CommandTplConsoleSize struct {
	Width  uint `json:"w"`
	Height uint `json:"h"`
}

func (s *CommandTpl) GetType() base.SettingType {
	return base.SettingTypeCommandTpl
}

func (s *CommandTpl) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *CommandTpl) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *CommandTpl) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCommandTplVersion {
		return false, nil
	}
	if setting.Version > CurrentCommandTplVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCommandTplVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsCommandTpl() (*CommandTpl, error) {
	return parseSettingAs[*CommandTpl](s)
}

func (s *Setting) MustAsCommandTpl() *CommandTpl {
	return gofn.Must(s.AsCommandTpl())
}
