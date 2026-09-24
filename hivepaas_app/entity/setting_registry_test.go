package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// storedRegistry returns a setting shaped like a database row: Data filled in, no
// parsed cache. Reading back through the same *Setting that SetData was called on
// would return the cached struct without touching Data, which proves nothing.
func storedRegistry(t *testing.T, data *entity.RegistrySettings) *entity.Setting {
	t.Helper()

	s := &entity.Setting{ID: "s1", Type: base.SettingTypeRegistry}
	if err := s.SetData(data); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

func TestRegistrySurvivesPersistence(t *testing.T) {
	stored := storedRegistry(t, &entity.RegistrySettings{
		Enabled: true,
		Type:    base.RegistryTypeZot,
		Managed: true,
		Domain:  "registry.example.com",
		Storage: entity.RegistryStorage{
			Type:   base.RegistryStorageTypeVolume,
			Volume: entity.ObjectID{ID: "vol-1"},
		},
		Cleanup: entity.RegistryCleanup{
			Enabled:  true,
			Mode:     base.RegistryCleanupModePolicy,
			KeepLast: 10,
			KeepDays: 30,
		},
		AppID:          "app-1",
		RegistryAuthID: "auth-1",
	})

	got, err := stored.AsRegistrySettings()
	if err != nil {
		t.Fatalf("AsRegistrySettings: %v", err)
	}

	assert.True(t, got.Enabled)
	assert.Equal(t, base.RegistryTypeZot, got.Type)
	assert.Equal(t, "registry.example.com", got.Domain)
	assert.Equal(t, base.RegistryStorageTypeVolume, got.Storage.Type)
	assert.Equal(t, "vol-1", got.Storage.Volume.ID)
	assert.Equal(t, 10, got.Cleanup.KeepLast)
	assert.Equal(t, "app-1", got.AppID)
}

// A stored false must come back false. Cleanup.Enabled defaults to true in New, so
// it may not carry omitempty: with it, turning cleanup off would be dropped from
// the JSON and the default would resurrect it on the next read.
func TestRegistryCleanupDisabledSurvives(t *testing.T) {
	stored := storedRegistry(t, &entity.RegistrySettings{
		Cleanup: entity.RegistryCleanup{Enabled: false, KeepLast: 10, KeepDays: 30},
	})

	got, err := stored.AsRegistrySettings()
	if err != nil {
		t.Fatalf("AsRegistrySettings: %v", err)
	}
	assert.False(t, got.Cleanup.Enabled)
}

// Nothing has been saved yet: the defaults are what the dashboard shows.
func TestRegistryDefaults(t *testing.T) {
	s := &entity.Setting{ID: "s1", Type: base.SettingTypeRegistry}
	data, err := s.AsRegistrySettings()
	if err != nil {
		t.Fatalf("AsRegistrySettings: %v", err)
	}

	assert.False(t, data.Enabled)
	assert.True(t, data.Managed)
	assert.Equal(t, base.RegistryTypeZot, data.Type)
	assert.Equal(t, base.RegistryStorageTypeVolume, data.Storage.Type)
	assert.True(t, data.Cleanup.Enabled)
	assert.Equal(t, base.RegistryCleanupModePolicy, data.Cleanup.Mode)
	assert.Equal(t, 10, data.Cleanup.KeepLast)
	assert.Equal(t, 30, data.Cleanup.KeepDays)
}

// The volume and the bucket must be refusable and undeletable while the registry
// holds them, which is what a reference id is for.
func TestRegistryReportsItsReferences(t *testing.T) {
	onVolume := &entity.RegistrySettings{Storage: entity.RegistryStorage{
		Type: base.RegistryStorageTypeVolume, Volume: entity.ObjectID{ID: "vol-1"},
	}}
	assert.Equal(t, []string{"vol-1"}, onVolume.GetRefObjectIDs().RefSettingIDs)

	onS3 := &entity.RegistrySettings{Storage: entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}}
	assert.Equal(t, []string{"cs-1"}, onS3.GetRefObjectIDs().RefSettingIDs)

	// The one that is not in use is not a reference, however it got filled in.
	both := &entity.RegistrySettings{Storage: entity.RegistryStorage{
		Type:         base.RegistryStorageTypeS3,
		Volume:       entity.ObjectID{ID: "vol-1"},
		CloudStorage: entity.ObjectID{ID: "cs-1"},
	}}
	assert.Equal(t, []string{"cs-1"}, both.GetRefObjectIDs().RefSettingIDs)
}

// The ids provisioning wrote name objects of this installation only, and Apply
// creates them again wherever a spec is imported.
func TestRegistrySpecPolicyStripsProvisionedIDs(t *testing.T) {
	data := &entity.RegistrySettings{
		Enabled: true, Domain: "registry.example.com",
		AppID: "app-1", RegistryAuthID: "auth-1",
	}

	policy := entity.SpecPolicyFor(base.SettingTypeRegistry)
	assert.NotNil(t, policy)
	assert.True(t, policy.Decide(&entity.Setting{Type: base.SettingTypeRegistry}).Export)

	policy.Strip(data)
	assert.Empty(t, data.AppID)
	assert.Empty(t, data.RegistryAuthID)
	assert.Equal(t, "registry.example.com", data.Domain, "the configuration itself stays")
}
