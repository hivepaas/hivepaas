package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeFirstBootEnv(t *testing.T, appPath, content string, perm os.FileMode) {
	t.Helper()
	assert.NoError(t, os.WriteFile(FirstBootEnvPath(appPath), []byte(content), perm))
}

func TestReadFirstBootEnv(t *testing.T) {
	t.Run("a missing file is not an error", func(t *testing.T) {
		values, err := readFirstBootEnv(FirstBootEnvPath(t.TempDir()))
		assert.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("reads KEY=VALUE lines, the value as it stands", func(t *testing.T) {
		appPath := t.TempDir()
		writeFirstBootEnv(t, appPath, "# written by install.sh\n\n"+
			"HP_USER_ADMIN_EMAIL=you@example.com\r\n"+
			"HP_USER_ADMIN_PASSWORD=p=ss \"w0rd\" 'x' $y\n", 0o600)

		values, err := readFirstBootEnv(FirstBootEnvPath(appPath))

		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"HP_USER_ADMIN_EMAIL":    "you@example.com",
			"HP_USER_ADMIN_PASSWORD": `p=ss "w0rd" 'x' $y`,
		}, values)
	})

	t.Run("refuses what is not a setting of this app", func(t *testing.T) {
		appPath := t.TempDir()
		writeFirstBootEnv(t, appPath, "PATH=/nowhere\n", 0o600)
		_, err := readFirstBootEnv(FirstBootEnvPath(appPath))
		assert.ErrorContains(t, err, "PATH")
	})

	t.Run("refuses a line that is not KEY=VALUE", func(t *testing.T) {
		appPath := t.TempDir()
		writeFirstBootEnv(t, appPath, "HP_USER_ADMIN_EMAIL\n", 0o600)
		_, err := readFirstBootEnv(FirstBootEnvPath(appPath))
		assert.ErrorContains(t, err, "line 1")
	})

	t.Run("refuses a file others can read", func(t *testing.T) {
		appPath := t.TempDir()
		writeFirstBootEnv(t, appPath, "HP_USER_ADMIN_PASSWORD=secret\n", 0o644)
		_, err := readFirstBootEnv(FirstBootEnvPath(appPath))
		assert.True(t, errors.Is(err, ErrFirstBootEnvPermissive), "got %v", err)
	})
}

func TestLoadConfigReadsTheFirstBootEnv(t *testing.T) {
	resetLoadState()
	appPath := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\nsecret = \"from-config-file\"\n"), 0o600))
	writeFirstBootEnv(t, appPath, "HP_USER_ADMIN_PASSWORD=from-the-file\nHP_USER_ADMIN_EMAIL=file@example.com\n",
		0o600)
	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_USER_ADMIN_EMAIL", "env@example.com")

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, "from-the-file", cfg.Users.Admin.Password)
	assert.Equal(t, "env@example.com", cfg.Users.Admin.Email, "the service's environment wins")
	assert.Empty(t, os.Getenv("HP_USER_ADMIN_PASSWORD"), "cleared from the process like the rest")
}

func TestRemoveFirstBootEnv(t *testing.T) {
	appPath := t.TempDir()
	t.Setenv("HP_APP_PATH", appPath)
	writeFirstBootEnv(t, appPath, "HP_USER_ADMIN_PASSWORD=secret\n", 0o600)

	assert.NoError(t, RemoveFirstBootEnv())
	_, err := os.Stat(FirstBootEnvPath(appPath))
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.NoError(t, RemoveFirstBootEnv(), "nothing to remove is not an error")
}
