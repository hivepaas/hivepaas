package domainhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubdomain(t *testing.T) {
	tests := []struct {
		domain   string
		sub      string
		expected bool
	}{
		{"example.com", "sub.example.com", true},
		{"example.com", "another.sub.example.com", true},
		{"example.com", "example.com", false},
		{"*.example.com", "sub.example.com", true},
		{"example.com", "*.sub.example.com", true},
		{"google.com", "example.com", false},
		{"com", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain+"_"+tt.sub, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSubdomain(tt.domain, tt.sub))
		})
	}
}

func TestCalcMatchingDomains(t *testing.T) {
	tests := []struct {
		subdomain string
		expected  []string
	}{
		{
			"a.b.c",
			[]string{"a.b.c", "b.c", "*.b.c", "c", "*.c"},
		},
		{
			"*.a.b",
			[]string{"*.a.b", "b", "*.b"},
		},
		{
			"example.com",
			[]string{"example.com", "com", "*.com"},
		},
		{
			"com",
			[]string{"com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.subdomain, func(t *testing.T) {
			got := CalcMatchingDomains(tt.subdomain)
			assert.ElementsMatch(t, tt.expected, got)
		})
	}
}

func TestIsSubdomainOrEqual(t *testing.T) {
	tests := []struct {
		domain   string
		sub      string
		expected bool
	}{
		{"example.com", "sub.example.com", true},
		{"example.com", "another.sub.example.com", true},
		{"example.com", "example.com", true},
		{"*.example.com", "sub.example.com", true},
		{"example.com", "*.sub.example.com", true},
		{"*.example.com", "*.example.com", true},
		{"google.com", "example.com", false},
		{"com", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain+"_"+tt.sub, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSubdomainOrEqual(tt.domain, tt.sub))
		})
	}
}
