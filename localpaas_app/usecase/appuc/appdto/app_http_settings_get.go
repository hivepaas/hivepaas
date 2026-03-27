package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	Enabled   bool          `json:"enabled"`
	Domains   []*DomainResp `json:"domains"`
	UpdateVer int           `json:"updateVer"`
}

type DomainResp struct {
	Enabled         bool                      `json:"enabled"`
	Domain          string                    `json:"domain"`
	DomainRedirect  string                    `json:"domainRedirect"`
	SSLCert         *settings.BaseSettingResp `json:"sslCert"`
	ContainerPort   int                       `json:"containerPort"`
	ForceHttps      bool                      `json:"forceHttps"`
	WebsocketConfig string                    `json:"websocketConfig"`
	BasicAuth       *settings.BaseSettingResp `json:"basicAuth"`
	NginxSettings   *NginxSettingsResp        `json:"nginxSettings"`
}

type NginxSettingsResp struct {
	ClientConfig    string                    `json:"clientConfig"`
	GzipConfig      string                    `json:"gzipConfig"` // on/off/default/custom
	LimitZoneConfig string                    `json:"limitZoneConfig"`
	CustomConfig    string                    `json:"customConfig"`
	Locations       []*NginxLocationBlockResp `json:"locations"`
}

type NginxLocationBlockResp struct {
	Location          string                    `json:"location"`
	ProxyHeaderConfig string                    `json:"proxyHeaderConfig"`
	WebsocketConfig   string                    `json:"websocketConfig"`
	BasicAuth         *settings.BaseSettingResp `json:"basicAuth,omitzero"`
	LimitReqConfig    string                    `json:"limitReqConfig"`
	CustomConfig      string                    `json:"customConfig"`
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

	for _, domain := range resp.Domains {
		if domain.SSLCert != nil && domain.SSLCert.ID != "" {
			setting := input.RefSettingMap[domain.SSLCert.ID]
			domain.SSLCert, _ = settings.TransformSettingBase(setting)
		} else {
			domain.SSLCert = nil
		}
		if domain.BasicAuth != nil && domain.BasicAuth.ID != "" {
			setting := input.RefSettingMap[domain.BasicAuth.ID]
			domain.BasicAuth, _ = settings.TransformSettingBase(setting)
		} else {
			domain.BasicAuth = nil
		}

		if domain.NginxSettings == nil {
			continue
		}
		for _, locationBlock := range domain.NginxSettings.Locations {
			setting := input.RefSettingMap[locationBlock.BasicAuth.ID]
			if locationBlock.BasicAuth != nil && locationBlock.BasicAuth.ID != "" {
				locationBlock.BasicAuth, _ = settings.TransformSettingBase(setting)
			} else {
				locationBlock.BasicAuth = nil
			}
		}
	}

	return resp, nil
}
