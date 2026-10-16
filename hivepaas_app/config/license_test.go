package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// With nothing configured, the portal is production's and the daily refresh is
// on. Either is changed by an env variable.
func TestLicenseDefaults(t *testing.T) {
	t.Cleanup(func() { SetCurrent(nil) })
	load := func(env map[string]string) (*Config, error) {
		t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")
		for name, value := range env {
			t.Setenv(name, value)
		}
		SetCurrent(nil)
		return LoadConfig()
	}

	cfg, err := load(nil)
	if assert.NoError(t, err) {
		assert.Equal(t, DefaultLicensePortalURL, cfg.License.PortalURL)
		assert.False(t, cfg.License.RefreshDisabled)
	}

	cfg, err = load(map[string]string{"HP_LICENSE_PORTAL_URL": "http://localhost:4100",
		"HP_LICENSE_REFRESH_DISABLED": "true"})
	if assert.NoError(t, err) {
		assert.Equal(t, "http://localhost:4100", cfg.License.PortalURL)
		assert.True(t, cfg.License.RefreshDisabled)
	}

	_, err = load(map[string]string{"HP_LICENSE_PORTAL_URL": "portal.hivepaas.com"})
	assert.ErrorIs(t, err, ErrLicensePortalURLInvalid)
}

func TestLicenseValidate(t *testing.T) {
	for _, valid := range []string{"https://portal.hivepaas.com", "http://localhost:4100", "https://p.example/sub"} {
		assert.NoError(t, License{PortalURL: valid}.Validate(), valid)
	}
	for _, invalid := range []string{"", "portal.hivepaas.com", "ftp://portal.hivepaas.com", "https://", "://x"} {
		assert.ErrorIs(t, License{PortalURL: invalid}.Validate(), ErrLicensePortalURLInvalid, invalid)
	}
}

// The portal's pages carry the installation's ID when it is known, escaped.
func TestLicensePortalURLs(t *testing.T) {
	cfg := &Config{License: License{PortalURL: "https://portal.example/"}}

	assert.Equal(t, "https://portal.example/api/licenses/refresh", cfg.LicensePortalRefreshURL())
	assert.Equal(t, "https://portal.example/pricing?installationId=x2kq7", cfg.LicensePortalPricingURL("x2kq7"))
	assert.Equal(t, "https://portal.example/pricing", cfg.LicensePortalPricingURL(""))
	assert.Equal(t, "https://portal.example/trial?installationId=a%26b", cfg.LicensePortalTrialURL("a&b"))
	assert.Equal(t, "https://portal.example/account/licenses", cfg.LicensePortalAccountURL())

	cfg.License.PortalURL = "https://portal.example/sub"
	assert.Equal(t, "https://portal.example/sub/api/licenses/refresh", cfg.LicensePortalRefreshURL())
}
