package registryserviceimpl

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func decodeConfig(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("the rendered configuration is not JSON: %v", err)
	}
	return out
}

func volumeSettings() *entity.RegistrySettings {
	return &entity.RegistrySettings{
		Enabled: true, Type: base.RegistryTypeZot, Managed: true,
		Domain:  "registry.example.com",
		Storage: entity.RegistryStorage{Type: base.RegistryStorageTypeVolume},
		Cleanup: entity.RegistryCleanup{
			Enabled: true, Mode: base.RegistryCleanupModePolicy, KeepLast: 10, KeepDays: 30,
		},
	}
}

// Every daemon on the classic image store pushes docker v2s2 manifests. Without
// this key zot answers 415 after accepting every blob, which looks like a broken
// build rather than a missing option.
func TestConfigAlwaysEnablesDockerCompat(t *testing.T) {
	raw, err := renderZotConfig(volumeSettings(), zotConfigInput{})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	cfg := decodeConfig(t, raw)
	httpBlock, _ := cfg["http"].(map[string]any)
	assert.Equal(t, []any{"docker2s2"}, httpBlock["compat"])
	assert.Equal(t, "5000", httpBlock["port"])
}

func TestConfigOnAVolumeDedupes(t *testing.T) {
	raw, err := renderZotConfig(volumeSettings(), zotConfigInput{})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	storage, _ := decodeConfig(t, raw)["storage"].(map[string]any)
	assert.Equal(t, true, storage["dedupe"])
	assert.Equal(t, "/var/lib/registry", storage["rootDirectory"])
	assert.Equal(t, "2h", storage["gcDelay"])
	assert.Equal(t, "1h", storage["gcInterval"])
	assert.Nil(t, storage["storageDriver"])
}

// zot refuses to start with dedupe on and a remote store unless a remote cache is
// configured, which HivePaaS does not run. Rendering it that way would produce a
// registry that never comes up.
func TestConfigOnS3DoesNotDedupe(t *testing.T) {
	cfg := volumeSettings()
	cfg.Storage = entity.RegistryStorage{Type: base.RegistryStorageTypeS3}

	raw, err := renderZotConfig(cfg, zotConfigInput{S3: &zotS3Input{
		Bucket: "hp-registry", Region: "us-east-1",
		Endpoint: "s3.example.com", AccessKey: "AK", SecretKey: "SK", Secure: true,
	}})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	storage, _ := decodeConfig(t, raw)["storage"].(map[string]any)
	assert.Equal(t, false, storage["dedupe"])

	driver, _ := storage["storageDriver"].(map[string]any)
	assert.Equal(t, "s3", driver["name"])
	assert.Equal(t, "hp-registry", driver["bucket"])
	assert.Equal(t, "s3.example.com", driver["regionendpoint"])
	assert.Equal(t, "SK", driver["secretkey"])
	assert.Equal(t, true, driver["secure"])
	assert.Equal(t, true, driver["forcepathstyle"])
}

// S3 without a resolved bucket is a registry that starts and fails on its first
// request, so it is refused where it can still be explained.
func TestConfigOnS3NeedsABucket(t *testing.T) {
	cfg := volumeSettings()
	cfg.Storage = entity.RegistryStorage{Type: base.RegistryStorageTypeS3}

	_, err := renderZotConfig(cfg, zotConfigInput{})
	assert.Error(t, err)
}

func TestConfigCleanupRules(t *testing.T) {
	cfg := volumeSettings()
	cfg.Cleanup.KeepLast = 5
	cfg.Cleanup.KeepDays = 7

	raw, err := renderZotConfig(cfg, zotConfigInput{})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	storage, _ := decodeConfig(t, raw)["storage"].(map[string]any)
	retention, _ := storage["retention"].(map[string]any)
	policies, _ := retention["policies"].([]any)
	first, _ := policies[0].(map[string]any)
	assert.Equal(t, true, first["deleteUntagged"])

	rules, _ := first["keepTags"].([]any)
	assert.Len(t, rules, 3)
	byCount, _ := rules[0].(map[string]any)
	assert.Equal(t, float64(5), byCount["mostRecentlyPushedCount"])
	byPush, _ := rules[1].(map[string]any)
	assert.Equal(t, "168h", byPush["pushedWithin"])
	byPull, _ := rules[2].(map[string]any)
	assert.Equal(t, "168h", byPull["pulledWithin"])
}

// Cleanup off means zot prunes nothing at all. The garbage collector stays on, so
// that a manifest somebody deletes by hand still frees its bytes.
func TestConfigWithoutCleanupHasNoRetention(t *testing.T) {
	cfg := volumeSettings()
	cfg.Cleanup.Enabled = false

	raw, err := renderZotConfig(cfg, zotConfigInput{})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	storage, _ := decodeConfig(t, raw)["storage"].(map[string]any)
	assert.Nil(t, storage["retention"])
	assert.Equal(t, true, storage["gc"])
}
