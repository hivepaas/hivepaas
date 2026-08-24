package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "my-cert",
			expected: "my-cert",
		},
		{
			input:    "*.example.com",
			expected: "wildcard.example.com",
		},
		{
			input:    "domain/with:invalid*chars?<>|",
			expected: "domain_with_invalidwildcardchars",
		},
		{
			input:    "  .leading-and-trailing.  ",
			expected: "leading-and-trailing",
		},
		{
			input:    "cert:name/test",
			expected: "cert_name_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := SanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
