package traefikserviceimpl

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// replaced is what Traefik's ReplacePathRegex makes of a path, with the labels
// a rewrite was given.
func replaced(t *testing.T, labels map[string]string, path string) string {
	t.Helper()
	regex := labels["traefik.http.middlewares.app-web-replacepathregex.replacepathregex.regex"]
	replacement := labels["traefik.http.middlewares.app-web-replacepathregex.replacepathregex.replacement"]
	return regexp.MustCompile(regex).ReplaceAllString(path, replacement)
}

// A path written out is replaced, and what is under it; another path that
// starts the same is not, nor any other.
func TestAPlainPathIsReplacedWithWhatIsUnderIt(t *testing.T) {
	labels := map[string]string{}
	var middlewares []string

	(&service{}).createPathRewriteConfig(&entity.HTTPPathRewriteConfig{
		Enabled:         true,
		PathReplace:     "/old",
		PathReplaceWith: "/new",
	}, "app-web", labels, &middlewares)

	assert.Equal(t, []string{"app-web-replacepathregex@swarm"}, middlewares)
	assert.Equal(t, "/new", replaced(t, labels, "/old"))
	assert.Equal(t, "/new/page", replaced(t, labels, "/old/page"))
	assert.Equal(t, "/older", replaced(t, labels, "/older"))
	assert.Equal(t, "/other", replaced(t, labels, "/other"))
}

// A path written out is taken as written: its dot is a dot, not any character.
// A trailing slash on either side changes nothing, and a $ in the replacement
// is a $.
func TestAPlainPathIsTakenAsWritten(t *testing.T) {
	labels := map[string]string{}
	var middlewares []string

	(&service{}).createPathRewriteConfig(&entity.HTTPPathRewriteConfig{
		Enabled:         true,
		PathReplace:     "/v1.0/",
		PathReplaceWith: "/v$2/",
	}, "app-web", labels, &middlewares)

	assert.Equal(t, "/v$2/items", replaced(t, labels, "/v1.0/items"))
	assert.Equal(t, "/v1x0/items", replaced(t, labels, "/v1x0/items"))
}

// A pattern is the user's own, with its groups.
func TestAPatternReplacesThePathsItMatches(t *testing.T) {
	labels := map[string]string{}
	var middlewares []string

	(&service{}).createPathRewriteConfig(&entity.HTTPPathRewriteConfig{
		Enabled:            true,
		PathReplace:        "^/blog/([0-9]+)$",
		PathReplaceIsRegex: true,
		PathReplaceWith:    "/posts/$1/view",
	}, "app-web", labels, &middlewares)

	assert.Equal(t, "/posts/42/view", replaced(t, labels, "/blog/42"))
	assert.Equal(t, "/blog/abc", replaced(t, labels, "/blog/abc"))
}

// Nothing to replace a path with replaces nothing.
func TestAReplacementLeftEmptyAddsNoMiddleware(t *testing.T) {
	for _, isRegex := range []bool{false, true} {
		labels := map[string]string{}
		var middlewares []string

		(&service{}).createPathRewriteConfig(&entity.HTTPPathRewriteConfig{
			Enabled:            true,
			PathReplace:        "/old",
			PathReplaceIsRegex: isRegex,
		}, "app-web", labels, &middlewares)

		assert.Empty(t, middlewares)
		for key := range labels {
			assert.False(t, strings.Contains(key, "replacepath"), key)
		}
	}
}
