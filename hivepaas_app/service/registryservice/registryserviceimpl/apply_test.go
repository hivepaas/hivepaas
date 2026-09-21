package registryserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func planSettings() *entity.RegistrySettings {
	return &entity.RegistrySettings{
		Enabled: true, Type: base.RegistryTypeZot, Managed: true,
		Domain:      "registry.example.com",
		MemoryLimit: 512 * unit.MB,
		Storage: entity.RegistryStorage{
			Type: base.RegistryStorageTypeVolume, Volume: entity.ObjectID{ID: "vol-1"},
		},
		Cleanup: entity.RegistryCleanup{Enabled: true, KeepLast: 10, KeepDays: 30},
	}
}

func TestPlanCarriesTheVolumeAndTheConfiguration(t *testing.T) {
	got, err := planAppDoc(planSettings(), planInput{
		VolumeID: "01M32ATCXAVPK6JYV0HNCFCJYM",
		Htpasswd: "hivepaas:$2y$10$hash\n",
	})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	assert.Equal(t, base.HivepaasRegistryKey, got.Key)
	assert.Equal(t, "registry.example.com", got.Domain)
	assert.Equal(t, "01M32ATCXAVPK6JYV0HNCFCJYM", got.VolumeID)
	assert.Equal(t, "512mb", got.MemoryLimit)
	assert.Contains(t, got.ZotConfig, "docker2s2")
	assert.Contains(t, got.Htpasswd, "$2y$10$hash")
}

// S3 mounts nothing: zot rebuilds its local cache from the bucket, and a mount
// would pin the app to one node for no reason at all.
func TestPlanOnS3HasNoVolume(t *testing.T) {
	cfg := planSettings()
	cfg.Storage = entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}

	got, err := planAppDoc(cfg, planInput{S3: &zotS3Input{
		Bucket: "hp-registry", Region: "us-east-1", Endpoint: "s3.example.com",
		AccessKey: "AK", SecretKey: "SK", Secure: true,
	}})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	assert.Empty(t, got.VolumeID)
	assert.Contains(t, got.ZotConfig, `"name": "s3"`)
	assert.True(t, strings.Contains(got.ZotConfig, `"dedupe": false`))
}

// A volume that was left in the settings from before the operator switched to S3
// must not come back as a mount.
func TestPlanIgnoresAVolumeLeftOverFromBefore(t *testing.T) {
	cfg := planSettings()
	cfg.Storage.Type = base.RegistryStorageTypeS3
	cfg.Storage.CloudStorage = entity.ObjectID{ID: "cs-1"}

	got, err := planAppDoc(cfg, planInput{
		VolumeID: "01M32ATCXAVPK6JYV0HNCFCJYM",
		S3:       &zotS3Input{Bucket: "hp-registry"},
	})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	assert.Empty(t, got.VolumeID)
}

// The document the plan produces has to be one BuildApp will accept, or the
// failure lands halfway through provisioning instead of here.
func TestPlanProducesABuildableDocument(t *testing.T) {
	got, err := planAppDoc(planSettings(), planInput{VolumeID: "01M32ATCXAVPK6JYV0HNCFCJYM", Htpasswd: "x:y\n"})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	_, err = renderAppDoc(got)
	assert.NoError(t, err)
}
