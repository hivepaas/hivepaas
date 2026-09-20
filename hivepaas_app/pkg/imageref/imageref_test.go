package imageref_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

func TestParseSplitsEveryShapeTheUpdaterSees(t *testing.T) {
	tests := []struct {
		ref  string
		want imageref.Ref
	}{
		{"redis:8.6-alpine", imageref.Ref{Repository: "redis", Tag: "8.6-alpine"}},
		{"victoriametrics/victoria-logs:v1.52.0",
			imageref.Ref{Repository: "victoriametrics/victoria-logs", Tag: "v1.52.0"}},
		// What a running service actually reports: Docker Desktop pins the digest
		// of everything it deploys.
		{"hivepaas/hivepaas-dev:latest@sha256:6ecdf4e6",
			imageref.Ref{Repository: "hivepaas/hivepaas-dev", Tag: "latest", Digest: "sha256:6ecdf4e6"}},
		// The colon here is a port, not a tag separator.
		{"registry.example:5000/redis", imageref.Ref{Repository: "registry.example:5000/redis"}},
		{"registry.example:5000/redis:8.6", imageref.Ref{Repository: "registry.example:5000/redis", Tag: "8.6"}},
		{"redis", imageref.Ref{Repository: "redis"}},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, imageref.Parse(tt.ref), tt.ref)
	}
}

func TestIsUpgradeOnlyWhenTheTargetIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		target  string
		want    bool
	}{
		{"newer patch", "victoriametrics/victoria-logs:v1.52.0", "victoriametrics/victoria-logs:v1.52.1", true},
		{"newer minor", "redis:8.6-alpine", "redis:8.7-alpine", true},
		{"same tag", "traefik:v3.7", "traefik:v3.7", false},
		{"older target is refused", "redis:8.7-alpine", "redis:8.6-alpine", false},
		{"much older target is refused", "postgres:18.3-alpine", "postgres:17.9-alpine", false},
		{"1.2 and 1.2.0 are the same version", "x/y:1.2", "x/y:1.2.0", false},
		{"1.2 to 1.2.1", "x/y:1.2", "x/y:1.2.1", true},
		{"no target", "redis:8.6-alpine", "", false},
		{"nothing running yet", "", "redis:8.6-alpine", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := imageref.IsUpgrade(tt.current, tt.target)
			assert.Equal(t, tt.want, got, reason)
			assert.NotEmpty(t, reason, "every answer has to be explainable")
		})
	}
}

// The digest moves whenever a tag is re-pushed. That is the same version, and
// restarting the service for it is exactly what the version check exists to
// avoid.
func TestIsUpgradeIgnoresADigestUnderTheSameTag(t *testing.T) {
	got, reason := imageref.IsUpgrade(
		"victoriametrics/victoria-logs:v1.52.0@sha256:aaaa",
		"victoriametrics/victoria-logs:v1.52.0",
	)

	assert.False(t, got, reason)
}

// A release moving to a different repository has made a decision there is no
// version line to second-guess.
func TestIsUpgradeAppliesADifferentRepository(t *testing.T) {
	got, reason := imageref.IsUpgrade("redis:8.6-alpine", "valkey/valkey:8.6-alpine")

	assert.True(t, got)
	assert.Contains(t, reason, "valkey/valkey")
}

// Where it cannot order the two, it applies and says why - refusing would block
// a release that is perfectly legitimate, and silence would hide the reason.
func TestIsUpgradeAppliesWhatItCannotOrder(t *testing.T) {
	tests := []struct{ current, target string }{
		{"redis:latest", "redis:8.6-alpine"},
		{"redis:8.6-alpine", "redis:latest"},
		// Same numbers, different suffix: a variant change and a prerelease
		// leaving prerelease have the same shape and opposite meanings.
		{"redis:8.6-alpine", "redis:8.6-bookworm"},
		{"x/y:1.0.0-beta1", "x/y:1.0.0"},
		{"redis", "redis:8.6-alpine"},
	}

	for _, tt := range tests {
		got, reason := imageref.IsUpgrade(tt.current, tt.target)
		assert.True(t, got, "%s -> %s: %s", tt.current, tt.target, reason)
	}
}

func TestMajorVersionReadsTheLeadingNumber(t *testing.T) {
	tests := []struct {
		ref   string
		want  int
		known bool
	}{
		{"postgres:18.3-alpine", 18, true},
		{"postgres:17.9-alpine", 17, true},
		{"victoriametrics/victoria-logs:v1.52.0", 1, true},
		{"redis:8.6-alpine@sha256:aaaa", 8, true},
		{"postgres:latest", 0, false},
		{"postgres", 0, false},
	}

	for _, tt := range tests {
		got, known := imageref.MajorVersion(tt.ref)
		assert.Equal(t, tt.known, known, tt.ref)
		assert.Equal(t, tt.want, got, tt.ref)
	}
}

// Postgres publishes 19beta1 and 19beta2 long before 19, and those are exactly
// the images a major-version check most needs to see.
func TestVersionsAreReadOutOfPrereleaseTags(t *testing.T) {
	got, known := imageref.MajorVersion("postgres:19beta2-alpine")
	assert.True(t, known)
	assert.Equal(t, 19, got)

	upgrade, reason := imageref.IsUpgrade("postgres:18.6-alpine", "postgres:19beta2-alpine")
	assert.True(t, upgrade, reason)

	// And the other direction is still refused.
	upgrade, reason = imageref.IsUpgrade("postgres:19beta2-alpine", "postgres:18.6-alpine")
	assert.False(t, upgrade, reason)
}

func TestCompareTagsOrdersWhatItCan(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
		ok   bool
	}{
		{"18.6", "18.7", -1, true},
		{"18.7-alpine3.24", "18.6-alpine3.24", 1, true},
		{"18.6", "18.6", 0, true},
		{"18.6-alpine", "18.6", 0, false},
		{"stable", "18.6", 0, false},
	} {
		got, ok := imageref.CompareTags(tc.a, tc.b)
		assert.Equal(t, tc.ok, ok, "%s vs %s", tc.a, tc.b)
		if tc.ok {
			assert.Equal(t, tc.want, got, "%s vs %s", tc.a, tc.b)
		}
	}
}

func TestNormalizeRepository(t *testing.T) {
	cases := map[string]string{
		// Docker Hub, written the four ways people write it.
		"postgres":                              "docker.io/library/postgres",
		"docker.io/library/postgres":            "docker.io/library/postgres",
		"index.docker.io/library/postgres":      "docker.io/library/postgres",
		"registry-1.docker.io/library/postgres": "docker.io/library/postgres",
		// A namespace on Docker Hub, with and without the host.
		"grafana/grafana-oss":           "docker.io/grafana/grafana-oss",
		"docker.io/grafana/grafana-oss": "docker.io/grafana/grafana-oss",
		// Another registry is left alone: its first segment is a host because it
		// carries a dot, and nothing is added under it.
		"ghcr.io/immich-app/immich-server": "ghcr.io/immich-app/immich-server",
		"quay.io/coreos/etcd":              "quay.io/coreos/etcd",
		// A host is a host because of the dot, the colon or the word localhost.
		"registry.example:5000/redis": "registry.example:5000/redis",
		"localhost/redis":             "localhost/redis",
		"":                            "",
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, want, imageref.NormalizeRepository(input))
		})
	}
}

func TestSameRepository(t *testing.T) {
	// The report that prompted this: a template pinning postgres, and an image
	// pasted in the form a registry shows it.
	assert.True(t, imageref.SameRepository(
		"postgres:18.6-alpine3.24",
		"registry-1.docker.io/library/postgres:18.6-alpine3.23"))
	assert.True(t, imageref.SameRepository("postgres:18.6-alpine3.24", "docker.io/library/postgres:18.6"))
	assert.True(t, imageref.SameRepository("grafana/grafana-oss:13.0.2", "docker.io/grafana/grafana-oss:13.1.0"))

	// Still different repositories, however they are spelled.
	assert.False(t, imageref.SameRepository("postgres:18.6", "mysql:8.4.11"))
	assert.False(t, imageref.SameRepository("postgres:18.6", "ghcr.io/library/postgres:18.6"))
	assert.False(t, imageref.SameRepository("postgres:18.6", "mirror.example.com/library/postgres:18.6"))
}
