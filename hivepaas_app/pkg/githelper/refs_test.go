package githelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRepoRef(t *testing.T) {
	tests := []struct {
		input    string
		expected ReferenceName
	}{
		{"", "HEAD"},
		{"refs/heads/main", "refs/heads/main"},
		{"tags/v1.0", NewTagReferenceName("v1.0")},
		{"heads/feature", NewBranchReferenceName("feature")},
		{"somebranch", NewBranchReferenceName("somebranch")},
	}

	for _, tt := range tests {
		result := NormalizeRepoRef(tt.input)
		assert.Equal(t, tt.expected, result, "NormalizeRepoRef(%s)", tt.input)
	}
}

func TestIsCommitHash(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", true},
		{"A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2", true},
		{"a1B2c3D4e5F6a1B2c3D4e5F6a1B2c3D4e5F6a1B2", true},
		{"", false},
		{"g1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", false},
		{"a1b2c3d4", false},
		{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2a1", false},
		{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b ", false},
	}

	for _, tt := range tests {
		result := IsCommitHash(tt.input)
		assert.Equal(t, tt.expected, result, "IsCommitHash(%s)", tt.input)
	}
}
