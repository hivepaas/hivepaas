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
