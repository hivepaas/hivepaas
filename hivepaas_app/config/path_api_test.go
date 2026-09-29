package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A login option's link to its provider is under the API's base path, as
// configured.
func TestSsoAuthPath(t *testing.T) {
	cfg := &Config{}
	cfg.HTTPServer.BasePath = "/api"

	assert.Equal(t, "/api/auth/sso/01JAB9XED0GTXBSQDFVYAJ8WJ1", cfg.SsoAuthPath("01JAB9XED0GTXBSQDFVYAJ8WJ1"))
}

// The API is under /api unless the configuration says otherwise.
func TestAPIBasePathDefault(t *testing.T) {
	t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")

	SetCurrent(nil)
	t.Cleanup(func() { SetCurrent(nil) })
	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, "/api", cfg.HTTPServer.BasePath)
}
