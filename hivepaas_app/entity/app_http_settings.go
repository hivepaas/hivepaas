package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/traefik"
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
	ExposePublicly bool         `json:"exposePublicly"`
	Domains        []*AppDomain `json:"domains,omitempty"`
	Reset          bool         `json:"reset,omitempty"`
}

type AppDomain struct {
	Enabled           bool                   `json:"enabled"`
	Domain            string                 `json:"domain"`
	DomainRedirect    string                 `json:"domainRedirect,omitempty"`
	SSLCert           ObjectID               `json:"sslCert,omitzero"`
	ContainerPort     int                    `json:"containerPort,omitempty"`
	ForceHttps        bool                   `json:"forceHttps,omitempty"`
	LBConfig          *HTTPLBConfig          `json:"lbConfig,omitempty"`
	BasicAuth         *HTTPBasicAuthConfig   `json:"basicAuth,omitempty"`
	ClientConfig      *HTTPClientConfig      `json:"clientConfig,omitempty"`
	CompressionConfig *HTTPCompressionConfig `json:"compressionConfig,omitempty"`
	RateLimitConfig   *HTTPRateLimitConfig   `json:"rateLimitConfig,omitempty"`
	HeaderConfig      *HTTPHeaderConfig      `json:"headerConfig,omitempty"`
	PathRewriteConfig *HTTPPathRewriteConfig `json:"pathRewriteConfig,omitempty"`
	Paths             []*HTTPPathConfig      `json:"paths,omitempty"`
}

type HTTPLBConfig struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

type HTTPBasicAuthConfig struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

type HTTPClientConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody,omitempty"`
	MemRequestBody unit.DataSize `json:"memRequestBody,omitempty"`
	AllowedIPs     []string      `json:"allowedIPs,omitempty"`
}

type HTTPHeaderConfig struct {
	Enabled               bool              `json:"enabled"`
	AutoContentType       bool              `json:"autoContentType,omitempty"`
	ToAddToRequests       map[string]string `json:"toAddToRequests,omitempty"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests,omitempty"`
	ToAddToResponses      map[string]string `json:"toAddToResponses,omitempty"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses,omitempty"`
}

type HTTPCompressionConfig struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes,omitempty"`
	IncludedContentTypes []string      `json:"includedContentTypes,omitempty"`
	MinResponseBody      unit.DataSize `json:"minResponseBody,omitempty"`
	DefaultEncoding      string        `json:"defaultEncoding,omitempty"`
}

type HTTPRateLimitConfig struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average,omitempty"`
	Period         timeutil.Duration `json:"period,omitempty"`
	Burst          int               `json:"burst,omitempty"`
	MaxInFlightReq int               `json:"maxInFlightReq,omitempty"`
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

type HTTPPathConfig struct {
	Enabled           bool                   `json:"enabled"`
	Path              string                 `json:"path"`
	Mode              base.HTTPPathMode      `json:"mode"`
	BasicAuth         *HTTPBasicAuthConfig   `json:"basicAuth,omitempty"`
	ClientConfig      *HTTPClientConfig      `json:"clientConfig,omitempty"`
	RateLimitConfig   *HTTPRateLimitConfig   `json:"rateLimitConfig,omitempty"`
	HeaderConfig      *HTTPHeaderConfig      `json:"headerConfig,omitempty"`
	CompressionConfig *HTTPCompressionConfig `json:"compressionConfig,omitempty"`
	PathRewriteConfig *HTTPPathRewriteConfig `json:"pathRewriteConfig,omitempty"`
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

func (s *AppHttpSettings) GetActiveDomainNames() (res []string) {
	if !s.ExposePublicly {
		return
	}
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		res = append(res, domain.Domain)
	}
	return res
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

func (s *AppHttpSettings) CalcResLinks(setting *Setting) []*ResLink {
	resLinks := s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)

	// Links domain to the current setting
	timeNow := timeutil.NowUTC()
	for _, domain := range s.GetActiveDomainNames() {
		if setting.ObjectID == "" {
			continue
		}
		resLinks = append(resLinks, &ResLink{
			SrcType:   base.ResourceTypeSetting,
			SrcID:     setting.ID,
			DstType:   base.ResourceTypeDomain,
			DstID:     domain,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	return resLinks
}

func (s *AppHttpSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppHttpSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentAppHttpSettingsVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
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
