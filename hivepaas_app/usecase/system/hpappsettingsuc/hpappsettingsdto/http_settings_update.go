package hpappsettingsdto

import (
	"fmt"
	"strings"
	"unicode"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type UpdateHttpSettingsReq struct {
	Domains   []*DomainReq `json:"domains"`
	UpdateVer int          `json:"updateVer"`
}

func (req *UpdateHttpSettingsReq) ApplyTo(setting *entity.AppHttpSettings) error {
	currDomains := setting.Domains
	setting.Domains = []*entity.AppDomain{}
	for _, domain := range req.Domains {
		targetDomain, _ := gofn.Find(currDomains, func(d *entity.AppDomain) bool {
			return strings.EqualFold(d.Domain, domain.Domain)
		})
		if targetDomain == nil {
			targetDomain = &entity.AppDomain{
				Enabled:       domain.Enabled,
				Domain:        domain.Domain,
				ContainerPort: config.Current.HTTPServer.Port,
				ForceHttps:    true,
			}
		}
		if err := domain.ApplyTo(targetDomain); err != nil {
			return apperrors.Wrap(err)
		}
		setting.Domains = append(setting.Domains, targetDomain)
	}
	return nil
}

type DomainReq struct {
	Enabled         bool                    `json:"enabled"`
	Domain          string                  `json:"domain"`
	SSLCert         basedto.ObjectIDReq     `json:"sslCert"`
	ClientConfig    *HTTPClientConfigReq    `json:"clientConfig"`
	RateLimitConfig *HTTPRateLimitConfigReq `json:"rateLimitConfig"`
}

func (req *DomainReq) ApplyTo(targetDomain *entity.AppDomain) error {
	targetDomain.Enabled = req.Enabled
	targetDomain.Domain = req.Domain
	targetDomain.SSLCert = *req.SSLCert.ToEntity()

	if req.ClientConfig != nil {
		if targetDomain.ClientConfig == nil {
			targetDomain.ClientConfig = &entity.HTTPClientConfig{}
		}
		if err := req.ClientConfig.ApplyTo(targetDomain.ClientConfig); err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		targetDomain.ClientConfig = nil
	}

	if req.RateLimitConfig != nil {
		if targetDomain.RateLimitConfig == nil {
			targetDomain.RateLimitConfig = &entity.HTTPRateLimitConfig{}
		}
		if err := req.RateLimitConfig.ApplyTo(targetDomain.RateLimitConfig); err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		targetDomain.RateLimitConfig = nil
	}

	return nil
}

//nolint:unparam
func (req *DomainReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.Domain = strings.ToLower(strings.TrimSpace(req.Domain))
	if err := req.ClientConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.RateLimitConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

//nolint:unparam
func (req *DomainReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDomain(&req.Domain, true, base.DomainNameMaxLen,
		false, field+"domain")...)
	res = append(res, req.ClientConfig.validate(field+"clientConfig")...)
	res = append(res, req.RateLimitConfig.validate(field+"rateLimitConfig")...)
	return res
}

type HTTPClientConfigReq struct {
	Enabled    bool     `json:"enabled"`
	AllowedIPs []string `json:"allowedIPs"`
}

func (req *HTTPClientConfigReq) ApplyTo(clientConfig *entity.HTTPClientConfig) error {
	clientConfig.Enabled = req.Enabled
	clientConfig.AllowedIPs = req.AllowedIPs
	return nil
}

//nolint:unparam
func (req *HTTPClientConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.AllowedIPs = strings.FieldsFunc(strings.Join(req.AllowedIPs, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return nil
}

//nolint:unparam
func (req *HTTPClientConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	return res
}

type HTTPRateLimitConfigReq struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

func (req *HTTPRateLimitConfigReq) ApplyTo(rateLimitConfig *entity.HTTPRateLimitConfig) error {
	rateLimitConfig.Enabled = req.Enabled
	rateLimitConfig.Average = req.Average
	rateLimitConfig.Period = req.Period
	rateLimitConfig.Burst = req.Burst
	rateLimitConfig.MaxInFlightReq = req.MaxInFlightReq
	return nil
}

//nolint:unparam
func (req *HTTPRateLimitConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (req *HTTPRateLimitConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	return res
}

func NewUpdateHttpSettingsReq() *UpdateHttpSettingsReq {
	return &UpdateHttpSettingsReq{}
}

func (req *UpdateHttpSettingsReq) ModifyRequest() error {
	for _, domainReq := range req.Domains {
		if err := domainReq.modifyRequest(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateHttpSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, vld.Slice(req.Domains).ForEach(
		func(r *DomainReq, index int, elemValidator vld.ItemValidator) {
			elemValidator.Validate(r.validate(fmt.Sprintf("domains[%d]", index))...)
		}))
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateHttpSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
