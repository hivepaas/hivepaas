package hpappsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// validateProxy runs the validators, since validate() only builds them.
func validateProxy(req *HivePaaSProxySettingsReq) hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(req.validate("proxySettings")...))
}

func proxyReq(provider string, trustedIPs []string, hops int) *HivePaaSProxySettingsReq {
	return &HivePaaSProxySettingsReq{
		ProxyProvider: provider,
		TrustedIPs:    trustedIPs,
		ProxyHops:     hops,
	}
}

func TestProxySettingsValidate(t *testing.T) {
	t.Run("a complete declaration", func(t *testing.T) {
		req := proxyReq("cloudflare", []string{"173.245.48.0/20"}, 2)
		assert.NoError(t, req.modifyRequest())
		assert.Empty(t, validateProxy(req))
	})

	// The shape that hurts. Traefik is never told to trust anything, so forwarded
	// headers are ignored, every caller behind the proxy counts as one, and the
	// rate limits throttle the install instead of protecting it - with the
	// settings page showing a configured proxy the whole time.
	t.Run("a proxy without trusted IPs is refused", func(t *testing.T) {
		req := proxyReq("cloudflare", nil, 2)
		assert.NoError(t, req.modifyRequest())
		assert.NotEmpty(t, validateProxy(req))
	})

	t.Run("a proxy without hops is refused", func(t *testing.T) {
		req := proxyReq("cloudflare", []string{"173.245.48.0/20"}, 0)
		assert.NoError(t, req.modifyRequest())
		assert.NotEmpty(t, validateProxy(req))
	})

	t.Run("an unparseable trusted IP is refused", func(t *testing.T) {
		req := proxyReq("cloudflare", []string{"not-an-address"}, 2)
		assert.NoError(t, req.modifyRequest())
		assert.NotEmpty(t, validateProxy(req))
	})

	t.Run("hops beyond the bound are refused", func(t *testing.T) {
		req := proxyReq("cloudflare", []string{"10.0.0.0/8"}, proxyHopsMax+1)
		assert.NoError(t, req.modifyRequest())
		assert.NotEmpty(t, validateProxy(req))
	})

	// A form still holding yesterday's values must not stop the operator from
	// saying there is no proxy any more.
	t.Run("clearing the provider drops the rest", func(t *testing.T) {
		req := proxyReq("", []string{"10.0.0.0/8"}, 2)
		assert.NoError(t, req.modifyRequest())

		assert.Empty(t, req.TrustedIPs)
		assert.Zero(t, req.ProxyHops)
		assert.Empty(t, validateProxy(req))
	})
}

func TestTransformRequestInfo(t *testing.T) {
	t.Run("a chain through one proxy", func(t *testing.T) {
		resp := TransformRequestInfo(&RequestInfoTransformInput{
			RemoteAddr: "10.0.0.9:41234",
			ClientIP:   "203.0.113.7",
			HeaderValues: map[string]string{
				"X-Forwarded-For": "203.0.113.7, 172.68.1.1",
			},
		})

		assert.Equal(t, []string{"203.0.113.7", "172.68.1.1"}, resp.ForwardedFor)
		assert.Equal(t, 1, resp.SuggestedProxyHops)
	})

	t.Run("a chain through two proxies", func(t *testing.T) {
		resp := TransformRequestInfo(&RequestInfoTransformInput{
			HeaderValues: map[string]string{
				"X-Forwarded-For": "203.0.113.7, 172.68.1.1, 10.0.0.9",
			},
		})
		assert.Equal(t, 2, resp.SuggestedProxyHops)
	})

	// One entry is the one Traefik adds itself, so nothing is in front.
	t.Run("no proxy in front", func(t *testing.T) {
		resp := TransformRequestInfo(&RequestInfoTransformInput{
			HeaderValues: map[string]string{"X-Forwarded-For": "203.0.113.7"},
		})
		assert.Zero(t, resp.SuggestedProxyHops)
		assert.Contains(t, resp.Explanation, "Leave proxyHops unset")
	})

	t.Run("no header at all", func(t *testing.T) {
		resp := TransformRequestInfo(&RequestInfoTransformInput{})
		assert.Empty(t, resp.ForwardedFor)
		assert.Zero(t, resp.SuggestedProxyHops)
	})

	// Reflecting whatever arrived would reflect the caller's own credentials.
	t.Run("only proxy headers are echoed", func(t *testing.T) {
		headers := CollectProxyHeaders(func(name string) string {
			switch name {
			case "X-Forwarded-For":
				return "203.0.113.7"
			case "Authorization":
				return "Bearer super-secret"
			default:
				return ""
			}
		})

		assert.Equal(t, map[string]string{"X-Forwarded-For": "203.0.113.7"}, headers)
	})
}
