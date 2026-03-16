package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentImageBuildVersion = 1

	defaultCPUPeriod = 100000
)

var _ = registerSettingParser(base.SettingTypeImageBuild, &imageBuildParser{})

type imageBuildParser struct {
}

func (s *imageBuildParser) New() SettingData {
	return &ImageBuild{}
}

type ImageBuild struct {
	Resources *ImageBuildResources `json:"resources"`
	NoCache   bool                 `json:"noCache,omitempty"`
	NoVerbose bool                 `json:"noVerbose,omitempty"`
}

type ImageBuildResources struct {
	CPUs      int32 `json:"cpus"`
	MemMB     int64 `json:"memMB"`
	MemSwapMB int64 `json:"memSwapMB,omitempty"`
	ShmSizeMB int64 `json:"shmSizeMB,omitempty"`
}

// CPUsAsPeriodAndQuota calculates CPU period and quota from CPUs
// Ref: https://docs.docker.com/engine/containers/resource_constraints
func (s *ImageBuildResources) CPUsAsPeriodAndQuota() (int64, int64) {
	if s.CPUs == 0 {
		return 0, 0
	}
	return defaultCPUPeriod, int64(defaultCPUPeriod * s.CPUs)
}

func (s *ImageBuild) GetType() base.SettingType {
	return base.SettingTypeImageBuild
}

func (s *ImageBuild) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ImageBuild) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentImageBuildVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentImageBuildVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsImageBuild() (*ImageBuild, error) {
	return parseSettingAs[*ImageBuild](s)
}

func (s *Setting) MustAsImageBuild() *ImageBuild {
	return gofn.Must(s.AsImageBuild())
}
