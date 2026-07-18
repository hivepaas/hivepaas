package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentCommandTemplateVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCommandTemplate, &commandTemplateParser{})

type commandTemplateParser struct {
}

func (s *commandTemplateParser) New() SettingData {
	return &CommandTemplate{}
}

type CommandTemplate struct {
	Command     string                     `json:"command"`
	Script      string                     `json:"script,omitempty"`
	WorkingDir  string                     `json:"workingDir,omitempty"`
	EnvVars     []*EnvVar                  `json:"envVars,omitempty"`
	ArgGroups   []*CommandTemplateArgGroup `json:"argGroups,omitempty"`
	ConsoleSize CommandTemplateConsoleSize `json:"consoleSize,omitzero"`
	TTY         bool                       `json:"tty,omitempty"`
}

type CommandTemplateArgGroup struct {
	Enabled   bool                  `json:"enabled"`
	ExportEnv string                `json:"exportEnv"`
	Separator string                `json:"separator"`
	Args      []*CommandTemplateArg `json:"args,omitempty"`
}

type CommandTemplateArg struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CommandTemplateConsoleSize struct {
	Width  uint `json:"w"`
	Height uint `json:"h"`
}

func (s *CommandTemplate) GetType() base.SettingType {
	return base.SettingTypeCommandTemplate
}

func (s *CommandTemplate) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *CommandTemplate) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *CommandTemplate) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCommandTemplateVersion {
		return false, nil
	}
	if setting.Version > CurrentCommandTemplateVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCommandTemplateVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsCommandTemplate() (*CommandTemplate, error) {
	return parseSettingAs[*CommandTemplate](s)
}

func (s *Setting) MustAsCommandTemplate() *CommandTemplate {
	return gofn.Must(s.AsCommandTemplate())
}
