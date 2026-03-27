package appdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type UpdateAppHttpSettingsReq struct {
	ProjectID string       `json:"-"`
	AppID     string       `json:"-"`
	Enabled   bool         `json:"enabled"`
	Domains   []*DomainReq `json:"domains"`
	UpdateVer int          `json:"updateVer"`
}

func (req *UpdateAppHttpSettingsReq) ToEntity() *entity.AppHttpSettings {
	return &entity.AppHttpSettings{
		Enabled: req.Enabled,
		Domains: gofn.MapSlice(req.Domains, func(r *DomainReq) *entity.AppDomain {
			return r.ToEntity()
		}),
	}
}

type DomainReq struct {
	Enabled         bool                `json:"enabled"`
	Domain          string              `json:"domain"`
	DomainRedirect  string              `json:"domainRedirect"`
	SSLCert         basedto.ObjectIDReq `json:"sslCert"`
	ContainerPort   int                 `json:"containerPort"`
	ForceHttps      bool                `json:"forceHttps"`
	WebsocketConfig string              `json:"websocketConfig"`
	BasicAuth       basedto.ObjectIDReq `json:"basicAuth"`
	NginxSettings   *NginxSettingsReq   `json:"nginxSettings"`
}

func (req *DomainReq) ToEntity() *entity.AppDomain {
	return &entity.AppDomain{
		Enabled:         req.Enabled,
		Domain:          req.Domain,
		DomainRedirect:  req.DomainRedirect,
		SSLCert:         entity.ObjectID{ID: req.SSLCert.ID},
		ContainerPort:   req.ContainerPort,
		ForceHttps:      req.ForceHttps,
		WebsocketConfig: req.WebsocketConfig,
		BasicAuth:       entity.ObjectID{ID: req.BasicAuth.ID},
		NginxSettings:   req.NginxSettings.ToEntity(),
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
	// TODO:
	return res
}

type NginxSettingsReq struct {
	ClientConfig    string                   `json:"clientConfig"`
	GzipConfig      string                   `json:"gzipConfig"` // on/off/default/custom
	LimitZoneConfig string                   `json:"limitZoneConfig"`
	CustomConfig    string                   `json:"customConfig"`
	Locations       []*NginxLocationBlockReq `json:"locations"`
}

func (r *NginxSettingsReq) ToEntity() *entity.NginxSettings {
	if r == nil {
		return nil
	}
	return &entity.NginxSettings{
		ClientConfig:    r.ClientConfig,
		GzipConfig:      r.GzipConfig,
		LimitZoneConfig: r.LimitZoneConfig,
		CustomConfig:    r.CustomConfig,
		Locations: gofn.MapSlice(r.Locations, func(item *NginxLocationBlockReq) *entity.NginxLocationBlock {
			return item.ToEntity()
		}),
	}
}

type NginxLocationBlockReq struct {
	Location          string              `json:"location"`
	ProxyHeaderConfig string              `json:"proxyHeaderConfig"`
	WebsocketConfig   string              `json:"websocketConfig"`
	BasicAuth         basedto.ObjectIDReq `json:"basicAuth"`
	LimitReqConfig    string              `json:"limitReqConfig"`
	CustomConfig      string              `json:"customConfig"`
}

func (r *NginxLocationBlockReq) ToEntity() *entity.NginxLocationBlock {
	if r == nil {
		return nil
	}
	return &entity.NginxLocationBlock{
		Location:          r.Location,
		ProxyHeaderConfig: r.ProxyHeaderConfig,
		WebsocketConfig:   r.WebsocketConfig,
		BasicAuth:         entity.ObjectID{ID: r.BasicAuth.ID},
		LimitReqConfig:    r.LimitReqConfig,
		CustomConfig:      r.CustomConfig,
	}
}

func NewUpdateAppHttpSettingsReq() *UpdateAppHttpSettingsReq {
	return &UpdateAppHttpSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppHttpSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate http settings input
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
