package hpappsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

type GetHttpSettingsReq struct {
}

func NewGetHttpSettingsReq() *GetHttpSettingsReq {
	return &GetHttpSettingsReq{}
}

func (req *GetHttpSettingsReq) Validate() apperrors.ValidationErrors {
	return nil
}

type GetHttpSettingsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *HttpSettingsResp `json:"data"`
}

type HttpSettingsResp struct {
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

type HttpSettingsTransformInput struct {
	App           *entity.App
	HttpSettings  *entity.Setting
	RefSettingMap map[string]*entity.Setting
}

func TransformHttpSettings(input *HttpSettingsTransformInput) (resp *HttpSettingsResp, err error) {
	resp = &HttpSettingsResp{}
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
	}

	return resp, nil
}
