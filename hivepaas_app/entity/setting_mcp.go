package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentMCPSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeMCP, &mcpSettingsParser{})

type mcpSettingsParser struct {
}

// New is the never-configured default: off, as the server is until an
// administrator turns it on.
func (s *mcpSettingsParser) New() SettingData {
	return &MCPSettings{}
}

// MCPSettings is whether HivePaaS serves the Model Context Protocol, and what
// its tools may do. It is global. See
// docs/superpowers/specs/2026-09-26-mcp-server-design.md.
type MCPSettings struct {
	// Enabled serves <API base path>/mcp. Off, the path answers 404.
	Enabled bool `json:"enabled"`
	// AllowWrite lists the tools that change things. They come in phase 2; until
	// then nothing reads it.
	AllowWrite bool `json:"allowWrite"`
}

func (s *MCPSettings) GetType() base.SettingType {
	return base.SettingTypeMCP
}

func (s *MCPSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *MCPSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsMCPSettings() (*MCPSettings, error) {
	return parseSettingAs[*MCPSettings](s)
}

func (s *Setting) MustAsMCPSettings() *MCPSettings {
	return gofn.Must(s.AsMCPSettings())
}
