package fileutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func TestTypeByExtension(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{
			name:     "with dot",
			ext:      ".json",
			expected: "application/json",
		},
		{
			name:     "without dot",
			ext:      "json",
			expected: "application/json",
		},
		{
			name:     "html with dot",
			ext:      ".html",
			expected: "text/html; charset=utf-8",
		},
		{
			name:     "html without dot",
			ext:      "html",
			expected: "text/html; charset=utf-8",
		},
		{
			name:     "png",
			ext:      "png",
			expected: "image/png",
		},
		{
			name:     "unknown extension",
			ext:      ".unknownext123",
			expected: "",
		},
		{
			name:     "unknown extension without dot",
			ext:      "unknownext123",
			expected: "",
		},
		{
			name:     "empty extension",
			ext:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileutil.TypeByExtension(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}
