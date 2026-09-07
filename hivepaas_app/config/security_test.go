package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestSaveSecuritySettings(t *testing.T) {
	t.Run("persists and applies the settings", func(t *testing.T) {
		SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "app-secret"})

		assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: true}))
		assert.True(t, Current().Security.ReturnSecretsViaAPI)

		saved, err := loadManagedSettings(Current().AppPath)
		assert.NoError(t, err)
		assert.NotNil(t, saved.Security.ReturnSecretsViaAPI)
		assert.True(t, *saved.Security.ReturnSecretsViaAPI)
	})

	// A bool written only when true would make "off" a write of nothing, and an
	// absent key means "leave the loaded value alone" - so turning the flag off
	// would silently leave it on.
	t.Run("turning a flag off is written, not omitted", func(t *testing.T) {
		SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "app-secret"})
		assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: true}))

		assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: false}))
		assert.False(t, Current().Security.ReturnSecretsViaAPI)

		saved, err := loadManagedSettings(Current().AppPath)
		assert.NoError(t, err)
		assert.NotNil(t, saved.Security.ReturnSecretsViaAPI, "off must be recorded, not left absent")
		assert.False(t, *saved.Security.ReturnSecretsViaAPI)
	})

	// The file is rewritten whole. Writing one setting from a fresh struct would
	// erase the app secret, and nothing encrypted with it would open again.
	t.Run("keeps the app secret", func(t *testing.T) {
		SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "must-survive"})
		assert.NoError(t, SaveAppSecret("must-survive"))

		assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: true}))

		saved, err := loadManagedSettings(Current().AppPath)
		assert.NoError(t, err)
		assert.Equal(t, "must-survive", saved.Secret)
	})

	t.Run("refuses when the config is not loaded", func(t *testing.T) {
		SetCurrent(nil)
		assert.ErrorIs(t, SaveSecuritySettings(&Security{}), ErrSecuritySettingsUnavailable)
	})

	t.Run("refuses nil settings", func(t *testing.T) {
		SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir()})
		assert.ErrorIs(t, SaveSecuritySettings(nil), ErrSecuritySettingsUnavailable)
	})
}

// The same clobber in the other direction: rotating the app secret must not drop
// the security settings written earlier.
func TestSaveAppSecretKeepsSecuritySettings(t *testing.T) {
	SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "old-secret"})
	assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: true}))

	assert.NoError(t, SaveAppSecret("new-secret"))

	saved, err := loadManagedSettings(Current().AppPath)
	assert.NoError(t, err)
	assert.Equal(t, "new-secret", saved.Secret)
	assert.NotNil(t, saved.Security.ReturnSecretsViaAPI, "the security settings must survive a rotation")
	assert.True(t, *saved.Security.ReturnSecretsViaAPI)
}

func TestManagedSecurityApplyTo(t *testing.T) {
	t.Run("an explicit false overrides a loaded true", func(t *testing.T) {
		config := &Config{Security: Security{ReturnSecretsViaAPI: true}}
		(&ManagedSettings{Security: ManagedSecurity{ReturnSecretsViaAPI: new(false)}}).applyTo(config)
		assert.False(t, config.Security.ReturnSecretsViaAPI)
	})

	t.Run("an unset flag leaves the loaded value alone", func(t *testing.T) {
		config := &Config{Security: Security{ReturnSecretsViaAPI: true}}
		(&ManagedSettings{}).applyTo(config)
		assert.True(t, config.Security.ReturnSecretsViaAPI)
	})
}

// The end to end path the reveal gate depends on: what the app writes for itself
// is what the next start reads, over anything the environment still says.
func TestLoadConfigAppliesManagedSecurityOverEnv(t *testing.T) {
	resetLoadState()
	appPath := writeManagedSettings(t,
		"secret = \"an-app-secret\"\n\n[security]\n  return_secrets_via_api = false\n", 0o600)
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\n\n[security]\n  return_secrets_via_api = true\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_SECURITY_RETURN_SECRETS_VIA_API", "true")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.False(t, cfg.Security.ReturnSecretsViaAPI,
		"the managed file is the operator's word and must win over the env and the base config")
}

// The regression the accessor exists for. Under -race this fails immediately if
// the config is mutated in place, or handed out through a plain variable that a
// reload writes while others read it.
func TestCurrentIsSafeUnderConcurrentWrites(t *testing.T) {
	SetCurrent(&Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "start-secret"})

	stop := make(chan struct{})
	var readers sync.WaitGroup
	for range 8 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					cfg := Current()
					_, _ = cfg.Security.ReturnSecretsViaAPI, cfg.Secret
					_ = CurrentSystemInfo().NextStep
				}
			}
		}()
	}

	for i := range 50 {
		assert.NoError(t, SaveSecuritySettings(&Security{ReturnSecretsViaAPI: i%2 == 0}))
		assert.NoError(t, SaveAppSecret(fmt.Sprintf("secret-%d", i)))
		SetInstallationStep(base.InstallationStep(fmt.Sprintf("step-%d", i)))
	}

	close(stop)
	readers.Wait()
}
