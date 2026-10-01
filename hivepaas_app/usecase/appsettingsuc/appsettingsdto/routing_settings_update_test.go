package appsettingsdto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestDomainExtraALPNProtocolsKeptForATCPDomainEndingTLSOnly(t *testing.T) {
	extra := []string{"x-amzn-mqtt-ca"}

	tcp := &DomainReq{Domain: "mqtt.example.com", Protocol: base.NetworkProtocolTCP, ExtraALPNProtocols: extra}
	assert.Equal(t, extra, tcp.ToEntity().ExtraALPNProtocols)

	passthrough := &DomainReq{Domain: "mqtt.example.com", Protocol: base.NetworkProtocolTCP,
		TLSPassthrough: true, ExtraALPNProtocols: extra}
	assert.Empty(t, passthrough.ToEntity().ExtraALPNProtocols, "the app answers the handshake itself")

	http := &DomainReq{Domain: "app.example.com", Protocol: base.NetworkProtocolHTTP, ExtraALPNProtocols: extra}
	assert.Empty(t, http.ToEntity().ExtraALPNProtocols)
}

func TestDomainExtraALPNProtocolsTrimmed(t *testing.T) {
	req := &DomainReq{Domain: "mqtt.example.com", Protocol: base.NetworkProtocolTCP,
		ExtraALPNProtocols: []string{" mqtt ", "", "  "}}
	assert.NoError(t, req.modifyRequest())
	assert.Equal(t, []string{"mqtt"}, req.ExtraALPNProtocols)
}

func TestDomainExtraALPNProtocolsValidated(t *testing.T) {
	for name, tc := range map[string]struct {
		protocols []string
		valid     bool
	}{
		"registered ones":  {[]string{"tds/8.0", "x-amzn-mqtt-ca"}, true},
		"none":             {nil, true},
		"with a space":     {[]string{"my proto"}, false},
		"repeated":         {[]string{"mqtt", "mqtt"}, false},
		"over 255 bytes":   {[]string{strings.Repeat("a", 256)}, false},
		"more than ten":    {[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}, false},
		"non-ASCII letter": {[]string{"prötocol"}, false},
	} {
		t.Run(name, func(t *testing.T) {
			req := &UpdateAppRoutingSettingsReq{
				ProjectID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", ProjectEnvID: "01HZZZZZZZZZZZZZZZZZZZZZZZ",
				AppID: "01HZZZZZZZZZZZZZZZZZZZZZZZ",
				Domains: []*DomainReq{{Enabled: true, Domain: "mqtt.example.com", Protocol: base.NetworkProtocolTCP,
					ExtraALPNProtocols: tc.protocols}},
			}
			errs := req.Validate()
			if tc.valid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}
