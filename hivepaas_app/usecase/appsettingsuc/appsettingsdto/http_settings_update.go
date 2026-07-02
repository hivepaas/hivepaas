package appsettingsdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/traefik"
)

type UpdateAppHttpSettingsReq struct {
	ProjectID      string       `json:"-"`
	AppID          string       `json:"-"`
	ExposePublicly bool         `json:"exposePublicly"`
	Domains        []*DomainReq `json:"domains"`
	UpdateVer      int          `json:"updateVer"`
}

func (req *UpdateAppHttpSettingsReq) ToEntity() *entity.AppHttpSettings {
	return &entity.AppHttpSettings{
		ExposePublicly: req.ExposePublicly,
		Domains: gofn.MapSlice(req.Domains, func(r *DomainReq) *entity.AppDomain {
			return r.ToEntity()
		}),
	}
}

type DomainReq struct {
	Enabled           bool                      `json:"enabled"`
	Domain            string                    `json:"domain"`
	DomainRedirect    string                    `json:"domainRedirect"`
	SSLCert           basedto.ObjectIDReq       `json:"sslCert"`
	ContainerPort     int                       `json:"containerPort"`
	ForceHttps        bool                      `json:"forceHttps"`
	LBConfig          *HTTPLBConfigReq          `json:"lbConfig"`
	BasicAuth         *HTTPBasicAuthConfigReq   `json:"basicAuth"`
	ClientConfig      *HTTPClientConfigReq      `json:"clientConfig"`
	HeaderConfig      *HTTPHeaderConfigReq      `json:"headerConfig"`
	CompressionConfig *HTTPCompressionConfigReq `json:"compressionConfig"`
	RateLimitConfig   *HTTPRateLimitConfigReq   `json:"rateLimitConfig"`
	Paths             []*HTTPPathConfigReq      `json:"paths"`
}

func (req *DomainReq) ToEntity() *entity.AppDomain {
	return &entity.AppDomain{
		Enabled:           req.Enabled,
		Domain:            req.Domain,
		DomainRedirect:    req.DomainRedirect,
		SSLCert:           entity.ObjectID{ID: req.SSLCert.ID},
		ContainerPort:     req.ContainerPort,
		ForceHttps:        req.ForceHttps,
		LBConfig:          req.LBConfig.ToEntity(),
		BasicAuth:         req.BasicAuth.ToEntity(),
		ClientConfig:      req.ClientConfig.ToEntity(),
		HeaderConfig:      req.HeaderConfig.ToEntity(),
		CompressionConfig: req.CompressionConfig.ToEntity(),
		RateLimitConfig:   req.RateLimitConfig.ToEntity(),
		Paths: gofn.MapSlice(req.Paths, func(item *HTTPPathConfigReq) *entity.HTTPPathConfig {
			return item.ToEntity()
		}),
	}
}

// nolint
func (req *DomainReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDomain(&req.Domain, true, base.DomainNameMaxLen,
		false, field+"domain")...)
	res = append(res, basedto.ValidateDomain(&req.DomainRedirect, false, base.DomainNameMaxLen,
		false, field+"domainRedirect")...)
	res = append(res, basedto.ValidatePort(&req.ContainerPort, true, 1, field+"containerPort")...)
	// TODO: add validation for others
	return res
}

type HTTPLBConfigReq struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

func (r *HTTPLBConfigReq) ToEntity() *entity.HTTPLBConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPLBConfig{
		Strategy: r.Strategy,
	}
}

type HTTPBasicAuthConfigReq struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

func (r *HTTPBasicAuthConfigReq) ToEntity() *entity.HTTPBasicAuthConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPBasicAuthConfig{
		Enabled: r.Enabled,
		ID:      r.ID,
	}
}

type HTTPClientConfigReq struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

func (r *HTTPClientConfigReq) ToEntity() *entity.HTTPClientConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPClientConfig{
		Enabled:        r.Enabled,
		MaxRequestBody: r.MaxRequestBody,
		MemRequestBody: r.MemRequestBody,
		AllowedIPs:     r.AllowedIPs,
	}
}

type HTTPHeaderConfigReq struct {
	Enabled               bool              `json:"enabled"`
	ToAddToRequests       map[string]string `json:"toAddToRequests"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests"`
	ToAddToResponses      map[string]string `json:"toAddToResponses"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses"`
}

func (r *HTTPHeaderConfigReq) ToEntity() *entity.HTTPHeaderConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPHeaderConfig{
		Enabled:               r.Enabled,
		ToAddToRequests:       r.ToAddToRequests,
		ToRemoveFromRequests:  r.ToRemoveFromRequests,
		ToAddToResponses:      r.ToAddToResponses,
		ToRemoveFromResponses: r.ToRemoveFromResponses,
	}
}

type HTTPCompressionConfigReq struct {
	Enabled              bool          `json:"enabled"`
	ExcludedContentTypes []string      `json:"excludedContentTypes"`
	IncludedContentTypes []string      `json:"includedContentTypes"`
	MinResponseBody      unit.DataSize `json:"minResponseBody"`
	DefaultEncoding      string        `json:"defaultEncoding"`
}

func (r *HTTPCompressionConfigReq) ToEntity() *entity.HTTPCompressionConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPCompressionConfig{
		Enabled:              r.Enabled,
		ExcludedContentTypes: r.ExcludedContentTypes,
		IncludedContentTypes: r.IncludedContentTypes,
		MinResponseBody:      r.MinResponseBody,
		DefaultEncoding:      r.DefaultEncoding,
	}
}

type HTTPRateLimitConfigReq struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

func (r *HTTPRateLimitConfigReq) ToEntity() *entity.HTTPRateLimitConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPRateLimitConfig{
		Enabled:        r.Enabled,
		Average:        r.Average,
		Period:         r.Period,
		Burst:          r.Burst,
		MaxInFlightReq: r.MaxInFlightReq,
	}
}

type HTTPPathConfigReq struct {
	Enabled           bool                      `json:"enabled"`
	Path              string                    `json:"path"`
	Mode              base.HTTPPathMode         `json:"mode"`
	BasicAuth         *HTTPBasicAuthConfigReq   `json:"basicAuth"`
	ClientConfig      *HTTPClientConfigReq      `json:"clientConfig"`
	RateLimitConfig   *HTTPRateLimitConfigReq   `json:"rateLimitConfig"`
	CompressionConfig *HTTPCompressionConfigReq `json:"compressionConfig"`
}

func (r *HTTPPathConfigReq) ToEntity() *entity.HTTPPathConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPPathConfig{
		Enabled:           r.Enabled,
		Path:              r.Path,
		Mode:              r.Mode,
		BasicAuth:         r.BasicAuth.ToEntity(),
		ClientConfig:      r.ClientConfig.ToEntity(),
		RateLimitConfig:   r.RateLimitConfig.ToEntity(),
		CompressionConfig: r.CompressionConfig.ToEntity(),
	}
}

func NewUpdateAppHttpSettingsReq() *UpdateAppHttpSettingsReq {
	return &UpdateAppHttpSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppHttpSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, vld.Slice(req.Domains).ForEach(
		func(r *DomainReq, index int, elemValidator vld.ItemValidator) {
			elemValidator.Validate(r.validate(fmt.Sprintf("domains[%d]", index))...)
		}))
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppHttpSettingsResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *UpdateAppHttpSettingsDataResp `json:"data"`
}

type UpdateAppHttpSettingsDataResp struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
