package domainserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertCoversExactNamesAndOneWildcardLabel(t *testing.T) {
	cases := map[string]struct {
		certDomain string
		domain     string
		want       bool
	}{
		"the same name":              {"app.example.com", "app.example.com", true},
		"a different name":           {"other.example.com", "app.example.com", false},
		"a wildcard over one label":  {"*.example.com", "app.example.com", true},
		"a wildcard over the parent": {"*.example.com", "example.com", false},
		"a wildcard over two labels": {"*.example.com", "one.more.example.com", false},
		"a wildcard one level down":  {"*.more.example.com", "one.more.example.com", true},
		"a wildcard against itself":  {"*.example.com", "*.example.com", true},
		"a certificate with no name": {"", "app.example.com", false},
		"case and spacing":           {"  APP.Example.com ", "app.example.com", true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, certCovers(tc.certDomain, tc.domain))
		})
	}
}
