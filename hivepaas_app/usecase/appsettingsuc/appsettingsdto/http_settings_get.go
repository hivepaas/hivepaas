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

type GetAppHttpSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppHttpSettingsReq() *GetAppHttpSettingsReq {
	return &GetAppHttpSettingsReq{}
}

func (req *GetAppHttpSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppHttpSettingsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *HttpSettingsResp `json:"data"`
}

type HttpSettingsResp struct {
	DomainSuggestion string        `json:"domainSuggestion"`
	Port             int           `json:"port"`
	ExposePublicly   bool          `json:"exposePublicly"`
	Domains          []*DomainResp `json:"domains"`
	UpdateVer        int           `json:"updateVer"`
}

type DomainResp struct {
	Enabled              bool                          `json:"enabled"`
	Domain               string                        `json:"domain"`
	DomainRedirect       string                        `json:"domainRedirect,omitempty"`
	SSLCert              *sslcertdto.SSLCertResp       `json:"sslCert,omitempty"`
	ContainerPort        int                           `json:"containerPort"`
	ForceHttps           bool                          `json:"forceHttps,omitempty"`
	LBConfig             *HTTPLBConfigResp             `json:"lbConfig,omitempty"`
	BasicAuth            *HTTPBasicAuthConfigResp      `json:"basicAuth,omitempty"`
	ClientConfig         *HTTPClientConfigResp         `json:"clientConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfigResp         `json:"headerConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfigResp    `json:"compressionConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfigResp      `json:"rateLimitConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfigResp    `json:"pathRewriteConfig,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigResp `json:"circuitBreakerConfig,omitempty"`
	Paths                []*HTTPPathConfigResp         `json:"paths,omitempty"`
}

type HTTPLBConfigResp struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

type HTTPBasicAuthConfigResp struct {
	Enabled bool `json:"enabled"`
	*settings.BaseSettingResp
}

type HTTPClientConfigResp struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

type HTTPHeaderConfigResp struct {
	Enabled               bool              `json:"enabled"`
	AutoContentType       bool              `json:"autoContentType,omitempty"`
	ToAddToRequests       map[string]string `json:"toAddToRequests"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests"`
	ToAddToResponses      map[string]string `json:"toAddToResponses"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses"`
}

type HTTPRateLimitConfigResp struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

type HTTPCompressionConfigResp struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes"`
	IncludedContentTypes []string      `json:"includedContentTypes"`
	MinResponseBody      unit.DataSize `json:"minResponseBody"`
	DefaultEncoding      string        `json:"defaultEncoding"`
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

type HTTPCircuitBreakerConfigResp struct {
	Enabled          bool              `json:"enabled"`
	Expression       string            `json:"expression,omitempty"`
	CheckPeriod      timeutil.Duration `json:"checkPeriod,omitempty"`
	FallbackDuration timeutil.Duration `json:"fallbackDuration,omitempty"`
	RecoveryDuration timeutil.Duration `json:"recoveryDuration,omitempty"`
	ResponseCode     int               `json:"responseCode,omitempty"`
}

type HTTPPathConfigResp struct {
	Enabled              bool                          `json:"enabled"`
	Path                 string                        `json:"path"`
	Mode                 base.HTTPPathMode             `json:"mode"`
	BasicAuth            *HTTPBasicAuthConfigResp      `json:"basicAuth,omitempty"`
	ClientConfig         *HTTPClientConfigResp         `json:"clientConfig,omitempty"`
	HeaderConfig         *HTTPHeaderConfigResp         `json:"headerConfig,omitempty"`
	CompressionConfig    *HTTPCompressionConfigResp    `json:"compressionConfig,omitempty"`
	RateLimitConfig      *HTTPRateLimitConfigResp      `json:"rateLimitConfig,omitempty"`
	PathRewriteConfig    *HTTPPathRewriteConfigResp    `json:"pathRewriteConfig,omitempty"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigResp `json:"circuitBreakerConfig,omitempty"`
}

type AppHttpSettingsTransformInput struct {
	App           *entity.App
	HttpSettings  *entity.Setting
	RefSettingMap map[string]*entity.Setting
}

func TransformHttpSettings(input *AppHttpSettingsTransformInput) (resp *HttpSettingsResp, err error) {
	resp = &HttpSettingsResp{}
	resp.DomainSuggestion = fmt.Sprintf("<name>.%v", config.Current.RootDomain)

	if input.HttpSettings == nil {
		return resp, nil
	}

	if err = copier.Copy(&resp, input.HttpSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}
	appHttpSettings := input.HttpSettings.MustAsAppHttpSettings()
	if err = copier.Copy(&resp, appHttpSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefSettingMap[domain.SSLCert.ID]
			domain.SSLCert, _ = sslcertdto.TransformSSLCertBasic(setting, &entity.RefObjects{})
		} else {
			domain.SSLCert = nil
		}
		if domain.BasicAuth != nil && domain.BasicAuth.ID != "" {
			setting := input.RefSettingMap[domain.BasicAuth.ID]
			domain.BasicAuth.BaseSettingResp, _ = settings.TransformSettingBase(setting)
		} else {
			domain.BasicAuth = nil
		}

		for _, pathConfig := range domain.Paths {
			setting := input.RefSettingMap[pathConfig.BasicAuth.ID]
			if pathConfig.BasicAuth != nil && pathConfig.BasicAuth.ID != "" {
				pathConfig.BasicAuth.BaseSettingResp, _ = settings.TransformSettingBase(setting)
			} else {
				pathConfig.BasicAuth = nil
			}
		}
	}

	return resp, nil
}
