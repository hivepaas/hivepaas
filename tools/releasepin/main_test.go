package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// digests is a registry that knows these tags.
type digests map[string]string

func (d digests) resolve(ref string) (string, error) {
	if digest, ok := d[ref]; ok {
		return digest, nil
	}
	return "", errors.New("manifest unknown: " + ref)
}

const release = `{
  "beta": {
    "appVersion": "v1.0.0-beta1",
    "appImage": "ghcr.io/hivepaas/hivepaas:1.0.0-beta1",
    "redisImage": "redis:8.6.2-alpine@sha256:old",
    "traefikImage": "traefik:v3.7",
    "blockMajorUpgrade": [
      "db"
    ]
  }
}
`

var registry = digests{
	"ghcr.io/hivepaas/hivepaas:1.0.0-beta1": "sha256:app",
	"redis:8.6.2-alpine":                    "sha256:new",
	"traefik:v3.7":                          "sha256:traefik",
}

// Every image of every channel is pinned to the digest its tag has now, and
// nothing else in the file moves: it is signed byte for byte afterwards.
func TestEveryImageIsPinnedAndNothingElseMoves(t *testing.T) {
	out, report, err := pin([]byte(release), registry.resolve)

	assert.NoError(t, err)
	want := strings.NewReplacer(
		`"ghcr.io/hivepaas/hivepaas:1.0.0-beta1"`, `"ghcr.io/hivepaas/hivepaas:1.0.0-beta1@sha256:app"`,
		`"redis:8.6.2-alpine@sha256:old"`, `"redis:8.6.2-alpine@sha256:new"`,
		`"traefik:v3.7"`, `"traefik:v3.7@sha256:traefik"`,
	).Replace(release)
	assert.Equal(t, want, string(out))
	assert.Len(t, report.Changed, 3)
}

// A tag that names only a minor is pinned all the same, and flagged: the
// digest fixes the image, but its name no longer says which patch it is.
func TestATagWithoutAPatchIsFlagged(t *testing.T) {
	_, report, err := pin([]byte(release), registry.resolve)

	assert.NoError(t, err)
	assert.Equal(t, []string{"traefik:v3.7"}, report.Floating)
}

// Pinning what is already pinned to the current digest changes nothing.
func TestAPinnedFileIsLeftAsItIs(t *testing.T) {
	once, _, err := pin([]byte(release), registry.resolve)
	assert.NoError(t, err)

	twice, report, err := pin(once, registry.resolve)

	assert.NoError(t, err)
	assert.Equal(t, string(once), string(twice))
	assert.Empty(t, report.Changed)
}

// An image the registry does not know stops the run: a release must not ship
// with one image pinned and another not.
func TestAnImageTheRegistryDoesNotKnowIsAnError(t *testing.T) {
	_, _, err := pin([]byte(release), digests{}.resolve)

	assert.ErrorContains(t, err, "manifest unknown")
}

const releaseWithFunctions = `{
  "beta": {
    "appImage": "ghcr.io/hivepaas/hivepaas:1.0.0-beta1@sha256:app",
    "functionRuntimes": {
      "node24": "ghcr.io/hivepaas/function-runtime-node24:1.0.0",
      "go127-build": "ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:old"
    },
    "blockMajorUpgrade": [
      "db"
    ]
  }
}
`

// The images functions are built on are released with HivePaaS too, in a map
// of their own, and are pinned like every other image.
func TestFunctionRuntimesArePinnedToo(t *testing.T) {
	registry := digests{
		"ghcr.io/hivepaas/hivepaas:1.0.0-beta1":               "sha256:app",
		"ghcr.io/hivepaas/function-runtime-node24:1.0.0":      "sha256:node",
		"ghcr.io/hivepaas/function-runtime-go127-build:1.0.0": "sha256:gobuild",
	}

	out, report, err := pin([]byte(releaseWithFunctions), registry.resolve)

	assert.NoError(t, err)
	want := strings.NewReplacer(
		`"ghcr.io/hivepaas/function-runtime-node24:1.0.0"`,
		`"ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:node"`,
		`"ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:old"`,
		`"ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:gobuild"`,
	).Replace(releaseWithFunctions)
	assert.Equal(t, want, string(out))
	assert.Len(t, report.Changed, 2)
	assert.Empty(t, report.Floating)
}

// The app and the agent of a release are built after its release commit: before
// that their tags are not in the registry, and they are left to be pinned once
// they are, while every other image is pinned now.
func TestTheAppAndAgentNotBuiltYetAreLeftForLater(t *testing.T) {
	const unbuilt = `{
  "beta": {
    "appImage": "ghcr.io/hivepaas/hivepaas:1.0.0-beta2",
    "agentImage": "ghcr.io/hivepaas/hivepaas-agent:1.0.0-beta2",
    "redisImage": "redis:8.6.2-alpine"
  }
}
`
	out, report, err := pin([]byte(unbuilt), digests{"redis:8.6.2-alpine": "sha256:new"}.resolve)

	assert.NoError(t, err)
	assert.Equal(t, strings.Replace(unbuilt, `"redis:8.6.2-alpine"`, `"redis:8.6.2-alpine@sha256:new"`, 1), string(out))
	assert.Equal(t, []string{"ghcr.io/hivepaas/hivepaas:1.0.0-beta2", "ghcr.io/hivepaas/hivepaas-agent:1.0.0-beta2"},
		report.NotBuilt)
}
