package hpappsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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

	// PendingChange is set while a change is on trial and has still to be
	// confirmed. It is repeated here, rather than only on the update that started
	// it, so a dashboard reloaded mid-trial can pick the countdown back up.
	PendingChange *PendingChangeResp `json:"pendingChange,omitempty"`
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
	App            *entity.App
	RoutingSetting *entity.Setting
	RefObjects     *entity.RefObjects
}

func TransformRoutingSettings(input *RoutingSettingsTransformInput) (resp *RoutingSettingsResp, err error) {
	resp = &RoutingSettingsResp{}
	if input.RoutingSetting == nil {
		return resp, nil
	}

	if input.RefObjects == nil {
		input.RefObjects = entity.NewRefObjects()
	}

	if err = copier.Copy(&resp, input.RoutingSetting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	routingSettings := input.RoutingSetting.MustAsAppRoutingSettings()
	if err = copier.Copy(&resp, routingSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefObjects.RefSettings[domain.SSLCert.ID]
			certResp, _ := sslcertdto.TransformSSLCertBasic(setting, input.RefObjects)
			if certResp == nil {
				certResp = &sslcertdto.SSLCertResp{
					BaseSettingResp: settings.NewMissingSetting(domain.SSLCert.ID, base.SettingTypeSSLCert),
				}
			}
			domain.SSLCert = certResp
		} else {
			domain.SSLCert = nil
		}
	}

	return resp, nil
}
