package entity

import (
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentHivePaaSServiceVersion = 2
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

// LegacyProxyHops is the depth that used to be hard-coded in the Traefik label
// builder, before the topology became something the operator states. Version 2 of
// this setting writes it into any install that already had a proxy configured, so
// upgrading does not change how any existing deployment identifies its callers.
const LegacyProxyHops = 2

type HivePaaSProxySettings struct {
	ProxyProvider string   `json:"proxyProvider,omitempty"`
	TrustedIPs    []string `json:"trustedIPs,omitempty"`

	// ProxyHops is passed to Traefik as ipStrategy.depth: the position, counted
	// from the right of X-Forwarded-For, that holds the real caller.
	//
	// It cannot be derived from the number of proxies alone - whether a proxy adds
	// itself to the chain is up to that proxy - so it is measured rather than
	// guessed. GET /system/hivepaas/request-info reports what this install
	// actually receives and what depth follows from it.
	ProxyHops int `json:"proxyHops,omitempty"`
}

// HasProxy reports whether a proxy is declared in front of Traefik.
//
// The provider name is the declaration; the other two fields are what makes it
// actionable. Validation keeps them together, so anything past validation either
// has all three or none.
// Equal reports whether two proxy configurations would be read the same way.
//
// It is what decides whether a settings change goes on trial, so it covers every
// field that feeds the client-address calculation. A field added to the struct and
// forgotten here would apply silently, with no way back.
func (s *HivePaaSProxySettings) Equal(other *HivePaaSProxySettings) bool {
	if s == nil || other == nil {
		return s == other
	}
	if s.ProxyProvider != other.ProxyProvider || s.ProxyHops != other.ProxyHops {
		return false
	}
	return slices.Equal(s.TrustedIPs, other.TrustedIPs)
}

func (s *HivePaaSProxySettings) HasProxy() bool {
	return s != nil && s.ProxyProvider != ""
}

// ClientIPDepth is the X-Forwarded-For depth that identifies the caller, or 0
// when there is no proxy and the peer address is the caller.
//
// Zero is not a weaker answer, it is a different one: Traefik reads depth 0 as
// "use the connection's remote address", which is exactly right when nothing sits
// in front. Guessing a depth with no proxy would read a header the client wrote.
func (s *HivePaaSProxySettings) ClientIPDepth() int {
	if !s.HasProxy() || s.ProxyHops <= 0 {
		return 0
	}
	return s.ProxyHops
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
