package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// UsesClientIP decides which apps a proxy topology change has to be swept across.
// An app it wrongly answers "no" for keeps a stale ip-strategy depth in its
// labels, and goes on reading the wrong forwarded position until something
// unrelated redeploys it.
func TestAppRoutingSettingsUsesClientIP(t *testing.T) {
	t.Run("nothing asks who the caller is", func(t *testing.T) {
		assert.False(t, (&AppRoutingSettings{
			ExposePublicly: true,
			Domains:        []*AppDomain{{Domain: "x.example.com"}},
		}).UsesClientIP())
	})

	t.Run("a rate limit on a path asks", func(t *testing.T) {
		assert.True(t, (&AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*AppDomain{{Paths: []*HTTPPathConfig{{
				RateLimitConfig: &HTTPRateLimitConfig{Enabled: true, Average: 10},
			}}}},
		}).UsesClientIP())
	})

	t.Run("an allowlist asks", func(t *testing.T) {
		assert.True(t, (&AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*AppDomain{{
				ClientConfig: &HTTPClientConfig{Enabled: true, AllowedIPs: []string{"10.0.0.0/8"}},
			}},
		}).UsesClientIP())
	})

	t.Run("a disabled rate limit does not", func(t *testing.T) {
		assert.False(t, (&AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*AppDomain{{
				RateLimitConfig: &HTTPRateLimitConfig{Enabled: false, Average: 10},
			}},
		}).UsesClientIP())
	})

	t.Run("an app that is not exposed publicly has no labels to carry a depth", func(t *testing.T) {
		assert.False(t, (&AppRoutingSettings{
			ExposePublicly: false,
			Domains: []*AppDomain{{
				ClientConfig: &HTTPClientConfig{Enabled: true, AllowedIPs: []string{"10.0.0.0/8"}},
			}},
		}).UsesClientIP())
	})

	t.Run("an allowlist that is enabled but empty does not ask", func(t *testing.T) {
		assert.False(t, (&AppRoutingSettings{
			ExposePublicly: true,
			Domains: []*AppDomain{{
				ClientConfig: &HTTPClientConfig{Enabled: true},
			}},
		}).UsesClientIP())
	})

	assert.False(t, (*AppRoutingSettings)(nil).UsesClientIP())
}
