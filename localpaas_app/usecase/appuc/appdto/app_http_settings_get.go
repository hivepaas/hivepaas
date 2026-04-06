package appdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppHttpSettingsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *HttpSettingsResp `json:"data"`
}

type HttpSettingsResp struct {
	InternalEndpoints []string      `json:"internalEndpoints"`
	DomainSuggestion  string        `json:"domainSuggestion"`
	ExposePublicly    bool          `json:"exposePublicly"`
	Domains           []*DomainResp `json:"domains"`
	UpdateVer         int           `json:"updateVer"`
}

type DomainResp struct {
	Enabled           bool                       `json:"enabled"`
	Domain            string                     `json:"domain"`
	DomainRedirect    string                     `json:"domainRedirect,omitempty"`
	SSLCert           *sslcertdto.SSLCertResp    `json:"sslCert,omitempty"`
	ContainerPort     int                        `json:"containerPort"`
	ForceHttps        bool                       `json:"forceHttps,omitempty"`
	BasicAuth         *settings.BaseSettingResp  `json:"basicAuth,omitempty"`
	ClientConfig      *HTTPClientConfigResp      `json:"clientConfig,omitempty"`
	HeaderConfig      *HTTPHeaderConfigResp      `json:"headerConfig,omitempty"`
	CompressionConfig *HTTPCompressionConfigResp `json:"compressionConfig,omitempty"`
	RateLimitConfig   *HTTPRateLimitConfigResp   `json:"rateLimitConfig,omitempty"`
	Paths             []*HTTPPathConfigResp      `json:"paths,omitempty"`
}

type HTTPClientConfigResp struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

type HTTPHeaderConfigResp struct {
	ToAddToRequests       map[string]string `json:"toAddToRequests"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests"`
	ToAddToResponses      map[string]string `json:"toAddToResponses"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses"`
}

type HTTPRateLimitConfigResp struct {
	Enabled           bool              `json:"enabled"`
	Average           int               `json:"average"`
	Period            timeutil.Duration `json:"period"`
	Burst             int               `json:"burst"`
	InFlightReqAmount int               `json:"inFlightReqAmount"`
}

type HTTPCompressionConfigResp struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes"`
	IncludedContentTypes []string      `json:"includedContentTypes"`
	MinResponseBody      unit.DataSize `json:"minResponseBody"`
	DefaultEncoding      string        `json:"defaultEncoding"`
}

type HTTPPathConfigResp struct {
	Path            string                    `json:"path"`
	Mode            base.HTTPPathMode         `json:"mode"`
	BasicAuth       *settings.BaseSettingResp `json:"basicAuth,omitzero"`
	ClientConfig    *HTTPClientConfigResp     `json:"clientConfig"`
	RateLimitConfig *HTTPRateLimitConfigResp  `json:"rateLimitConfig"`
}

type AppHttpSettingsTransformInput struct {
	App           *entity.App
	HttpSettings  *entity.Setting
	RefSettingMap map[string]*entity.Setting
}

func TransformHttpSettings(input *AppHttpSettingsTransformInput) (resp *HttpSettingsResp, err error) {
	if input.HttpSettings == nil {
		return nil, nil
	}

	if err = copier.Copy(&resp, input.HttpSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}
	appHttpSettings := input.HttpSettings.MustAsAppHttpSettings()
	if err = copier.Copy(&resp, appHttpSettings); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.InternalEndpoints = []string{
		fmt.Sprintf("http://%s:<port>", input.App.Key),
	}
	resp.DomainSuggestion = fmt.Sprintf("<name>.%v", config.Current.RootDomain)

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefSettingMap[domain.SSLCert.ID]
			domain.SSLCert, _ = sslcertdto.TransformSSLCertNoCertData(setting, &entity.RefObjects{})
		} else {
			domain.SSLCert = nil
		}
		if domain.BasicAuth != nil && domain.BasicAuth.ID != "" {
			setting := input.RefSettingMap[domain.BasicAuth.ID]
			domain.BasicAuth, _ = settings.TransformSettingBase(setting)
		} else {
			domain.BasicAuth = nil
		}

		for _, pathConfig := range domain.Paths {
			setting := input.RefSettingMap[pathConfig.BasicAuth.ID]
			if pathConfig.BasicAuth != nil && pathConfig.BasicAuth.ID != "" {
				pathConfig.BasicAuth, _ = settings.TransformSettingBase(setting)
			} else {
				pathConfig.BasicAuth = nil
			}
		}
	}

	return resp, nil
}
