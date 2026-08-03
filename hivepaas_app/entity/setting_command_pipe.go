package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentCommandPipeVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCommandPipe, &commandPipeParser{})

type commandPipeParser struct {
}

func (s *commandPipeParser) New() SettingData {
	return &CommandPipe{}
}

type CommandPipe struct {
	SourceCommand ObjectID `json:"sourceCommand,omitzero"`
	TargetCommand ObjectID `json:"targetCommand,omitzero"`
}

func (s *CommandPipe) GetType() base.SettingType {
	return base.SettingTypeCommandPipe
}

func (s *CommandPipe) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.SourceCommand.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.SourceCommand.ID)
	}
	if s.TargetCommand.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.TargetCommand.ID)
	}
	return refIDs
}

func (s *CommandPipe) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsCommandPipe() (*CommandPipe, error) {
	return parseSettingAs[*CommandPipe](s)
}

func (s *Setting) MustAsCommandPipe() *CommandPipe {
	return gofn.Must(s.AsCommandPipe())
}
