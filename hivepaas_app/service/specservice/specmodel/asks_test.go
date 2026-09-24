package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishedPortsLeavesOutAPortPublishedNowhere(t *testing.T) {
	doc := &AppDoc{Deployment: &Deployment{Networks: &Networks{EndpointSpec: &EndpointSpec{
		Ports: []*PortConfig{{Target: 80, Published: 8080}, {Target: 9000}, nil},
	}}}}

	assert.Equal(t, []PortConfig{{Target: 80, Published: 8080}}, PublishedPorts(doc))
	assert.Empty(t, PublishedPorts(&AppDoc{}))
}

// An exported routing block names a certificate the export did not hold by an
// external reference, where the certificate's id was. That is not an address,
// and it must not stop the addresses being read.
func TestActiveDomainsReadsARoutingBlockWithAnExternalCertificate(t *testing.T) {
	doc := &AppDoc{Settings: map[string]any{"routing": map[string]any{
		"port": 8080, "exposePublicly": true,
		"domains": []any{
			map[string]any{"enabled": true, "domain": "api.example.com", "sslCert": map[string]any{
				"id": map[string]any{ExternalRefKey: map[string]any{"type": "ssl-cert", "name": "wildcard"}},
			}},
			map[string]any{"enabled": false, "domain": "old.example.com"},
		},
		SettingMetaKey: map[string]any{"version": 1},
	}}}

	domains, err := ActiveDomains(doc)

	assert.NoError(t, err)
	assert.Equal(t, []string{"api.example.com"}, domains)
}

func TestActiveDomainsOfADocumentWithoutRouting(t *testing.T) {
	domains, err := ActiveDomains(&AppDoc{})

	assert.NoError(t, err)
	assert.Empty(t, domains)
}

func TestGrantedCapabilities(t *testing.T) {
	withCapabilities := func(c *Capabilities) *AppDoc {
		return &AppDoc{Deployment: &Deployment{Resources: &Resources{Capabilities: c}}}
	}

	assert.Equal(t, []string{"NET_ADMIN", "sysctls"}, GrantedCapabilities(withCapabilities(&Capabilities{
		CapabilityAdd: []string{"NET_ADMIN"}, Sysctls: map[string]string{"net.core.somaxconn": "1024"},
	})))
	assert.Equal(t, []string{"deployment.resources.capabilities"},
		GrantedCapabilities(withCapabilities(&Capabilities{CapabilityDrop: []string{"ALL"}})),
		"dropping is still the privileged block")
	assert.Empty(t, GrantedCapabilities(withCapabilities(&Capabilities{})))
	assert.Empty(t, GrantedCapabilities(&AppDoc{}))
}
