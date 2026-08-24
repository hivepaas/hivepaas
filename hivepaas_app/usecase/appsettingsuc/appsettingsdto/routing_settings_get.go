package appsettingsdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
	"github.com/hivepaas/hivepaas/services/traefik"
)

type GetAppRoutingSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppRoutingSettingsReq() *GetAppRoutingSettingsReq {
	return &GetAppRoutingSettingsReq{}
}

func (req *GetAppRoutingSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppRoutingSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *RoutingSettingsResp `json:"data"`
}

type RoutingSettingsResp struct {
	DomainSuggestion string        `json:"domainSuggestion"`
	Port             int           `json:"port"`
	ExposePublicly   bool          `json:"exposePublicly"`
	Domains          []*DomainResp `json:"domains"`
	UpdateVer        int           `json:"updateVer"`
}

type DomainResp struct {
	Enabled        bool                    `json:"enabled"`
	Domain         string                  `json:"domain"`
	Protocol       base.NetworkProtocol    `json:"protocol"`
	ContainerPort  int                     `json:"containerPort"`
	SSLCert        *sslcertdto.SSLCertResp `json:"sslCert,omitempty"`
	TLSPassthrough bool                    `json:"tlsPassthrough,omitempty"`

	// HTTP (layer 7) configuration
	DomainRedirect       string                        `json:"domainRedirect,omitempty"`
	ForceHttps           bool                          `json:"forceHttps,omitempty"`
	LBConfig             *HTTPLBConfigResp             `json:"lbConfig,omitempty"`
	BasicAuth            *HTTPBasicAuthConfigResp      `json:"basicAuth,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigResp `json:"circuitBreakerConfig,omitempty"`
	ClientConfig         *HTTPClientConfigResp         `json:"clientConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfigResp    `json:"compressionConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfigResp         `json:"headerConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfigResp    `json:"pathRewriteConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfigResp      `json:"rateLimitConfig,omitempty"`
	WebsocketConfig      *HTTPWebsocketConfigResp      `json:"websocketConfig,omitempty"`
	Paths                []*HTTPPathConfigResp         `json:"paths,omitempty"`
}

type HTTPLBConfigResp struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

type HTTPBasicAuthConfigResp struct {
	Enabled bool `json:"enabled"`
	*settings.BaseSettingResp
}

type HTTPCircuitBreakerConfigResp struct {
	Enabled          bool              `json:"enabled"`
	Expression       string            `json:"expression,omitempty"`
	CheckPeriod      timeutil.Duration `json:"checkPeriod,omitempty"`
	FallbackDuration timeutil.Duration `json:"fallbackDuration,omitempty"`
	RecoveryDuration timeutil.Duration `json:"recoveryDuration,omitempty"`
	ResponseCode     int               `json:"responseCode,omitempty"`
}

type HTTPClientConfigResp struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

type HTTPCompressionConfigResp struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes"`
	IncludedContentTypes []string      `json:"includedContentTypes"`
	MinResponseBody      unit.DataSize `json:"minResponseBody"`
	DefaultEncoding      string        `json:"defaultEncoding"`
}

type HTTPHeaderConfigResp struct {
	Enabled               bool              `json:"enabled"`
	AutoContentType       bool              `json:"autoContentType,omitempty"`
	ToAddToRequests       map[string]string `json:"toAddToRequests"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests"`
	ToAddToResponses      map[string]string `json:"toAddToResponses"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses"`
}

type HTTPPathRewriteConfigResp struct {
	Enabled            bool   `json:"enabled"`
	PrefixAdd          string `json:"prefixAdd,omitempty"`
	PrefixStrip        string `json:"prefixStrip,omitempty"`
	PrefixStripIsRegex bool   `json:"prefixStripIsRegex,omitempty"`
	PathReplace        string `json:"pathReplace,omitempty"`
	PathReplaceIsRegex bool   `json:"pathReplaceIsRegex,omitempty"`
	PathReplaceWith    string `json:"pathReplaceWith,omitempty"`
}

type HTTPRateLimitConfigResp struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

type HTTPWebsocketConfigResp struct {
	Enabled bool `json:"enabled"`
}

type HTTPPathConfigResp struct {
	Enabled              bool                          `json:"enabled"`
	Path                 string                        `json:"path"`
	Mode                 base.HTTPPathMode             `json:"mode"`
	BasicAuth            *HTTPBasicAuthConfigResp      `json:"basicAuth,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigResp `json:"circuitBreakerConfig,omitempty"`
	ClientConfig         *HTTPClientConfigResp         `json:"clientConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfigResp    `json:"compressionConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfigResp         `json:"headerConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfigResp    `json:"pathRewriteConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfigResp      `json:"rateLimitConfig,omitempty"`
	WebsocketConfig      *HTTPWebsocketConfigResp      `json:"websocketConfig,omitempty"`
}

type AppRoutingSettingsTransformInput struct {
	App             *entity.App
	RoutingSettings *entity.Setting
	RefSettingMap   map[string]*entity.Setting
}

func TransformRoutingSettings(input *AppRoutingSettingsTransformInput) (resp *RoutingSettingsResp, err error) {
	resp = &RoutingSettingsResp{}
	resp.DomainSuggestion = fmt.Sprintf("<name>.%v", config.Current.RootDomain)

	if input.RoutingSettings == nil {
		return resp, nil
	}

	if err = copier.Copy(&resp, input.RoutingSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}
	routingSettings := input.RoutingSettings.MustAsAppRoutingSettings()
	if err = copier.Copy(&resp, routingSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefSettingMap[domain.SSLCert.ID]
			certResp, _ := sslcertdto.TransformSSLCertBasic(setting, &entity.RefObjects{})
			if certResp == nil {
				certResp = &sslcertdto.SSLCertResp{
					BaseSettingResp: settings.NewMissingSetting(domain.SSLCert.ID, base.SettingTypeSSLCert),
				}
			}
			domain.SSLCert = certResp
		} else {
			domain.SSLCert = nil
		}
		if domain.BasicAuth != nil && domain.BasicAuth.ID != "" {
			itemResp, _ := settings.TransformSettingBase(input.RefSettingMap[domain.BasicAuth.ID])
			if itemResp == nil {
				itemResp = settings.NewMissingSetting(domain.BasicAuth.ID, base.SettingTypeBasicAuth)
			}
			domain.BasicAuth.BaseSettingResp = itemResp
		} else {
			domain.BasicAuth = nil
		}

		for _, pathConfig := range domain.Paths {
			if pathConfig.BasicAuth != nil && pathConfig.BasicAuth.ID != "" {
				itemResp, _ := settings.TransformSettingBase(input.RefSettingMap[pathConfig.BasicAuth.ID])
				if itemResp == nil {
					itemResp = settings.NewMissingSetting(pathConfig.BasicAuth.ID, base.SettingTypeBasicAuth)
				}
				pathConfig.BasicAuth.BaseSettingResp = itemResp
			} else {
				pathConfig.BasicAuth = nil
			}
		}
	}

	return resp, nil
}
