package entity

import (
	"github.com/tiendc/gofn"

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
	AppSettings      HivePaaSAppSettings      `json:"appSettings"`
	WorkerSettings   HivePaaSWorkerSettings   `json:"workerSettings"`
	TaskSettings     HivePaaSTaskSettings     `json:"taskSettings"`
	PeriodicSettings HivePaaSPeriodicSettings `json:"periodicSettings"`
	ProxySettings    HivePaaSProxySettings    `json:"proxySettings"`
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

type HivePaaSPeriodicSettings struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
	BatchSize    int               `json:"batchSize,omitempty"`
}

type HivePaaSProxySettings struct {
	ProxyProvider string   `json:"proxyProvider,omitempty"`
	TrustedIPs    []string `json:"trustedIPs,omitempty"`
}

func (s *HivePaaSService) GetType() base.SettingType {
	return base.SettingTypeHivePaaSService
}

func (s *HivePaaSService) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *HivePaaSService) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsHivePaaSService() (*HivePaaSService, error) {
	return parseSettingAs[*HivePaaSService](s)
}

func (s *Setting) MustAsHivePaaSService() *HivePaaSService {
	return gofn.Must(s.AsHivePaaSService())
}
