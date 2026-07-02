package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Example@Domain.COM", "example@domain.com"},
		{"  spaced@Example.com  ", "spaced@example.com"},
		{"UPPERCASE@UPPER.COM", "uppercase@upper.com"},
		{"already.lower@site.org", "already.lower@site.org"},
		{"", ""},
	}

	for _, c := range cases {
		result := NormalizeEmail(c.input)
		assert.Equal(t, c.expected, result, "NormalizeEmail(%q)", c.input)
	}
}
