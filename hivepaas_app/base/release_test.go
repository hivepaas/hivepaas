package base

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas"
)

// The release compiled in is release.json's: one file says what a release runs,
// and nothing has to be kept in step with it by hand.
func TestTheCompiledBetaIsReleaseJSONs(t *testing.T) {
	var file map[string]*ReleaseInfo
	assert.NoError(t, json.Unmarshal(hivepaas.ReleaseJSON, &file))
	want := file["beta"]
	if !assert.NotNil(t, want) {
		return
	}

	assert.Equal(t, want.AppVersion, BetaVersion.AppVersion)
	assert.Equal(t, want.ReleaseDate, BetaVersion.ReleaseDate)
	assert.Equal(t, want.RedisImage, BetaVersion.RedisImage)
	assert.Equal(t, want.DbImage, BetaVersion.DbImage)
	assert.Equal(t, want.TraefikImage, BetaVersion.TraefikImage)
	assert.Equal(t, want.VictoriaLogsImage, BetaVersion.VictoriaLogsImage)
	assert.Equal(t, want.VlagentImage, BetaVersion.VlagentImage)
	assert.Equal(t, want.OBIImage, BetaVersion.OBIImage)
	assert.NotEmpty(t, BetaVersion.OBIImage, "the agents run it")
	assert.Equal(t, want.RegistryImage, BetaVersion.RegistryImage)
	assert.Equal(t, want.PlaceholderImage, BetaVersion.PlaceholderImage)
	assert.NotEmpty(t, BetaVersion.PlaceholderImage, "a new app starts on it")
	assert.Equal(t, want.FunctionRuntimes, BetaVersion.FunctionRuntimes)
	assert.Equal(t, want.BlockMajorUpgrade, BetaVersion.BlockMajorUpgrade)
	assert.Nil(t, BetaVersion.Templates, "templates are only read from the release info fetched")
}

// The app's and the agent's images are named by the version: their digests
// exist only once the release is built, after the binary is.
func TestTheAppAndAgentImagesAreNamedByTheVersion(t *testing.T) {
	version := strings.TrimPrefix(BetaVersion.AppVersion, "v")

	assert.Equal(t, "ghcr.io/hivepaas/hivepaas:"+version, BetaVersion.AppImage)
	assert.Equal(t, "ghcr.io/hivepaas/hivepaas-agent:"+version, BetaVersion.AgentImage)
}

// release.json names the same app and agent images, pinned once they are built.
func TestReleaseJSONNamesTheAppAndAgentOfItsVersion(t *testing.T) {
	var file map[string]*ReleaseInfo
	assert.NoError(t, json.Unmarshal(hivepaas.ReleaseJSON, &file))
	for channel, release := range file {
		compiled := compiledRelease(hivepaas.ReleaseJSON, channel, nil)
		app, _, _ := strings.Cut(release.AppImage, "@")
		agent, _, _ := strings.Cut(release.AgentImage, "@")
		assert.Equal(t, compiled.AppImage, app, channel)
		assert.Equal(t, compiled.AgentImage, agent, channel)
	}
}

// Until release.json names a stable release, the stable channel runs the
// beta's images under the version it has always had.
func TestStableIsTheBetasImagesUntilOneIsReleased(t *testing.T) {
	data := []byte(`{"beta": {"appVersion": "v1.0.0-beta1", "redisImage": "redis:8", ` +
		`"functionRuntimes": {"node24": "n"}, "templates": {"repo": "r"}}}`)
	fallback := &ReleaseInfo{AppVersion: "v0.1.0"}

	stable := compiledRelease(data, "stable", fallback)

	assert.Equal(t, "v0.1.0", stable.AppVersion)
	assert.Equal(t, "ghcr.io/hivepaas/hivepaas:0.1.0", stable.AppImage)
	assert.Equal(t, "redis:8", stable.RedisImage)
	assert.Equal(t, map[string]string{"node24": "n"}, stable.FunctionRuntimes)
	assert.Nil(t, stable.Templates)
}

// A release file that cannot be read stops the binary at start: it would
// otherwise run images nobody chose.
func TestAnUnreadableReleaseFileStopsTheBinary(t *testing.T) {
	assert.Panics(t, func() { compiledRelease([]byte(`{`), "beta", nil) })
	assert.Panics(t, func() { compiledRelease([]byte(`{}`), "beta", nil) }, "beta is always released")
}
