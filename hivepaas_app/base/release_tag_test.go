package base_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/version"
)

// The release workflow runs this with RELEASE_TAG set to the tag it builds: a
// tag the binary does not know as its own version would be installed as one
// version and report another, and never be offered its successor. A tag with a
// pre-release suffix is a beta, whose version is BetaVersion's; any other is
// StableVersion's.
func TestReleaseTagIsTheCompiledVersion(t *testing.T) {
	tag := os.Getenv("RELEASE_TAG")
	if tag == "" {
		t.Skip("RELEASE_TAG is set by the release workflow")
	}
	assert.NoError(t, releaseTagMatches(tag))
}

func releaseTagMatches(tag string) error {
	parsed, err := version.Parse(tag)
	if err != nil || !strings.HasPrefix(tag, "v") {
		return fmt.Errorf("tag %s is not vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-beta<N>", tag)
	}
	compiled := base.StableVersion.AppVersion
	if parsed.Suffix != "" {
		compiled = base.BetaVersion.AppVersion
	}
	if compiled != tag {
		return fmt.Errorf("tag %s is not the version compiled into the binary (%s): "+
			"update hivepaas_app/base/version.go", tag, compiled)
	}
	return nil
}

func TestReleaseTagMatches(t *testing.T) {
	assert.NoError(t, releaseTagMatches(base.BetaVersion.AppVersion))
	assert.NoError(t, releaseTagMatches(base.StableVersion.AppVersion))
	assert.Error(t, releaseTagMatches("v9.9.9-beta1"), "a beta the binary is not")
	assert.Error(t, releaseTagMatches("v9.9.9"), "a stable release the binary is not")
	assert.Error(t, releaseTagMatches("v1.0.0-beta.1"), "a format the updater cannot compare")
	assert.Error(t, releaseTagMatches("1.0.0"), "no leading v")
}
