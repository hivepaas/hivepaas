package hpappsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

type GetRoutingSettingsReq struct {
}

func NewGetRoutingSettingsReq() *GetRoutingSettingsReq {
	return &GetRoutingSettingsReq{}
}

func (req *GetRoutingSettingsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetRoutingSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *RoutingSettingsResp `json:"data"`
}

type RoutingSettingsResp struct {
	Domains   []*DomainResp `json:"domains"`
	UpdateVer int           `json:"updateVer"`
}

type DomainResp struct {
	Enabled         bool                     `json:"enabled"`
	Domain          string                   `json:"domain"`
	SSLCert         *sslcertdto.SSLCertResp  `json:"sslCert,omitempty"`
	ClientConfig    *HTTPClientConfigResp    `json:"clientConfig,omitempty"`
	RateLimitConfig *HTTPRateLimitConfigResp `json:"rateLimitConfig,omitempty"`
}

type HTTPClientConfigResp struct {
	Enabled    bool     `json:"enabled"`
	AllowedIPs []string `json:"allowedIPs"`
}

type HTTPRateLimitConfigResp struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

type RoutingSettingsTransformInput struct {
	App             *entity.App
	RoutingSettings *entity.Setting
	RefSettingMap   map[string]*entity.Setting
}

func TransformRoutingSettings(input *RoutingSettingsTransformInput) (resp *RoutingSettingsResp, err error) {
	resp = &RoutingSettingsResp{}
	if input.RoutingSettings == nil {
		return resp, nil
	}

	if err = copier.Copy(&resp, input.RoutingSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}
	routingSettings := input.RoutingSettings.MustAsAppRoutingSettings()
	if err = copier.Copy(&resp, routingSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefSettingMap[domain.SSLCert.ID]
			domain.SSLCert, _ = sslcertdto.TransformSSLCertBasic(setting, &entity.RefObjects{})
		} else {
			domain.SSLCert = nil
		}
	}

	return resp, nil
}
