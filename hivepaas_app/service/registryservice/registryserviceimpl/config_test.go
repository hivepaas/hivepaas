package registryserviceimpl

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	assert.Len(t, rules, len(retentionBaseTagPrefixes)+3)

	// The last three are the ones over every tag, whatever it starts with.
	tail := rules[len(rules)-3:]
	byCount, _ := tail[0].(map[string]any)
	assert.Equal(t, []any{".*"}, byCount["patterns"])
	assert.Equal(t, float64(5), byCount["mostRecentlyPushedCount"])
	byPush, _ := tail[1].(map[string]any)
	assert.Equal(t, "168h", byPush["pushedWithin"])
	byPull, _ := tail[2].(map[string]any)
	assert.Equal(t, "168h", byPull["pulledWithin"])
}

// With the environment in the tag, an app's environments share a repository. A
// single count would let the environment that deploys most evict the one that
// deploys least, so each prefix is counted on its own.
func TestConfigCountsEachEnvironmentSeparately(t *testing.T) {
	cfg := volumeSettings()
	cfg.Cleanup.KeepLast = 10

	raw, err := renderZotConfig(cfg, zotConfigInput{EnvKeys: []string{"dev", "prod", "canary"}})
	if err != nil {
		t.Fatalf("renderZotConfig: %v", err)
	}

	storage, _ := decodeConfig(t, raw)["storage"].(map[string]any)
	retention, _ := storage["retention"].(map[string]any)
	policies, _ := retention["policies"].([]any)
	first, _ := policies[0].(map[string]any)
	rules, _ := first["keepTags"].([]any)

	counted := map[string]float64{}
	for _, rule := range rules {
		entry, _ := rule.(map[string]any)
		patterns, _ := entry["patterns"].([]any)
		count, ok := entry["mostRecentlyPushedCount"].(float64)
		if !ok || len(patterns) != 1 {
			continue
		}
		pattern, _ := patterns[0].(string)
		counted[pattern] = count
	}

	assert.Equal(t, float64(10), counted["^dev"])
	assert.Equal(t, float64(10), counted["^prod"])
	// Not one of the names written into the list, so it came from the database.
	assert.Equal(t, float64(10), counted["^canary"])
}

// A rule that does not match the tags it is meant to keep deletes them silently,
// so both are built by the same function.
func TestRetentionPrefixesMatchTheTagsTheyKeep(t *testing.T) {
	for _, envKey := range []string{"dev", "prod", "canary", "Staging--EU", "default"} {
		app := &entity.App{ProjectEnv: &entity.ProjectEnv{Key: envKey}}
		tag, err := app.ImageTag("9f3c1de0a1b2")
		if err != nil {
			t.Fatalf("ImageTag(%q): %v", envKey, err)
		}

		matched := false
		for _, prefix := range retentionTagPrefixes([]string{envKey}) {
			if regexp.MustCompile("^" + regexp.QuoteMeta(prefix)).MatchString(tag) {
				matched = true
				break
			}
		}
		assert.True(t, matched, "no retention rule matches the tag %q", tag)
	}
}

// "production" is already kept by the rule for "prod": a second rule would keep
// the same tags, and the list is short so that it stays readable.
func TestRetentionPrefixesSkipCoveredEnvironments(t *testing.T) {
	prefixes := retentionTagPrefixes([]string{"production", "dev", "staging", "sandbox"})

	assert.NotContains(t, prefixes, "production")
	assert.NotContains(t, prefixes, "staging")
	assert.Contains(t, prefixes, "sandbox")
	assert.Equal(t, len(retentionBaseTagPrefixes)+1, len(prefixes))
}

func TestRetentionPrefixesAreCapped(t *testing.T) {
	envKeys := make([]string, 0, 64)
	for i := range 64 {
		envKeys = append(envKeys, fmt.Sprintf("zone%02d", i))
	}

	prefixes := retentionTagPrefixes(envKeys)
	assert.Len(t, prefixes, maxRetentionTagPrefixes)
	assert.Subset(t, prefixes, retentionBaseTagPrefixes)
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
