package registryserviceimpl

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Registry traffic through Cloudflare's proxy leaves the cluster, comes back, and
// meets an upload limit on the way. These headers are what says so from outside.
func TestProxyEvidenceFromCloudflare(t *testing.T) {
	header := http.Header{}
	header.Set("cf-ray", "8d2f00000000-SIN")
	header.Set("server", "cloudflare")

	evidence := proxyEvidence(header)

	assert.Len(t, evidence, 2)
	assert.Contains(t, evidence[0], "cf-ray")
	assert.Contains(t, evidence[1], "cloudflare")
}

func TestProxyEvidenceFromAGenericProxy(t *testing.T) {
	header := http.Header{}
	header.Set("via", "1.1 varnish")

	assert.Len(t, proxyEvidence(header), 1)
}

// A registry answering for itself is the state an operator is aiming for, and it
// must not be reported as a warning.
func TestProxyEvidenceFromZotItself(t *testing.T) {
	header := http.Header{}
	header.Set("content-type", "application/json")
	header.Set("docker-distribution-api-version", "registry/2.0")

	assert.Empty(t, proxyEvidence(header))
}
