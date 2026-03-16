package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentAppHttpSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppHttp, &appHttpSettingsParser{})

type appHttpSettingsParser struct {
}

func (s *appHttpSettingsParser) New() SettingData {
	return &AppHttpSettings{}
}

type AppHttpSettings struct {
	Enabled bool         `json:"enabled"`
	Domains []*AppDomain `json:"domains,omitempty"`
	Reset   bool         `json:"reset,omitempty"`
}

type AppDomain struct {
	Enabled         bool           `json:"enabled"`
	Domain          string         `json:"domain"`
	DomainRedirect  string         `json:"domainRedirect,omitempty"`
	SSLCert         ObjectID       `json:"sslCert,omitzero"`
	ContainerPort   int            `json:"containerPort,omitempty"`
	ForceHttps      bool           `json:"forceHttps,omitempty"`
	WebsocketConfig string         `json:"websocketConfig,omitempty"`
	BasicAuth       ObjectID       `json:"basicAuth,omitzero"`
	NginxSettings   *NginxSettings `json:"nginxSettings,omitempty"`
}

type NginxSettings struct {
	ClientConfig    string                `json:"clientConfig,omitempty"`
	GzipConfig      string                `json:"gzipConfig,omitempty"` // on/off/default/custom
	LimitZoneConfig string                `json:"limitZoneConfig,omitempty"`
	CustomConfig    string                `json:"customConfig,omitempty"`
	Locations       []*NginxLocationBlock `json:"locations"`
}

type NginxLocationBlock struct {
	Location          string   `json:"location"`
	ProxyHeaderConfig string   `json:"proxyHeaderConfig,omitempty"`
	WebsocketConfig   string   `json:"websocketConfig,omitempty"`
	BasicAuth         ObjectID `json:"basicAuth,omitzero"`
	LimitReqConfig    string   `json:"limitReqConfig,omitempty"`
	CustomConfig      string   `json:"customConfig,omitempty"`
}

func (s *AppHttpSettings) GetDomain(domain string) *AppDomain {
	domain = strings.ToLower(domain)
	for _, domainRec := range s.Domains {
		if domainRec.Domain == domain {
			return domainRec
		}
	}
	return nil
}

func (s *AppHttpSettings) GetType() base.SettingType {
	return base.SettingTypeAppHttp
}

func (s *AppHttpSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.GetSSLCertIDs(), s.GetBasicAuthIDs()),
	}
}

func (s *AppHttpSettings) GetSSLCertIDs() (res []string) {
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		if domain.SSLCert.ID != "" {
			res = append(res, domain.SSLCert.ID)
		}
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppHttpSettings) GetBasicAuthIDs() (res []string) {
	for _, domain := range s.Domains {
		if !domain.Enabled {
			continue
		}
		if domain.BasicAuth.ID != "" {
			res = append(res, domain.BasicAuth.ID)
		}
		if domain.NginxSettings == nil {
			continue
		}
		for _, locationBlock := range domain.NginxSettings.Locations {
			if locationBlock.BasicAuth.ID != "" {
				res = append(res, locationBlock.BasicAuth.ID)
			}
		}
	}
	res = gofn.ToSet(res)
	return
}

func (s *AppHttpSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentAppHttpSettingsVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAppHttpSettingsVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAppHttpSettings() (*AppHttpSettings, error) {
	return parseSettingAs[*AppHttpSettings](s)
}

func (s *Setting) MustAsAppHttpSettings() *AppHttpSettings {
	return gofn.Must(s.AsAppHttpSettings())
}
