package traefikserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// Compression prefers the modern encodings a client accepts, in this order, and
// names no default encoding: with one, a client that sent no Accept-Encoding
// was answered in Brotli, which it never said it could read.
func TestCompressionPrefersModernEncodingsAndAssumesNone(t *testing.T) {
	labels := map[string]string{}
	var middlewares []string

	(&service{}).createCompressionConfig(&entity.HTTPCompressionConfig{
		Enabled:         true,
		MinResponseBody: unit.KB,
	}, "app-web", labels, &middlewares)

	assert.Equal(t, "zstd,br,gzip", labels["traefik.http.middlewares.app-web-compress.compress.encodings"])
	assert.Equal(t, "1024", labels["traefik.http.middlewares.app-web-compress.compress.minresponsebodybytes"])
	assert.NotContains(t, labels, "traefik.http.middlewares.app-web-compress.compress.defaultencoding")
	assert.Len(t, middlewares, 1)
}

// Compression that is off adds nothing.
func TestCompressionOffAddsNoMiddleware(t *testing.T) {
	labels := map[string]string{}
	var middlewares []string

	(&service{}).createCompressionConfig(&entity.HTTPCompressionConfig{Enabled: false}, "app-web", labels, &middlewares)

	assert.Empty(t, labels)
	assert.Empty(t, middlewares)
}
