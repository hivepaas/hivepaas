package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func validSettings() *entity.RegistrySettings {
	return &entity.RegistrySettings{
		Enabled: true, Type: base.RegistryTypeZot, Managed: true,
		Domain: "registry.example.com",
		Storage: entity.RegistryStorage{
			Type: base.RegistryStorageTypeVolume, Volume: entity.ObjectID{ID: "vol-1"},
		},
		Cleanup: entity.RegistryCleanup{Enabled: true, KeepLast: 10, KeepDays: 30},
	}
}

func TestValidateAcceptsAWorkingConfiguration(t *testing.T) {
	assert.NoError(t, validateSettings(validSettings(), nil, false))
}

// Nothing below runs when it is off, so nothing below has to be answerable yet:
// an operator filling the form in over two sittings is not an error.
func TestValidateIgnoresEverythingWhenDisabled(t *testing.T) {
	cfg := validSettings()
	cfg.Enabled = false
	cfg.Domain = ""
	cfg.Storage.Volume = entity.ObjectID{}

	assert.NoError(t, validateSettings(cfg, nil, false))
}

func TestValidateRefusals(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*entity.RegistrySettings)
	}{
		{name: "no domain", mutate: func(c *entity.RegistrySettings) { c.Domain = "" }},
		{name: "no volume", mutate: func(c *entity.RegistrySettings) { c.Storage.Volume = entity.ObjectID{} }},
		{name: "no bucket", mutate: func(c *entity.RegistrySettings) {
			c.Storage = entity.RegistryStorage{Type: base.RegistryStorageTypeS3}
		}},
		{name: "unknown storage", mutate: func(c *entity.RegistrySettings) {
			c.Storage = entity.RegistryStorage{Type: base.RegistryStorageType("tape")}
		}},
		{name: "keeping nothing", mutate: func(c *entity.RegistrySettings) { c.Cleanup.KeepLast = 0 }},
		{name: "keeping no days", mutate: func(c *entity.RegistrySettings) { c.Cleanup.KeepDays = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSettings()
			tt.mutate(cfg)

			assert.Error(t, validateSettings(cfg, nil, false))
		})
	}
}

// Images do not move from a volume into a bucket, and a registry that silently
// forgot everything it held is worse than a refusal.
func TestValidateRefusesChangingStorageAfterProvisioning(t *testing.T) {
	current := validSettings()
	next := validSettings()
	next.Storage = entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}

	assert.Error(t, validateSettings(next, current, true))
}

// The same check must not fire before there is anything to lose.
func TestValidateAllowsChangingStorageBeforeProvisioning(t *testing.T) {
	current := validSettings()
	next := validSettings()
	next.Storage = entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}

	assert.NoError(t, validateSettings(next, current, false))
}

// The app id stored in the setting outlives the app when somebody deletes it in
// the app screen, which is how a registry is removed. Trusting that id made a
// deleted registry impossible to provision again, whatever storage was chosen.
func TestValidateIgnoresAnAppIDLeftByADeletedApp(t *testing.T) {
	current := validSettings()
	current.AppID = "deleted-app"
	current.Storage = entity.RegistryStorage{Type: base.RegistryStorageTypeVolume}

	next := validSettings()
	next.Storage.Volume = entity.ObjectID{ID: "vol-2"}

	assert.NoError(t, validateSettings(next, current, false))
}

// Cleanup off means the two numbers are not asked about at all.
func TestValidateIgnoresTheNumbersWhenCleanupIsOff(t *testing.T) {
	cfg := validSettings()
	cfg.Cleanup = entity.RegistryCleanup{Enabled: false}

	assert.NoError(t, validateSettings(cfg, nil, false))
}
