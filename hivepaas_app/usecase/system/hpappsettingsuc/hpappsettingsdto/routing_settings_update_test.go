package hpappsettingsdto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

func routingReq(domains ...*DomainReq) *UpdateRoutingSettingsReq {
	return &UpdateRoutingSettingsReq{Domains: domains}
}

func domainReq(allowedIPs ...string) *DomainReq {
	req := &DomainReq{Enabled: true, Domain: "app.example.com", SSLCert: basedto.ObjectIDReq{}}
	if len(allowedIPs) > 0 {
		req.ClientConfig = &HTTPClientConfigReq{Enabled: true, AllowedIPs: allowedIPs}
	}
	return req
}

// Every router label for the app is rebuilt from this list, so an empty one
// leaves the service with no route - proven by generating the labels.
func TestUpdateRoutingSettingsRequiresADomain(t *testing.T) {
	req := routingReq()
	assert.NoError(t, req.ModifyRequest())
	assert.NotEmpty(t, req.Validate())
}

func TestUpdateRoutingSettingsAcceptsADomain(t *testing.T) {
	req := routingReq(domainReq())
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
}

// An entry Traefik cannot parse takes down the router that carries it, which on
// this app is the dashboard.
func TestUpdateRoutingSettingsRejectsAnUnparseableAllowedIP(t *testing.T) {
	req := routingReq(domainReq("10.0.0.0/8", "not-an-address"))
	assert.NoError(t, req.ModifyRequest())
	assert.NotEmpty(t, req.Validate())
}

func TestUpdateRoutingSettingsAcceptsValidAllowedIPs(t *testing.T) {
	req := routingReq(domainReq("10.0.0.0/8", "203.0.113.7", "2001:db8::/32"))
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
}

// A disabled allowlist carries no addresses to check.
func TestUpdateRoutingSettingsIgnoresADisabledAllowlist(t *testing.T) {
	req := routingReq(domainReq())
	req.Domains[0].ClientConfig = &HTTPClientConfigReq{Enabled: false, AllowedIPs: []string{"nonsense"}}
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
}

// The dashboard sends this field as a duration string, and it is the only field
// in the payload that is not a plain JSON type. Getting the encoding wrong is
// silent in the worst possible way: the value parses as zero, the server falls
// back to its default, and nobody finds out until someone needed the longer
// window and did not get it.
func TestConfirmWindowDecodesFromDurationString(t *testing.T) {
	tests := []struct {
		name string
		body string
		want time.Duration
	}{
		{name: "minutes", body: `{"confirmWindow":"5m"}`, want: 5 * time.Minute},
		{name: "seconds", body: `{"confirmWindow":"90s"}`, want: 90 * time.Second},
		{name: "absent", body: `{}`, want: 0},
		{name: "null", body: `{"confirmWindow":null}`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &UpdateRoutingSettingsReq{}
			assert.NoError(t, json.Unmarshal([]byte(tt.body), req))
			assert.Equal(t, tt.want, req.ConfirmWindow.ToDuration())
		})
	}
}
