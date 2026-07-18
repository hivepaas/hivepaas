package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentImageBuildSettingsVersion = 1

	defaultCPUPeriod = 100000
)

var _ = registerSettingParser(base.SettingTypeImageBuildSettings, &imageBuildSettingsParser{})

type imageBuildSettingsParser struct {
}

func (s *imageBuildSettingsParser) New() SettingData {
	return &ImageBuildSettings{}
}

type ImageBuildSettings struct {
	Resources ImageBuildResourceSettings `json:"resources"`
	Sources   ImageBuildSourceSettings   `json:"sources"`
	NoCache   bool                       `json:"noCache,omitempty"`
	NoVerbose bool                       `json:"noVerbose,omitempty"`
}

type ImageBuildResourceSettings struct {
	CPUs    uint          `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap,omitempty"`
	ShmSize unit.DataSize `json:"shmSize,omitempty"`
}

type ImageBuildSourceSettings struct {
	RepoCache bool `json:"repoCache"`
}

// CPUsAsPeriodAndQuota calculates CPU period and quota from CPUs
// Ref: https://docs.docker.com/engine/containers/resource_constraints
func (s *ImageBuildResourceSettings) CPUsAsPeriodAndQuota() (int64, int64) {
	if s.CPUs == 0 {
		return 0, 0
	}
	return defaultCPUPeriod, int64(defaultCPUPeriod * s.CPUs) //nolint
}

func (s *ImageBuildSettings) GetType() base.SettingType {
	return base.SettingTypeImageBuildSettings
}

func (s *ImageBuildSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ImageBuildSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ImageBuildSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentImageBuildSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentImageBuildSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentImageBuildSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsImageBuildSettings() (*ImageBuildSettings, error) {
	return parseSettingAs[*ImageBuildSettings](s)
}

func (s *Setting) MustAsImageBuildSettings() *ImageBuildSettings {
	return gofn.Must(s.AsImageBuildSettings())
}
