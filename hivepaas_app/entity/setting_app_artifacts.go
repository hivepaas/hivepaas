package entity

import (
	"maps"
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppArtifactsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppArtifacts, &appArtifactsSettingsParser{})

type appArtifactsSettingsParser struct {
}

func (s *appArtifactsSettingsParser) New() SettingData {
	return &AppArtifactsSettings{}
}

// AppArtifactsSettings are the files HivePaaS mounts into an app's container
// beside its image: a managed template's artifacts, as the app was created with
// them. Each deploy has them on the node the app runs on, by hash, and binds
// them read-only. The mounts are HivePaaS's, not the app's: the storage screen
// does not show them, and a save there keeps them.
//
// See docs/superpowers/specs/2026-10-17-managed-templates-design.md §2.
type AppArtifactsSettings struct {
	Artifacts []*AppArtifact `json:"artifacts"`
}

// AppArtifact is one file, with the build for each architecture a node can be.
type AppArtifact struct {
	Name      string                            `json:"name"`
	Version   string                            `json:"version"`
	MountPath string                            `json:"mountPath"`
	Files     map[base.CPUArch]*AppArtifactFile `json:"files"`
}

type AppArtifactFile struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// SHA256s are the hashes of every file the app may run, on whichever node: what
// a node's cleanup keeps.
func (s *AppArtifactsSettings) SHA256s() []string {
	if s == nil {
		return nil
	}
	var hashes []string
	for _, artifact := range s.Artifacts {
		if artifact == nil {
			continue
		}
		for _, arch := range slices.Sorted(maps.Keys(artifact.Files)) {
			if file := artifact.Files[arch]; file != nil {
				hashes = append(hashes, file.SHA256)
			}
		}
	}
	return hashes
}

func (s *AppArtifactsSettings) GetType() base.SettingType {
	return base.SettingTypeAppArtifacts
}

func (s *AppArtifactsSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppArtifactsSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppArtifactsSettings() (*AppArtifactsSettings, error) {
	return parseSettingAs[*AppArtifactsSettings](s)
}

func (s *Setting) MustAsAppArtifactsSettings() *AppArtifactsSettings {
	return gofn.Must(s.AsAppArtifactsSettings())
}
