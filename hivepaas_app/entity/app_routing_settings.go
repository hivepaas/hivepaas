package entity

import (
	"strconv"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/traefik"
)

const (
	CurrentAppRoutingSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppRouting, &appRoutingSettingsParser{})

type appRoutingSettingsParser struct {
}

func (s *appRoutingSettingsParser) New() SettingData {
	return &AppRoutingSettings{}
}

type AppRoutingSettings struct {
	Port           int          `json:"port"`
	ExposePublicly bool         `json:"exposePublicly"`
	Domains        []*AppDomain `json:"domains,omitempty"`
	Reset          bool         `json:"reset,omitempty"`
}

type AppDomain struct {
	Enabled        bool                 `json:"enabled"`
	Domain         string               `json:"domain"`
	Protocol       base.NetworkProtocol `json:"protocol"`
	ContainerPort  int                  `json:"containerPort,omitempty"`
	TLSPassthrough bool                 `json:"tlsPassthrough,omitempty"`
	SSLCert        ObjectID             `json:"sslCert,omitzero"`

	// HTTP (layer 7) configuration
	DomainRedirect       string                    `json:"domainRedirect,omitempty"`
	ForceHttps           bool                      `json:"forceHttps,omitempty"`
	LBConfig             *HTTPLBConfig             `json:"lbConfig,omitempty"`
	BasicAuth            *HTTPBasicAuthConfig      `json:"basicAuth,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfig `json:"circuitBreakerConfig,omitempty"`
	ClientConfig         *HTTPClientConfig         `json:"clientConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfig    `json:"compressionConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfig         `json:"headerConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfig    `json:"pathRewriteConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfig      `json:"rateLimitConfig,omitempty"`
	WebsocketConfig      *HTTPWebsocketConfig      `json:"websocketConfig,omitempty"`
	Paths                []*HTTPPathConfig         `json:"paths,omitempty"`
}

type HTTPLBConfig struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

type HTTPBasicAuthConfig struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

type HTTPCircuitBreakerConfig struct {
	Enabled          bool              `json:"enabled"`
	Expression       string            `json:"expression,omitempty"`
	CheckPeriod      timeutil.Duration `json:"checkPeriod,omitempty"`
	FallbackDuration timeutil.Duration `json:"fallbackDuration,omitempty"`
	RecoveryDuration timeutil.Duration `json:"recoveryDuration,omitempty"`
	ResponseCode     int               `json:"responseCode,omitempty"`
}

type HTTPClientConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody,omitempty"`
	MemRequestBody unit.DataSize `json:"memRequestBody,omitempty"`
	AllowedIPs     []string      `json:"allowedIPs,omitempty"`
}

type HTTPCompressionConfig struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes,omitempty"`
	IncludedContentTypes []string      `json:"includedContentTypes,omitempty"`
	MinResponseBody      unit.DataSize `json:"minResponseBody,omitempty"`
	DefaultEncoding      string        `json:"defaultEncoding,omitempty"`
}

type HTTPHeaderConfig struct {
	Enabled               bool              `json:"enabled"`
	AutoContentType       bool              `json:"autoContentType,omitempty"`
	ToAddToRequests       map[string]string `json:"toAddToRequests,omitempty"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests,omitempty"`
	ToAddToResponses      map[string]string `json:"toAddToResponses,omitempty"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses,omitempty"`
}

type HTTPPathRewriteConfig struct {
	Enabled            bool   `json:"enabled"`
	PrefixAdd          string `json:"prefixAdd,omitempty"`
	PrefixStrip        string `json:"prefixStrip,omitempty"`
	PrefixStripIsRegex bool   `json:"prefixStripIsRegex,omitempty"`
	PathReplace        string `json:"pathReplace,omitempty"`
	PathReplaceIsRegex bool   `json:"pathReplaceIsRegex,omitempty"`
	PathReplaceWith    string `json:"pathReplaceWith,omitempty"`
}

type HTTPRateLimitConfig struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average,omitempty"`
	Period         timeutil.Duration `json:"period,omitempty"`
	Burst          int               `json:"burst,omitempty"`
	MaxInFlightReq int               `json:"maxInFlightReq,omitempty"`
}

type HTTPWebsocketConfig struct {
	Enabled bool `json:"enabled"`
}

type HTTPPathConfig struct {
	Enabled              bool                      `json:"enabled"`
	Path                 string                    `json:"path"`
	Mode                 base.HTTPPathMode         `json:"mode"`
	BasicAuth            *HTTPBasicAuthConfig      `json:"basicAuth,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfig `json:"circuitBreakerConfig,omitempty"`
	ClientConfig         *HTTPClientConfig         `json:"clientConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfig    `json:"compressionConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfig         `json:"headerConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfig    `json:"pathRewriteConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfig      `json:"rateLimitConfig,omitempty"`
	WebsocketConfig      *HTTPWebsocketConfig      `json:"websocketConfig,omitempty"`
}

func (s *AppRoutingSettings) GetDomain(domain string) *AppDomain {
	domain = strings.ToLower(domain)
	for _, domainRec := range s.Domains {
		if domainRec.Domain == domain {
			return domainRec
		}
	}
	return nil
}

// UsesClientIP reports whether any middleware in these settings has to decide
// which address a request came from.
//
// It is the filter for two things that must agree. Traefik's ip-strategy depth is
// written into an app's labels only when this is true, so it is also the set of
// apps a change to the proxy topology can affect - see the sweep in
// approutingservice. Answering it in two places would let the sweep skip an app
// whose labels do carry a depth, and that app would then keep reading the wrong
// forwarded position until something unrelated redeployed it.
func (s *AppRoutingSettings) UsesClientIP() bool {
	if s == nil || !s.ExposePublicly {
		return false
	}

	for _, domain := range s.Domains {
		if usesClientIP(domain.RateLimitConfig, domain.ClientConfig) {
			return true
		}
		for _, pathCfg := range domain.Paths {
			if usesClientIP(pathCfg.RateLimitConfig, pathCfg.ClientConfig) {
				return true
			}
		}
	}
	return false
}

func usesClientIP(rateLimit *HTTPRateLimitConfig, client *HTTPClientConfig) bool {
	if rateLimit != nil && rateLimit.Enabled {
		return true
	}
	return client != nil && client.Enabled && len(client.AllowedIPs) > 0
}

func (s *AppRoutingSettings) GetActiveDomains() (res []*AppDomain) {
	if s == nil || !s.ExposePublicly {
		return
	}
	res = make([]*AppDomain, 0, len(s.Domains))
	for _, domain := range s.Domains {
		if domain != nil && domain.Enabled && domain.Domain != "" {
			res = append(res, domain)
		}
	}
	return res
}

func (s *AppRoutingSettings) GetActiveDomainNames() (res []string) {
	activeDomains := s.GetActiveDomains()
	res = make([]string, 0, len(activeDomains))
	for _, domain := range activeDomains {
		res = append(res, domain.Domain)
	}
	return res
}

func (s *AppRoutingSettings) GetActivePorts() []int {
	activePorts := []int{s.Port}
	for _, domain := range s.GetActiveDomains() {
		if gofn.Contain(activePorts, domain.ContainerPort) {
			continue
		}
		activePorts = append(activePorts, domain.ContainerPort)
	}
	return activePorts
}

func (s *AppRoutingSettings) GetType() base.SettingType {
	return base.SettingTypeAppRouting
}

func (s *AppRoutingSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.GetSSLCertIDs(), s.GetBasicAuthIDs()),
	}
}

func (s *AppRoutingSettings) GetSSLCertIDs() (res []string) {
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		if domain.SSLCert.ID != "" {
			res = append(res, domain.SSLCert.ID)
		}
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppRoutingSettings) GetBasicAuthIDs() (res []string) {
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		if domain.BasicAuth != nil && domain.BasicAuth.ID != "" {
			res = append(res, domain.BasicAuth.ID)
		}
		for _, pathConfig := range domain.Paths {
			if pathConfig.BasicAuth != nil && pathConfig.BasicAuth.ID != "" {
				res = append(res, pathConfig.BasicAuth.ID)
			}
		}
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppRoutingSettings) GetResourceLinks(setting *Setting) []*ResLink {
	resLinks := s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)

	// Links domains to the current setting
	timeNow := timeutil.NowUTC()
	for i, domain := range s.GetActiveDomainNames() {
		if setting.ObjectID == "" {
			continue
		}
		resLinks = append(resLinks, &ResLink{
			SrcType:   base.ResourceTypeSetting,
			SrcID:     setting.ID,
			DstType:   base.ResourceTypeDomain,
			DstID:     domain,
			Index:     i,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	// Links ports
	for i, port := range s.GetActivePorts() {
		resLinks = append(resLinks, &ResLink{
			SrcType:   base.ResourceTypeSetting,
			SrcID:     setting.ID,
			DstType:   base.ResourceTypePort,
			DstID:     strconv.Itoa(port),
			Index:     i,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	return resLinks
}

func (s *Setting) AsAppRoutingSettings() (*AppRoutingSettings, error) {
	return parseSettingAs[*AppRoutingSettings](s)
}

func (s *Setting) MustAsAppRoutingSettings() *AppRoutingSettings {
	return gofn.Must(s.AsAppRoutingSettings())
}
