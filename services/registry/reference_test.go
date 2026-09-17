package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Docker's short forms are conveniences of the CLI. The V2 API wants a host and a
// full repository path, and library/ is only implied for single-segment names on
// Docker Hub.
func TestParseRepository(t *testing.T) {
	for image, want := range map[string]Reference{
		"postgres:18.6-alpine3.24":    {Host: "registry-1.docker.io", Name: "library/postgres"},
		"minio/minio:RELEASE.2025":    {Host: "registry-1.docker.io", Name: "minio/minio"},
		"ghcr.io/owner/app:1.2.3":     {Host: "ghcr.io", Name: "owner/app"},
		"registry.example:5000/app:1": {Host: "registry.example:5000", Name: "app"},
		"localhost:5000/team/app:1.0": {Host: "localhost:5000", Name: "team/app"},
		"quay.io/org/sub/app:2.0":     {Host: "quay.io", Name: "org/sub/app"},
	} {
		assert.Equal(t, want, ParseRepository(image), image)
	}
}

func TestReferenceString(t *testing.T) {
	assert.Equal(t, "registry-1.docker.io/library/postgres", ParseRepository("postgres:18.6").String())
}
