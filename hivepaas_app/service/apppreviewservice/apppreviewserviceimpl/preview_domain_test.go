package apppreviewserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A preview is beside its app, where the app's wildcard certificate and DNS
// record already cover it - under it only where beside would leave the zone.
func TestPreviewDomain(t *testing.T) {
	for _, tc := range []struct {
		subdomain, domain, want string
	}{
		{"pr-42", "shop.example.com", "pr-42-shop.example.com"},
		{"pr-42", "api.shop.example.com", "pr-42-api.shop.example.com"},
		{"pr-42", "Shop.Example.com.", "pr-42-shop.example.com"},
		// The zone's apex: beside it would be another zone.
		{"pr-42", "example.com", "pr-42.example.com"},
		{"pr-42", "shop.co.uk", "pr-42.shop.co.uk"},
		{"pr-42", "api.shop.co.uk", "pr-42-api.shop.co.uk"},
		// A subdomain asked for whole, or with dots in it, is taken as given.
		{"pr-42.shop.example.com", "shop.example.com", "pr-42-shop.example.com"},
		{"review.pr-42", "shop.example.com", "review.pr-42.shop.example.com"},
		// Names no public suffix knows.
		{"pr-42", "app.localhost", "pr-42.app.localhost"},
		{"pr-42", "web.app.localhost", "pr-42-web.app.localhost"},
	} {
		got, err := previewDomain(tc.subdomain, tc.domain)
		assert.NoError(t, err, tc.domain)
		assert.Equal(t, tc.want, got, "%s of %s", tc.subdomain, tc.domain)
	}
}

func TestPreviewDomainLabelTooLong(t *testing.T) {
	_, err := previewDomain("pr-42", strings.Repeat("a", 58)+".example.com")
	assert.Error(t, err, "pr-42- and 58 letters is 64: one more than a label may be")
	_, err = previewDomain("pr-42", strings.Repeat("a", 57)+".example.com")
	assert.NoError(t, err)
}
