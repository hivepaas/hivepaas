package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func Test_LoadConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = os.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")

		SetCurrent(nil)
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

		SetCurrent(nil)
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

		SetCurrent(nil)
		cfg, err := LoadConfig()
		assert.Nil(t, err)
		assert.Equal(t, "db-from-section", cfg.DB.Host)
		assert.Equal(t, 25432, cfg.DB.Port)
	})

	t.Run("failure: no ENV to find config", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Unsetenv("HP_CONFIG_FILE")

		SetCurrent(nil)
		_, err := LoadConfig()
		assert.ErrorIs(t, err, ErrConfigFileUnset)
	})

	t.Run("failure: config not found", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Setenv("HP_CONFIG_FILE", "notexist/config.myenv.toml")

		SetCurrent(nil)
		_, err := LoadConfig()
		assert.ErrorIs(t, err, ErrConfigFileNotFound)
	})

	t.Run("failure: malformed TOML data", func(t *testing.T) {
		_ = os.Unsetenv("HP_ENV")
		_ = os.Setenv("HP_CONFIG_FILE", "testdata/config-malformed.toml")

		SetCurrent(nil)
		_, err := LoadConfig()
		assert.NotNil(t, err)
	})
}

// A deploy that waits for its pre-deploy jobs looks every 3 seconds, and gives a
// job without a timeout of its own 30 minutes.
func TestTaskTriggersDefaults(t *testing.T) {
	t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")

	SetCurrent(nil)
	t.Cleanup(func() { SetCurrent(nil) })
	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, 3*time.Second, cfg.Tasks.Triggers.WaitPollInterval)
	assert.Equal(t, 30*time.Minute, cfg.Tasks.Triggers.WaitTimeout)
}

// The agent's repository server keeps under 256 MiB and has 30 seconds to listen.
func TestAgentRepoServerDefaults(t *testing.T) {
	t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")

	SetCurrent(nil)
	t.Cleanup(func() { SetCurrent(nil) })
	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, "256MiB", cfg.Agent.RepoServer.MemLimit)
	assert.Equal(t, 30*time.Second, cfg.Agent.RepoServer.StartTimeout)
}

// The timezone is UTC unless given; given, schedules are read in it from then
// on. A name that is no zone stops the start.
func TestTimezone(t *testing.T) {
	t.Cleanup(func() {
		SetCurrent(nil)
		timeutil.SetLocation(nil)
	})
	// A load takes the HP_ variables out of the process: each is given again.
	load := func(timezone string) (*Config, error) {
		t.Setenv("HP_CONFIG_FILE", "testdata/config.myenv.toml")
		if timezone != "" {
			t.Setenv("HP_TIMEZONE", timezone)
		}
		SetCurrent(nil)
		return LoadConfig()
	}

	cfg, err := load("")
	if assert.NoError(t, err) {
		assert.Equal(t, "UTC", cfg.Timezone)
		assert.Equal(t, time.UTC, cfg.Location())
	}

	cfg, err = load("America/New_York")
	if assert.NoError(t, err) {
		assert.Equal(t, "America/New_York", cfg.Location().String())
		assert.Equal(t, "America/New_York", timeutil.Location().String())
	}

	for _, name := range []string{"Mars/Olympus_Mons", "Local", "+07:00"} {
		_, err = load(name)
		assert.ErrorIs(t, err, ErrTimezoneInvalid, name)
	}
}
