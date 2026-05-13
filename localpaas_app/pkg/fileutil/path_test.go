package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubpath(t *testing.T) {
	tests := []struct {
		base     string
		target   string
		expected bool
		wantErr  bool
	}{
		{"/a/b", "/a/b/c", true, false},
		{"/a/b", "/a/b", false, false}, // IsSubpath should be false for equal paths
		{"/a/b", "/a/x", false, false},
		{"/a/b", "/a/b/../x", false, false},
		{"a/b", "a/b/c", true, false},
		{"a/b", "a/x", false, false},
		{"/a/b", "a/b/c", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.base+"_"+tt.target, func(t *testing.T) {
			got, err := IsSubpath(tt.base, tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestIsEqualOrSubpath(t *testing.T) {
	tests := []struct {
		base     string
		target   string
		expected bool
		wantErr  bool
	}{
		{"/a/b", "/a/b/c", true, false},
		{"/a/b", "/a/b", true, false}, // IsEqualOrSubpath should be true for equal paths
		{"/a/b", "/a/x", false, false},
		{"/a/b", "/a/b/../x", false, false},
		{"a/b", "a/b/c", true, false},
		{"a/b", "a/x", false, false},
		{"/a/b", "a/b/c", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.base+"_"+tt.target, func(t *testing.T) {
			got, err := IsEqualOrSubpath(tt.base, tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestIsSamePath(t *testing.T) {
	tests := []struct {
		path1    string
		path2    string
		expected bool
		wantErr  bool
	}{
		{"/a/b", "/a/b", true, false},
		{"/a/b", "/a/b/", true, false},
		{"/a//b", "/a/b/", true, false},
		{"/a/b", "/a/b/../b", true, false},
		{"/a/b", "/a/c", false, false},
		{"a/b", "a/b", true, false},
		{"a/b", "a/b/../b", true, false},
		{"a/b", "a/c", false, false},
		{"/a/b", "a/b", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.path1+"_"+tt.path2, func(t *testing.T) {
			got, err := IsSamePath(tt.path1, tt.path2)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestPathContain(t *testing.T) {
	tests := []struct {
		name      string
		listPaths []string
		aPath     string
		expected  bool
		wantErr   bool
	}{
		{
			name:      "path contained exactly",
			listPaths: []string{"/a/b", "/c/d"},
			aPath:     "/a/b",
			expected:  true,
			wantErr:   false,
		},
		{
			name:      "path contained with normalization",
			listPaths: []string{"/a/b", "/c/d"},
			aPath:     "/a/b/../b",
			expected:  true,
			wantErr:   false,
		},
		{
			name:      "path not contained",
			listPaths: []string{"/a/b", "/c/d"},
			aPath:     "/e/f",
			expected:  false,
			wantErr:   false,
		},
		{
			name:      "empty list",
			listPaths: []string{},
			aPath:     "/a/b",
			expected:  false,
			wantErr:   false,
		},
		{
			name:      "absolute relative mismatch yields error",
			listPaths: []string{"/a/b"},
			aPath:     "a/b",
			expected:  false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathContain(tt.listPaths, tt.aPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/a/b/../c", "/a/c"},
		{"/a/b/.", "/a/b"},
		{"/a//b", "/a/b"},
		{"a/b/../c", "a/c"},
		{"a/b/.", "a/b"},
		{"a//b", "a/b"},
		{"", "."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizePath(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
