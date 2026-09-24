package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentAppDockerAPIVersion = 1

	// DockerAPINetworkEnv, in AppDockerAPISettings.Networks, is the app's own
	// project-env network.
	DockerAPINetworkEnv = "env"

	// DockerAPIModeProxy, the default, gives the app the proxy HivePaaS runs.
	DockerAPIModeProxy = "proxy"
	// DockerAPIModeHost gives the app the node's own socket, which only an
	// administrator can choose, with the privileged-apps switch on.
	DockerAPIModeHost = "host"
)

var _ = registerSettingParser(base.SettingTypeAppDockerAPI, &appDockerAPISettingsParser{})

type appDockerAPISettingsParser struct {
}

func (s *appDockerAPISettingsParser) New() SettingData {
	return &AppDockerAPISettings{}
}

// AppDockerAPISettings is what an app may do through the Docker API HivePaaS
// serves it: the images its children run, the directories they share with it,
// the networks they join, and the endpoints beyond the core - or, in host mode,
// that it has the node's own socket. An app without this setting has no socket
// at all. See docs/superpowers/specs/2026-09-24-docker-api-access-design.md.
type AppDockerAPISettings struct {
	// Mode is DockerAPIModeProxy, empty for it, or DockerAPIModeHost. In host
	// mode the fields below are kept, for going back to the proxy, and not used.
	Mode string `json:"mode,omitempty"`
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

// AppDockerAPILimits bound what the app's children use, in the units the
// deployment's own resource limits take. Zero is the default.
type AppDockerAPILimits struct {
	// Containers is how many children may exist at once.
	Containers int `json:"containers,omitempty"`
	// Memory is the most one child may have.
	Memory unit.DataSize `json:"memory,omitempty"`
	// CPUs is the most processor time one child may have.
	CPUs float64 `json:"cpus,omitempty"`
}

// IsHostMode reports access to the node's own socket. Nil is no access at all.
func (s *AppDockerAPISettings) IsHostMode() bool {
	return s != nil && s.Mode == DockerAPIModeHost
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
