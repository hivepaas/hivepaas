package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentAppHttpSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppHttp, &appHttpSettingsParser{})

type appHttpSettingsParser struct {
}

func (s *appHttpSettingsParser) New() SettingData {
	return &AppHttpSettings{}
}

type AppHttpSettings struct {
	Enabled bool         `json:"enabled"`
	Domains []*AppDomain `json:"domains,omitempty"`
	Reset   bool         `json:"reset,omitempty"`
}

type AppDomain struct {
	Enabled           bool                   `json:"enabled"`
	Domain            string                 `json:"domain"`
	DomainRedirect    string                 `json:"domainRedirect,omitempty"`
	SSLCert           ObjectID               `json:"sslCert,omitzero"`
	ContainerPort     int                    `json:"containerPort,omitempty"`
	ForceHttps        bool                   `json:"forceHttps,omitempty"`
	BasicAuth         ObjectID               `json:"basicAuth,omitzero"`
	ClientConfig      *HTTPClientConfig      `json:"clientConfig,omitempty"`
	CompressionConfig *HTTPCompressionConfig `json:"compressionConfig,omitempty"`
	RateLimitConfig   *HTTPRateLimitConfig   `json:"rateLimitConfig,omitempty"`
	HeaderConfig      *HTTPHeaderConfig      `json:"headerConfig,omitempty"`
	Paths             []*HTTPPathConfig      `json:"paths,omitempty"`
}

type HTTPClientConfig struct {
	Enabled             bool     `json:"enabled"`
	MaxRequestBodyBytes int      `json:"maxRequestBodyBytes,omitempty"`
	MemRequestBodyBytes int      `json:"memRequestBodyBytes,omitempty"`
	AllowedIPs          []string `json:"allowedIPs,omitempty"`
}

type HTTPHeaderConfig struct {
	ToAddToRequests       map[string]string `json:"toAddToRequests,omitempty"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests,omitempty"`
	ToAddToResponses      map[string]string `json:"toAddToResponses,omitempty"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses,omitempty"`
}

type HTTPCompressionConfig struct {
	Enabled              bool     `json:"enabled"`
	ExcludedContentTypes []string `json:"excludedContentTypes,omitempty"`
	IncludedContentTypes []string `json:"includedContentTypes,omitempty"`
	MinResponseBodyBytes int      `json:"minResponseBodyBytes,omitempty"`
	DefaultEncoding      string   `json:"defaultEncoding,omitempty"`
}

type HTTPRateLimitConfig struct {
	Enabled           bool              `json:"enabled"`
	Average           int               `json:"average,omitempty"`
	Period            timeutil.Duration `json:"period,omitempty"`
	Burst             int               `json:"burst,omitempty"`
	InFlightReqAmount int               `json:"inFlightReqAmount,omitempty"`
}

type HTTPPathConfig struct {
	Path            string               `json:"path"`
	IsRegex         bool                 `json:"isRegex"`
	BasicAuth       ObjectID             `json:"basicAuth,omitzero"`
	ClientConfig    *HTTPClientConfig    `json:"clientConfig,omitempty"`
	RateLimitConfig *HTTPRateLimitConfig `json:"rateLimitConfig,omitempty"`
	HeaderConfig    *HTTPHeaderConfig    `json:"headerConfig,omitempty"`
}

func (s *AppHttpSettings) GetDomain(domain string) *AppDomain {
	domain = strings.ToLower(domain)
	for _, domainRec := range s.Domains {
		if domainRec.Domain == domain {
			return domainRec
		}
	}
	return nil
}

func (s *AppHttpSettings) GetType() base.SettingType {
	return base.SettingTypeAppHttp
}

func (s *AppHttpSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.GetSSLCertIDs(), s.GetBasicAuthIDs()),
	}
}

func (s *AppHttpSettings) GetSSLCertIDs() (res []string) {
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

func (s *AppHttpSettings) GetBasicAuthIDs() (res []string) {
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		if domain.BasicAuth.ID != "" {
			res = append(res, domain.BasicAuth.ID)
		}
		for _, pathConfig := range domain.Paths {
			if pathConfig.BasicAuth.ID != "" {
				res = append(res, pathConfig.BasicAuth.ID)
			}
		}
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppHttpSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppHttpSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppHttpSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppHttpSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAppHttpSettings() (*AppHttpSettings, error) {
	return parseSettingAs[*AppHttpSettings](s)
}

func (s *Setting) MustAsAppHttpSettings() *AppHttpSettings {
	return gofn.Must(s.AsAppHttpSettings())
}
