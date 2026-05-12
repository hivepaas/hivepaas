package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentLocalPaaSSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeLocalPaaSSettings, &localPaaSSettingsParser{})

type localPaaSSettingsParser struct {
}

func (s *localPaaSSettingsParser) New() SettingData {
	return &LocalPaaSSettings{}
}

type LocalPaaSSettings struct {
	WorkerSettings      *LocalPaaSWorkerSettings      `json:"workerSettings"`
	TaskSettings        *LocalPaaSTaskSettings        `json:"taskSettings"`
	HealthcheckSettings *LocalPaaSHealthcheckSettings `json:"healthcheckSettings"`
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

func (s *LocalPaaSSettings) GetType() base.SettingType {
	return base.SettingTypeLocalPaaSSettings
}

func (s *LocalPaaSSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *LocalPaaSSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentLocalPaaSSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentLocalPaaSSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentLocalPaaSSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsLocalPaaSSettings() (*LocalPaaSSettings, error) {
	return parseSettingAs[*LocalPaaSSettings](s)
}

func (s *Setting) MustAsLocalPaaSSettings() *LocalPaaSSettings {
	return gofn.Must(s.AsLocalPaaSSettings())
}
