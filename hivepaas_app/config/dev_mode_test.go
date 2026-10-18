package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The features are names, comma-separated, and none at all where dev mode is off.
func TestLicenseFeatureList(t *testing.T) {
	on := &DevMode{Enabled: true, LicenseFeatures: " managed-databases , ,other"}
	assert.Equal(t, []string{"managed-databases", "other"}, on.LicenseFeatureList())

	assert.Nil(t, (&DevMode{Enabled: true}).LicenseFeatureList())
	assert.Nil(t, (&DevMode{LicenseFeatures: "managed-databases"}).LicenseFeatureList(),
		"production and beta never take them")
}
