package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = os.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")

		Current = nil
		cfg, err := LoadConfig()
		assert.Nil(t, err)
		assert.Equal(t, "myenv", cfg.Env)
		assert.Equal(t, "myplatform", cfg.Platform)
	})

	t.Run("success with override a key with ENV", func(t *testing.T) {
		_ = os.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")
		_ = os.Setenv("HP_APP_NAME", "overridden")

		Current = nil
		cfg, err := LoadConfig()
		assert.Nil(t, err)
		assert.Equal(t, "myenv", cfg.Env)
		assert.Equal(t, "myplatform", cfg.Platform)
	})

	t.Run("failure: no ENV to find config", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Unsetenv("HP_CONFIG_FILE")

		Current = nil
		_, err := LoadConfig()
		assert.ErrorIs(t, err, ErrConfigFileUnset)
	})

	t.Run("failure: config not found", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Setenv("HP_CONFIG_FILE", "notexist/config.myenv.toml")

		Current = nil
		_, err := LoadConfig()
		assert.ErrorIs(t, err, ErrConfigFileNotFound)
	})

	t.Run("failure: malformed TOML data", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Setenv("HP_CONFIG_FILE", "testdata/config-malformed.toml")

		Current = nil
		_, err := LoadConfig()
		assert.NotNil(t, err)
	})
}
