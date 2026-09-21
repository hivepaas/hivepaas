package registrydto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryuc/registrydto"
)

func TestToEntityCarriesEveryField(t *testing.T) {
	req := &registrydto.UpdateSettingsBaseReq{
		Enabled:     true,
		Domain:      "registry.example.com",
		MemoryLimit: 512 * unit.MB,
		Storage: registrydto.StorageReq{
			Type: base.RegistryStorageTypeVolume, Volume: &registrydto.ObjectIDReq{ID: "vol-1"},
		},
		Cleanup: registrydto.CleanupReq{Enabled: true, KeepLast: 10, KeepDays: 30},
	}

	got := req.ToEntity()

	assert.True(t, got.Enabled)
	assert.Equal(t, base.RegistryTypeZot, got.Type)
	assert.True(t, got.Managed)
	assert.Equal(t, "registry.example.com", got.Domain)
	assert.Equal(t, 512*unit.MB, got.MemoryLimit)
	assert.Equal(t, "vol-1", got.Storage.Volume.ID)
	assert.Equal(t, base.RegistryCleanupModePolicy, got.Cleanup.Mode)
	assert.Equal(t, 10, got.Cleanup.KeepLast)
}

// Every other settings DTO in this repo learned this the hard way: a nil request
// reaching ToEntity is a 500 on a save that should have been a validation error.
func TestToEntityOnNilIsNil(t *testing.T) {
	var req *registrydto.UpdateSettingsBaseReq
	assert.Nil(t, req.ToEntity())
}

// The ids the server wrote are not the client's to send back: the usecase carries
// them over from what is stored.
func TestToEntityIgnoresServerOwnedIDs(t *testing.T) {
	req := &registrydto.UpdateSettingsBaseReq{
		Enabled: true, Domain: "registry.example.com", MemoryLimit: 512 * unit.MB,
		Storage: registrydto.StorageReq{
			Type: base.RegistryStorageTypeVolume, Volume: &registrydto.ObjectIDReq{ID: "vol-1"},
		},
		Cleanup: registrydto.CleanupReq{Enabled: true, KeepLast: 10, KeepDays: 30},
	}

	got := req.ToEntity()

	assert.Empty(t, got.AppID)
	assert.Empty(t, got.RegistryAuthID)
	assert.True(t, got.CredentialRotatedAt.IsZero())
}

// A storage the operator has not chosen yet leaves no reference behind.
func TestToEntityWithoutAStorageChoice(t *testing.T) {
	req := &registrydto.UpdateSettingsBaseReq{
		Storage: registrydto.StorageReq{Type: base.RegistryStorageTypeVolume},
	}

	got := req.ToEntity()

	assert.Empty(t, got.Storage.Volume.ID)
	assert.Empty(t, got.Storage.CloudStorage.ID)
}
