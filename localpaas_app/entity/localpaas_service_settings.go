package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentLocalPaaSServiceVersion = 1
)

var _ = registerSettingParser(base.SettingTypeLocalPaaSService, &localPaaSServiceParser{})

type localPaaSServiceParser struct {
}

func (s *localPaaSServiceParser) New() SettingData {
	return &LocalPaaSService{}
}

type LocalPaaSService struct {
	AppSettings         LocalPaaSAppSettings         `json:"appSettings"`
	WorkerSettings      LocalPaaSWorkerSettings      `json:"workerSettings"`
	TaskSettings        LocalPaaSTaskSettings        `json:"taskSettings"`
	HealthcheckSettings LocalPaaSHealthcheckSettings `json:"healthcheckSettings"`
}

type LocalPaaSAppSettings struct {
	Replicas int `json:"replicas,omitempty"`
}

type LocalPaaSWorkerSettings struct {
	Replicas           int  `json:"replicas,omitempty"`
	Concurrency        int  `json:"concurrency,omitempty"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp,omitempty"`
}

type LocalPaaSTaskSettings struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

type LocalPaaSHealthcheckSettings struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
}

func (s *LocalPaaSService) GetType() base.SettingType {
	return base.SettingTypeLocalPaaSService
}

func (s *LocalPaaSService) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *LocalPaaSService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentLocalPaaSServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentLocalPaaSServiceVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentLocalPaaSServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsLocalPaaSService() (*LocalPaaSService, error) {
	return parseSettingAs[*LocalPaaSService](s)
}

func (s *Setting) MustAsLocalPaaSService() *LocalPaaSService {
	return gofn.Must(s.AsLocalPaaSService())
}
