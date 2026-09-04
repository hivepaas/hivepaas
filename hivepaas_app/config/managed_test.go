package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetLoadState clears the package globals a load leaves behind, so each test
// starts from the state a fresh process would have.
func resetLoadState() {
	Current = nil
	lastConfigFile = ""
	envSnapshotOnce = sync.Once{}
	envSnapshot = nil
}

// writeManagedSettings writes a managed settings file into a temp app path.
func writeManagedSettings(t *testing.T, content string, perm os.FileMode) string {
	t.Helper()
	appPath := t.TempDir()
	assert.NoError(t, os.WriteFile(ManagedSettingsPath(appPath), []byte(content), perm))
	return appPath
}

func TestLoadManagedSettings(t *testing.T) {
	t.Run("a missing file is not an error", func(t *testing.T) {
		settings, err := loadManagedSettings(t.TempDir())
		assert.NoError(t, err)
		assert.Equal(t, "", settings.Secret)
	})

	t.Run("reads the secret", func(t *testing.T) {
		appPath := writeManagedSettings(t, `secret = "from-managed-file"`, 0o600)

		settings, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, "from-managed-file", settings.Secret)
	})

	// The reason for choosing TOML: comments and multi-line literals, so a PEM key
	// can be pasted in verbatim one day without escaping anything.
	t.Run("supports comments and multi-line literals", func(t *testing.T) {
		appPath := writeManagedSettings(t, `
# rotated by the app on 2026-09-04
secret = '''
line one
line two
'''
`, 0o600)

		settings, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, "line one\nline two\n", settings.Secret)
	})

	// The struct is the allowlist: a writable app volume must not let anyone
	// repoint the database.
	t.Run("a key outside the allowlist is ignored", func(t *testing.T) {
		appPath := writeManagedSettings(t, `
secret = "kept"

[db]
  host = "attacker-db"
`, 0o600)

		settings, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, "kept", settings.Secret)
	})

	t.Run("a world-readable file is refused", func(t *testing.T) {
		appPath := writeManagedSettings(t, `secret = "leaked"`, 0o644)

		_, err := loadManagedSettings(appPath)
		assert.ErrorIs(t, err, ErrManagedSettingsPermissive)
	})

	t.Run("malformed TOML is reported", func(t *testing.T) {
		appPath := writeManagedSettings(t, `secret = `, 0o600)

		_, err := loadManagedSettings(appPath)
		assert.Error(t, err)
	})
}

func TestManagedSettingsApplyTo(t *testing.T) {
	t.Run("overrides the loaded value", func(t *testing.T) {
		config := &Config{Secret: "from-env"}
		(&ManagedSettings{Secret: "from-file"}).applyTo(config)
		assert.Equal(t, "from-file", config.Secret)
	})

	// A partial file must not wipe what the base config or the environment set.
	t.Run("an unset field leaves the loaded value alone", func(t *testing.T) {
		config := &Config{Secret: "from-env"}
		(&ManagedSettings{}).applyTo(config)
		assert.Equal(t, "from-env", config.Secret)
	})

	t.Run("a nil receiver is a no-op", func(t *testing.T) {
		config := &Config{Secret: "from-env"}
		var settings *ManagedSettings
		settings.applyTo(config)
		assert.Equal(t, "from-env", config.Secret)
	})
}

func TestLoadConfigAppliesManagedSettingsOverEnv(t *testing.T) {
	resetLoadState()
	appPath := writeManagedSettings(t, `secret = "rotated-by-the-app"`, 0o600)
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\nsecret = \"from-config-file\"\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_APP_SECRET", "stale-value-still-in-compose")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "rotated-by-the-app", cfg.Secret)
}

// Clearing HP_* keeps the app secret out of the environment handed to the
// commands users run, which are spawned with os.Environ().
func TestLoadConfigClearsEnvAfterwards(t *testing.T) {
	resetLoadState()
	appPath := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\nsecret = \"test-app-secret-value\"\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_APP_SECRET", "must-not-linger")

	_, err := LoadConfig()
	assert.NoError(t, err)
	assert.Empty(t, os.Getenv("HP_APP_SECRET"), "HP_* must not survive the load")
	assert.Empty(t, os.Getenv("HP_APP_PATH"))
}

// The regression the snapshot exists for: clearing the environment must not make
// a reload lose the settings that came from it.
func TestReloadConfigKeepsEnvValues(t *testing.T) {
	resetLoadState()
	appPath := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\nsecret = \"test-app-secret-value\"\n\n[db]\n  host = \"from-config-file\"\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_DB_HOST", "from-env")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "from-env", cfg.DB.Host)

	reloaded, err := ReloadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "from-env", reloaded.DB.Host,
		"a reload must still see the environment the process started with")
}

func TestEnsureAppSecret(t *testing.T) {
	t.Run("keeps a configured secret", func(t *testing.T) {
		config := &Config{Env: EnvDev, Secret: "already-set"}
		assert.NoError(t, ensureAppSecret(config, t.TempDir()))
		assert.Equal(t, "already-set", config.Secret)
	})

	// A generated secret must outlive the process, or the next start would invent
	// a different one and everything written so far becomes unreadable.
	t.Run("generates and persists one in development", func(t *testing.T) {
		appPath := t.TempDir()
		config := &Config{Env: EnvDev}

		assert.NoError(t, ensureAppSecret(config, appPath))
		assert.Len(t, config.Secret, appSecretLen*2) // hex encoded

		saved, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, config.Secret, saved.Secret)
	})

	// Generating outside development would be unsafe: replicas do not share the
	// app volume, so each would invent its own key.
	t.Run("refuses to start without one outside development", func(t *testing.T) {
		config := &Config{Env: EnvProd}
		err := ensureAppSecret(config, t.TempDir())
		assert.ErrorIs(t, err, ErrAppSecretUnset)
		assert.Empty(t, config.Secret)
	})
}

func TestSaveManagedSettings(t *testing.T) {
	t.Run("writes a file only the owner can read", func(t *testing.T) {
		appPath := t.TempDir()
		assert.NoError(t, saveManagedSettings(appPath, &ManagedSettings{Secret: "written"}))

		info, err := os.Stat(ManagedSettingsPath(appPath))
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		// loadManagedSettings refuses anything more permissive, so a round trip
		// also proves the two agree on the mode.
		loaded, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, "written", loaded.Secret)
	})

	t.Run("replaces an existing file and leaves no temp behind", func(t *testing.T) {
		appPath := t.TempDir()
		assert.NoError(t, saveManagedSettings(appPath, &ManagedSettings{Secret: "first"}))
		assert.NoError(t, saveManagedSettings(appPath, &ManagedSettings{Secret: "second"}))

		loaded, err := loadManagedSettings(appPath)
		assert.NoError(t, err)
		assert.Equal(t, "second", loaded.Secret)

		entries, err := os.ReadDir(appPath)
		assert.NoError(t, err)
		assert.Len(t, entries, 1, "the temp file must not survive the write")
	})
}
