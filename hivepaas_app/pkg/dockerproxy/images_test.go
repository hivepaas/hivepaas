package dockerproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchImage(t *testing.T) {
	tests := []struct {
		patterns []string
		ref      string
		want     bool
	}{
		{[]string{"*"}, "anything/at:all", true},
		{[]string{"*"}, "", false},
		{nil, "alpine", false},
		// A pattern without a tag takes every tag and digest of its repository.
		{[]string{"alpine"}, "alpine:3", true},
		{[]string{"alpine"}, "docker.io/library/alpine:3.22", true},
		{[]string{"alpine"}, "alpine@sha256:abc", true},
		// No tag is latest.
		{[]string{"alpine:3"}, "alpine", false},
		{[]string{"alpine:3"}, "alpine:3", true},
		{[]string{"alpine:3.*"}, "alpine:3.22", true},
		{[]string{"autobase/automation:2.11.0"}, "autobase/automation:2.11.0", true},
		{[]string{"autobase/automation:2.11.0"}, "autobase/automation:2.12.0", false},
		{[]string{"autobase/automation:2.11.0"}, "autobase/console:2.11.0", false},
		{[]string{"openruntimes/*"}, "openruntimes/node:v5-22", true},
		{[]string{"openruntimes/*"}, "evil/node:v5-22", false},
		{[]string{"ghcr.io/nextcloud/*"}, "ghcr.io/nextcloud/app:1", true},
		{[]string{"ghcr.io/nextcloud/*"}, "docker.io/nextcloud/app:1", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, matchImage(tt.patterns, tt.ref), "%v %s", tt.patterns, tt.ref)
	}
}
