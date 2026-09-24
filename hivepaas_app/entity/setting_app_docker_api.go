package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppDockerAPIVersion = 1

	// DockerAPINetworkEnv, in AppDockerAPISettings.Networks, is the app's own
	// project-env network.
	DockerAPINetworkEnv = "env"
)

var _ = registerSettingParser(base.SettingTypeAppDockerAPI, &appDockerAPISettingsParser{})

type appDockerAPISettingsParser struct {
}

func (s *appDockerAPISettingsParser) New() SettingData {
	return &AppDockerAPISettings{}
}

// AppDockerAPISettings is what an app may do through the Docker API HivePaaS
// serves it: the images its children run, the directories they share with it,
// the networks they join, and the endpoints beyond the core. An app without this
// setting has no socket at all. See
// docs/superpowers/specs/2026-09-24-docker-api-access-design.md.
type AppDockerAPISettings struct {
	// Images are patterns over what children may run; "*" is any image.
	Images []string `json:"images"`
	// SharedDirs are directories of the app's own storage a child may bind.
	SharedDirs []string `json:"sharedDirs,omitempty"`
	// Networks are networks children may join besides their own.
	Networks []string `json:"networks,omitempty"`
	// Allow are groups of endpoints beyond the core, as the proxy names them.
	Allow  []string           `json:"allow,omitempty"`
	Limits AppDockerAPILimits `json:"limits,omitzero"`
}

// AppDockerAPILimits bound what the app's children use. Zero is the default.
type AppDockerAPILimits struct {
	// Containers is how many children may exist at once.
	Containers int `json:"containers,omitempty"`
	// Memory is the most one child may have, in bytes.
	Memory int64 `json:"memory,omitempty"`
	// NanoCPUs is the most processor time one child may have, in billionths of
	// a CPU.
	NanoCPUs int64 `json:"nanoCpus,omitempty"`
}

func (s *AppDockerAPISettings) GetType() base.SettingType {
	return base.SettingTypeAppDockerAPI
}

func (s *AppDockerAPISettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppDockerAPISettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppDockerAPISettings() (*AppDockerAPISettings, error) {
	return parseSettingAs[*AppDockerAPISettings](s)
}

func (s *Setting) MustAsAppDockerAPISettings() *AppDockerAPISettings {
	return gofn.Must(s.AsAppDockerAPISettings())
}
