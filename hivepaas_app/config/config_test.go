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

	// The env tag must win over the value in the config file, including for a
	// setting inside a nested section.
	t.Run("success with override a key with ENV", func(t *testing.T) {
		t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")
		t.Setenv("HP_DB_HOST", "db-from-env")

		Current = nil
		cfg, err := LoadConfig()
		assert.Nil(t, err)
		assert.Equal(t, "myenv", cfg.Env)
		assert.Equal(t, "myplatform", cfg.Platform)
		// testdata/config.myenv.toml sets db.host to "localhost"
		assert.Equal(t, "db-from-env", cfg.DB.Host)
		// A section value the env does not touch keeps what the file said.
		assert.Equal(t, 15432, cfg.DB.Port)
	})

	// The prefix configor derives names with is ours, not the library default, so
	// a section can be fed as a whole. Guards against the prefix silently changing.
	t.Run("success with override a whole section with the prefixed ENV", func(t *testing.T) {
		t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")
		t.Setenv("HP_DB", "host: db-from-section\nport: 25432")

		Current = nil
		cfg, err := LoadConfig()
		assert.Nil(t, err)
		assert.Equal(t, "db-from-section", cfg.DB.Host)
		assert.Equal(t, 25432, cfg.DB.Port)
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
