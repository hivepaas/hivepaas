package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A domain answers on the app's port unless it names one of its own. The one it
// does not name is stored as a zero, which is not a port: a caller that took it
// for one would ask for port 0 to be reachable, and a resource link would claim
// the app listens there.
func TestActivePortsLeaveOutTheOneADomainDoesNotName(t *testing.T) {
	settings := &AppRoutingSettings{
		Port:           8080,
		ExposePublicly: true,
		Domains: []*AppDomain{
			{Domain: "a.example.com", Enabled: true},
			{Domain: "b.example.com", Enabled: true, ContainerPort: 9000},
			{Domain: "c.example.com", Enabled: true, ContainerPort: 8080},
			{Domain: "d.example.com", Enabled: false, ContainerPort: 7000},
		},
	}

	assert.Equal(t, []int{8080, 9000}, settings.GetActivePorts())
}

func TestActivePortsOfAnAppWithNoDomains(t *testing.T) {
	settings := &AppRoutingSettings{Port: 5432}

	assert.Equal(t, []int{5432}, settings.GetActivePorts())
}

// An app that has to be told its own address gets one only when it has a
// domain: what an app with none needs to be told is nothing at all.
func TestAppURLIsTheFirstAddressWithTheSchemeItIsServedOver(t *testing.T) {
	cases := map[string]struct {
		settings *AppRoutingSettings
		want     string
	}{
		"no routing at all": {nil, ""},
		"not exposed": {&AppRoutingSettings{
			Domains: []*AppDomain{{Domain: "app.example.com", Enabled: true, ForceHttps: true}},
		}, ""},
		"exposed with no domain": {&AppRoutingSettings{ExposePublicly: true}, ""},
		"a domain that is not enabled": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{{Domain: "app.example.com"}},
		}, ""},
		"plain http": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{{Domain: "app.example.com", Enabled: true}},
		}, "http://app.example.com"},
		"forced https": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{{Domain: "app.example.com", Enabled: true, ForceHttps: true}},
		}, "https://app.example.com"},
		"a certificate of its own": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{{Domain: "app.example.com", Enabled: true, SSLCert: ObjectID{ID: "cert-1"}}},
		}, "https://app.example.com"},
		"tls passed through": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{{Domain: "app.example.com", Enabled: true, TLSPassthrough: true}},
		}, "https://app.example.com"},
		"the first of several": {&AppRoutingSettings{ExposePublicly: true,
			Domains: []*AppDomain{
				{Domain: "first.example.com", Enabled: true, ForceHttps: true},
				{Domain: "second.example.com", Enabled: true},
			},
		}, "https://first.example.com"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.settings.GetAppURL())
		})
	}
}
