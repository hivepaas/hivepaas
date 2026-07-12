package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentHivePaaSServiceVersion = 1
)

var _ = registerSettingParser(base.SettingTypeHivePaaSService, &hivePaaSServiceParser{})

type hivePaaSServiceParser struct {
}

func (s *hivePaaSServiceParser) New() SettingData {
	return &HivePaaSService{}
}

type HivePaaSService struct {
	AppSettings         HivePaaSAppSettings         `json:"appSettings"`
	WorkerSettings      HivePaaSWorkerSettings      `json:"workerSettings"`
	TaskSettings        HivePaaSTaskSettings        `json:"taskSettings"`
	HealthcheckSettings HivePaaSHealthcheckSettings `json:"healthcheckSettings"`
}

type HivePaaSAppSettings struct {
	Replicas int `json:"replicas,omitempty"`
}

type HivePaaSWorkerSettings struct {
	Replicas           int  `json:"replicas,omitempty"`
	Concurrency        int  `json:"concurrency,omitempty"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp,omitempty"`
}

type HivePaaSTaskSettings struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

type HivePaaSHealthcheckSettings struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

func (s *HivePaaSService) GetType() base.SettingType {
	return base.SettingTypeHivePaaSService
}

func (s *HivePaaSService) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *HivePaaSService) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *HivePaaSService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentHivePaaSServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentHivePaaSServiceVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentHivePaaSServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsHivePaaSService() (*HivePaaSService, error) {
	return parseSettingAs[*HivePaaSService](s)
}

func (s *Setting) MustAsHivePaaSService() *HivePaaSService {
	return gofn.Must(s.AsHivePaaSService())
}
