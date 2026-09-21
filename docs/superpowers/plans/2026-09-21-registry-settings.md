# Registry Settings Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A global setting that provisions a self-hosted zot registry as an app in the hidden
`hivepaas` project, creates the credential HivePaaS pushes with, and lets zot prune itself.

**Architecture:** One new singleton setting (`registry`) drives a new `registryservice`. Its
`Apply` renders zot's configuration from the setting, renders an app document, and hands both
to `appprovisionservice.ProvisionApp` - the same call the template usecase makes - so the
registry is an ordinary app with routing, SSL, logs and deploys. Cleanup is zot's own
retention plus its online garbage collection; HivePaaS runs no job.

**Tech Stack:** Go 1.27, gin, fx, bun, testify, docker swarm, zot v2.1.21; React + TypeScript,
zod, react-hook-form on the dashboard.

**Spec:** [docs/superpowers/specs/2026-09-21-registry-settings-design.md](../specs/2026-09-21-registry-settings-design.md)

## Global Constraints

- Before calling any task done: `go build ./...`, `golangci-lint run ./...` over the **whole**
  repo (120-character lines, US spelling), `go test ./...`. Run `make gen-swag` in any task
  that changes a DTO; `docs/openapi/swagger.json` is generated and committed.
- The zot image is pinned in the backend: `ghcr.io/project-zot/zot:v2.1.21`.
- The rendered zot configuration always carries `"http": {"compat": ["docker2s2"]}`. Without
  it a docker daemon on the classic image store gets 415 on every manifest.
- `storage.dedupe` is `true` only for volume storage. With S3 it must be `false`: zot refuses
  to start with `dedupe: true` and a remote store unless a remote cache is configured.
- `storage.gcDelay: "2h"`, `storage.gcInterval: "1h"`.
- Defaults: `KeepLast` 10, `KeepDays` 30, `MemoryLimit` 512MB, minimum memory 256MB.
- The app: project key `hivepaas`, environment `default`, app key `registry`.
- The registry account's username is fixed: `hivepaas`.
- A wire-format change belongs in the same piece of work as its dashboard change.
- End every commit message with `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`.

---

## File Structure

**Backend - new**

| File | Responsibility |
|---|---|
| `hivepaas_app/base/registry.go` | The enums: registry type, storage type, cleanup mode |
| `hivepaas_app/entity/setting_registry.go` | `RegistrySettings` and its parser |
| `hivepaas_app/entity/setting_registry_migration.go` | The version bump hook |
| `hivepaas_app/service/registryservice/service.go` | The interface and its request/response types |
| `hivepaas_app/service/registryservice/registryserviceimpl/service.go` | Constructor and dependencies |
| `.../registryserviceimpl/validate.go` | What a saved configuration has to satisfy |
| `.../registryserviceimpl/config.go` | The zot configuration, rendered from the setting |
| `.../registryserviceimpl/appdoc.go` + `app.yaml.tmpl` | The app document handed to `BuildApp` |
| `.../registryserviceimpl/credential.go` | Password, htpasswd and the managed registry auth |
| `.../registryserviceimpl/apply.go` | Provision on the first save, reconcile on the rest |
| `.../registryserviceimpl/status.go` | What is running, and what the registry holds |
| `.../registryserviceimpl/probe.go` | Proxy detection for the domain |
| `.../registryserviceimpl/push_check.go` | The large-upload check through the public domain |
| `hivepaas_app/usecase/systemsettings/registryuc/*` | `uc.go`, `settings_get.go`, `settings_update.go`, `credential_rotate.go`, `domain_probe.go`, `push_check.go` |
| `.../registryuc/registrydto/*` | `settings_get.go`, `settings_update.go`, `actions.go` |
| `hivepaas_app/interface/api/handler/systemsettingshandler/sys_registry.go` | The five endpoints |

**Backend - modified**

| File | Change |
|---|---|
| `hivepaas_app/base/setting.go` | `SettingTypeRegistry` |
| `hivepaas_app/base/permission.go` | `ResourceTypeRegistry` |
| `hivepaas_app/base/hivepaas.go` | `HivepaasRegistryKey` |
| `hivepaas_app/service/specservice/specmodel/singleton.go` | `SettingTypeRegistry: "registry"` |
| `hivepaas_app/interface/api/handler/basesettinghandler/handler.go` | `RegistryUC` field |
| `hivepaas_app/interface/api/server/router_system.go` | The `/registry` group |
| `hivepaas_app/registry/provides.go` | `registryuc.New`, `registryserviceimpl.New` |

**Dashboard - new**, mirroring the logging feature file for file:
`domain/hivepaas-registry-settings.entity.ts`,
`api/services/hivepaas-registry-settings-services/{*.api.ts,*.api.contracts.ts,*.api.validator.ts,index.ts}`,
`api/hooks/use-hivepaas-registry-settings.api.ts`,
`data/queries/hivepaas-registry-settings.queries.ts`,
`data/commands/hivepaas-registry-settings.commands.ts`,
`layouts/registry/layout/registry-layout.com.tsx`,
`routes/registry/{route,form,schemas,building-blocks}`.

**Dashboard - modified:** `shared/constants/route.constants.ts`,
`shared/layouts/module/module-sidebar/module-sidebar.com.tsx`,
`modules/system-settings/system-settings.router.tsx`, and the three barrel `index.ts` files.

---

### Task 1: The setting type and its entity

**Files:**
- Create: `hivepaas_app/base/registry.go`
- Create: `hivepaas_app/entity/setting_registry.go`
- Create: `hivepaas_app/entity/setting_registry_migration.go`
- Create: `hivepaas_app/entity/setting_registry_test.go`
- Modify: `hivepaas_app/base/setting.go`, `hivepaas_app/base/permission.go`,
  `hivepaas_app/base/hivepaas.go`,
  `hivepaas_app/service/specservice/specmodel/singleton.go`

**Interfaces:**
- Produces: `base.SettingTypeRegistry`, `base.ResourceTypeRegistry`, `base.HivepaasRegistryKey`,
  `base.RegistryType`/`RegistryTypeZot`, `base.RegistryStorageType`/`RegistryStorageTypeVolume`/
  `RegistryStorageTypeS3`, `base.RegistryCleanupMode`/`RegistryCleanupModePolicy`,
  `entity.RegistrySettings` with `Storage entity.RegistryStorage` and
  `Cleanup entity.RegistryCleanup`, `entity.CurrentRegistrySettingsVersion`,
  `(*entity.Setting).AsRegistrySettings()` and `MustAsRegistrySettings()`.
- Note for the reviewer: the spec writes `VolumeID string` and `CloudStorageID string`. This
  plan stores them as `entity.ObjectID`, which is how every other setting references another
  one (`LoggingVictoriaLogs.Volume`), and what `GetRefObjectIDs` and the DTO helpers expect.

- [ ] **Step 1: Write the failing test**

`hivepaas_app/entity/setting_registry_test.go`:

```go
package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
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
		Enabled:     true,
		Type:        base.RegistryTypeZot,
		Managed:     true,
		Domain:      "registry.example.com",
		MemoryLimit: 512 * unit.MB,
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
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/entity/ -run TestRegistry -v`
Expected: FAIL - `undefined: entity.RegistrySettings`.

- [ ] **Step 3: Add the base constants**

`hivepaas_app/base/registry.go`:

```go
package base

// RegistryType names the registry software, independently of who runs it.
type RegistryType string

const (
	RegistryTypeZot RegistryType = "zot"
)

// RegistryStorageType is where the images are kept. It is decided when the
// registry is provisioned and refused afterwards: nothing copies images from one
// to the other.
type RegistryStorageType string

const (
	RegistryStorageTypeVolume RegistryStorageType = "volume"
	RegistryStorageTypeS3     RegistryStorageType = "s3"
)

// RegistryCleanupMode says how the set of images to keep is decided.
//
// "policy" is the only value today: two numbers, turned into zot's own retention
// rules. A second value would be one where HivePaaS computes the set from what is
// running and tells the registry exactly which tags to keep.
type RegistryCleanupMode string

const (
	RegistryCleanupModePolicy RegistryCleanupMode = "policy"
)
```

In `hivepaas_app/base/setting.go`, beside `SettingTypeRegistryAuth`:

```go
	SettingTypeRegistry          SettingType = "registry"
```

In `hivepaas_app/base/permission.go`, beside `ResourceTypeRegistryAuth`:

```go
	ResourceTypeRegistry          ResourceType = "registry"
```

In `hivepaas_app/base/hivepaas.go`, after the logging block:

```go
	// The registry is created by the app when an operator switches it on, not by
	// the stack file, so it carries no stack prefix either. Its absence means the
	// registry was never switched on, not that something is broken.
	HivepaasRegistryKey = "registry"
```

- [ ] **Step 4: Write the entity**

`hivepaas_app/entity/setting_registry.go`:

```go
package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentRegistrySettingsVersion = 1

	// DefaultRegistryKeepLast and DefaultRegistryKeepDays are the cleanup an
	// operator gets without saying anything: the last ten builds of every app,
	// and everything from the past month.
	DefaultRegistryKeepLast = 10
	DefaultRegistryKeepDays = 30

	DefaultRegistryMemoryLimit = 512 * unit.MB
	MinRegistryMemoryLimit     = 256 * unit.MB
)

var _ = registerSettingParser(base.SettingTypeRegistry, &registrySettingsParser{})

type registrySettingsParser struct {
}

// New is both the never-configured default and the struct stored data is
// unmarshalled into, so every field defaulted here to a non-zero value must be
// written unconditionally - with `omitempty`, a stored false is left out of the
// JSON and the default below survives the unmarshal as a silent true.
func (s *registrySettingsParser) New() SettingData {
	return &RegistrySettings{
		Type:    base.RegistryTypeZot,
		Managed: true,
		Storage: RegistryStorage{Type: base.RegistryStorageTypeVolume},
		Cleanup: RegistryCleanup{
			Enabled:  true,
			Mode:     base.RegistryCleanupModePolicy,
			KeepLast: DefaultRegistryKeepLast,
			KeepDays: DefaultRegistryKeepDays,
		},
		MemoryLimit: DefaultRegistryMemoryLimit,
	}
}

// RegistrySettings is the whole subsystem's configuration. It is global, and until
// Enabled is set nothing is provisioned.
type RegistrySettings struct {
	// Enabled is what provisions the registry. Clearing it removes nothing: the
	// images are on a volume or in a bucket, and deleting them is done in the
	// app's own screen.
	Enabled bool `json:"enabled,omitempty"`

	// Type and Managed are separate on purpose, the way logging's are: an
	// operator pointing HivePaaS at a registry it does not run is a case worth
	// keeping expressible, even though Managed is the only value today.
	Type base.RegistryType `json:"type,omitempty"`
	// Managed defaults to true in New, so it carries no omitempty.
	Managed bool `json:"managed"`

	// Domain is the address images are named with, and it is required to enable
	// the registry: a docker daemon speaks to a registry over HTTPS at a name.
	Domain string `json:"domain,omitempty"`

	Storage RegistryStorage `json:"storage"`
	Cleanup RegistryCleanup `json:"cleanup"`

	MemoryLimit unit.DataSize `json:"memoryLimit,omitempty"`

	// AppID and RegistryAuthID are what provisioning created. They are how Apply
	// finds its own work again, and how the dashboard links to it.
	AppID          string `json:"appId,omitempty"`
	RegistryAuthID string `json:"registryAuthId,omitempty"`
}

type RegistryStorage struct {
	Type base.RegistryStorageType `json:"type,omitempty"`
	// Volume is a cluster-volume setting, for Type volume. The volume decides
	// which node the registry runs on: it already carries that answer, and the
	// placement constraint is derived from it.
	Volume ObjectID `json:"volume,omitzero"`
	// CloudStorage is a cloud-storage setting, for Type s3. Bucket, region and
	// credentials come from it; nothing about S3 is duplicated here.
	CloudStorage ObjectID `json:"cloudStorage,omitzero"`
}

// InUse names the setting this storage actually reads, which is not the same as
// the fields that happen to be filled in: switching between them in the form
// leaves the other one behind.
func (s RegistryStorage) InUse() ObjectID {
	if s.Type == base.RegistryStorageTypeS3 {
		return s.CloudStorage
	}
	return s.Volume
}

type RegistryCleanup struct {
	// Enabled defaults to true in New, so it carries no omitempty: a registry
	// that never prunes is the problem this feature exists to solve.
	Enabled bool                     `json:"enabled"`
	Mode    base.RegistryCleanupMode `json:"mode,omitempty"`
	// KeepLast keeps this many of the newest tags of every repository.
	KeepLast int `json:"keepLast,omitempty"`
	// KeepDays keeps every tag pushed, or pulled, within this many days.
	KeepDays int `json:"keepDays,omitempty"`
}

func (s *RegistrySettings) GetType() base.SettingType {
	return base.SettingTypeRegistry
}

// GetRefObjectIDs reports the volume or the bucket the images are kept in, which
// is what keeps either from being deleted while the registry still uses it.
func (s *RegistrySettings) GetRefObjectIDs() *RefObjectIDs {
	ids := &RefObjectIDs{}
	if inUse := s.Storage.InUse(); inUse.ID != "" {
		ids.RefSettingIDs = append(ids.RefSettingIDs, inUse.ID)
	}
	return ids
}

func (s *RegistrySettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsRegistrySettings() (*RegistrySettings, error) {
	return parseSettingAs[*RegistrySettings](s)
}

func (s *Setting) MustAsRegistrySettings() *RegistrySettings {
	return gofn.Must(s.AsRegistrySettings())
}
```

`hivepaas_app/entity/setting_registry_migration.go`:

```go
package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *RegistrySettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRegistrySettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentRegistrySettingsVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so there is nothing to migrate from yet.

	setting.Version = CurrentRegistrySettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
```

- [ ] **Step 5: Classify the type in the spec model**

Without this, `TestEverySettingTypeIsClassifiedExactlyOnce` fails. In
`hivepaas_app/service/specservice/specmodel/singleton.go`, in the second group of
`singletonBlockNames` (the one holding `logging` and `systemCleanup`):

```go
	base.SettingTypeRegistry:          "registry",
```

- [ ] **Step 6: Register a spec policy**

A setting type also needs a `SpecPolicy`, or `TestEverySettingTypeHasASpecPolicy` fails - the
exporter refuses a type nobody decided about rather than leaking it. The registry's policy
exports the configuration and strips the two ids provisioning wrote, because Apply creates
both again wherever the spec lands. In `hivepaas_app/entity/setting_spec.go`, beside
`loggingSpecPolicy`:

```go
type registrySpecPolicy struct{}

func (registrySpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }

func (registrySpecPolicy) Strip(data SettingData) {
	if registry, ok := data.(*RegistrySettings); ok {
		registry.AppID = ""
		registry.RegistryAuthID = ""
	}
}
```

and register it: `_ = registerSpecPolicy(base.SettingTypeRegistry, registrySpecPolicy{})`.

- [ ] **Step 7: Run the tests**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/service/specservice/... -v`
Expected: PASS, including `TestEverySettingTypeIsClassifiedExactlyOnce` and
`TestEverySettingTypeHasASpecPolicy`.

- [ ] **Step 8: Build, lint and commit**

```bash
go build ./... && golangci-lint run ./... && go test ./hivepaas_app/entity/ ./hivepaas_app/service/specservice/...
git add hivepaas_app/base hivepaas_app/entity hivepaas_app/service/specservice/specmodel/singleton.go
git commit -m "$(cat <<'EOF'
feat(registry): add the registry setting type and its entity

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: The zot configuration

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/config.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/config_test.go`

**Interfaces:**
- Consumes: `entity.RegistrySettings` (Task 1).
- Produces: `zotConfigInput{S3 *zotS3Input}`, `zotS3Input{Bucket, Region, Endpoint, AccessKey,
  SecretKey string, Secure bool}`, and
  `renderZotConfig(cfg *entity.RegistrySettings, in zotConfigInput) ([]byte, error)` returning
  indented JSON.

- [ ] **Step 1: Write the failing test**

`config_test.go`:

```go
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
	http, _ := cfg["http"].(map[string]any)
	assert.Equal(t, []any{"docker2s2"}, http["compat"])
	assert.Equal(t, "5000", http["port"])
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
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestConfig -v`
Expected: FAIL - the package does not exist yet.

- [ ] **Step 3: Write the renderer**

`config.go`:

```go
// Package registryserviceimpl runs the system registry: it renders zot's
// configuration and the app document from the stored setting, provisions the app
// through the ordinary provisioning path, and keeps the credential HivePaaS
// pushes with.
package registryserviceimpl

import (
	"encoding/json"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// registryImage is pinned here rather than in the setting: the configuration
	// this package writes is the configuration this version of zot accepts, and
	// the two have to move together.
	registryImage = "ghcr.io/project-zot/zot:v2.1.21"

	registryPort     = "5000"
	registryRootDir  = "/var/lib/registry"
	registryConfDir  = "/etc/zot"
	registryS3Prefix = "/zot"

	// gcDelay shields a blob younger than itself from collection, so it has to
	// be longer than the slowest push. gcInterval is how often a pass starts; a
	// repository waits up to about twice that.
	registryGCDelay    = "2h"
	registryGCInterval = "1h"

	hoursPerDay = 24
)

// zotConfigInput is what the configuration needs that the setting does not hold.
type zotConfigInput struct {
	// S3 is nil for a registry on a volume.
	S3 *zotS3Input
}

type zotS3Input struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
}

// renderZotConfig writes the configuration file zot is started with.
//
// It is built as maps rather than a text template because two of its keys are
// conditional and one of them is a list whose length depends on the settings;
// a template with that many branches is a template nobody can read.
func renderZotConfig(cfg *entity.RegistrySettings, in zotConfigInput) ([]byte, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}

	storage := map[string]any{
		"rootDirectory": registryRootDir,
		// zot refuses to start with dedupe on and a remote store unless a remote
		// cache is configured, which HivePaaS does not run.
		"dedupe":     cfg.Storage.Type != base.RegistryStorageTypeS3,
		"gc":         true,
		"gcDelay":    registryGCDelay,
		"gcInterval": registryGCInterval,
	}
	if cfg.Cleanup.Enabled {
		storage["retention"] = retentionPolicy(cfg.Cleanup)
	}
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		if in.S3 == nil {
			return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
				WithExtraDetail("the registry is set to S3 but no bucket was resolved")
		}
		storage["storageDriver"] = map[string]any{
			"name":           "s3",
			"rootdirectory":  registryS3Prefix,
			"bucket":         in.S3.Bucket,
			"region":         in.S3.Region,
			"regionendpoint": in.S3.Endpoint,
			"secure":         in.S3.Secure,
			"forcepathstyle": true,
			"accesskey":      in.S3.AccessKey,
			"secretkey":      in.S3.SecretKey,
		}
	}

	doc := map[string]any{
		"distSpecVersion": "1.1.1",
		"storage":         storage,
		"http": map[string]any{
			"address": "0.0.0.0",
			"port":    registryPort,
			// Every daemon on the classic image store pushes docker v2s2
			// manifests, which zot answers 415 to without this.
			"compat": []string{"docker2s2"},
			"auth":   map[string]any{"htpasswd": map[string]any{"path": registryConfDir + "/htpasswd"}},
		},
		"log": map[string]any{"level": "info"},
		"extensions": map[string]any{
			// search answers with each repository's size and last update, which
			// is what the dashboard's status section reads.
			"search": map[string]any{"enable": true},
			"ui":     map[string]any{"enable": true},
		},
	}

	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return raw, nil
}

// retentionPolicy turns the two numbers into zot's rules.
//
// The rules are a union: a tag survives if any of them keeps it. pulledWithin is
// there for the image a long-running service fetched when it was last
// rescheduled, which may be older than anything pushedWithin would save.
func retentionPolicy(cleanup entity.RegistryCleanup) map[string]any {
	window := fmt.Sprintf("%dh", cleanup.KeepDays*hoursPerDay)
	return map[string]any{
		"dryRun": false,
		"policies": []any{map[string]any{
			"repositories":    []string{"**"},
			"deleteUntagged":  true,
			"deleteReferrers": true,
			"keepTags": []any{
				map[string]any{"patterns": []string{".*"}, "mostRecentlyPushedCount": cleanup.KeepLast},
				map[string]any{"patterns": []string{".*"}, "pushedWithin": window},
				map[string]any{"patterns": []string{".*"}, "pulledWithin": window},
			},
		}},
	}
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestConfig -v`
Expected: PASS, all five.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/registryservice
git commit -m "$(cat <<'EOF'
feat(registry): render zot's configuration from the setting

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: The app document

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/app.yaml.tmpl`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/appdoc.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/appdoc_test.go`

**Interfaces:**
- Consumes: `renderZotConfig` (Task 2), `specmodel.AppDoc`, `specmodel.CheckBuildable`.
- Produces: `appDocInput{Name, Key, Domain, MemoryLimit, VolumeName, ZotConfig, Htpasswd
  string}` and `renderAppDoc(in appDocInput) (*specmodel.AppDoc, error)`.

- [ ] **Step 1: Write the failing test**

`appdoc_test.go`:

```go
package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func docInput() appDocInput {
	return appDocInput{
		Name:        "Registry",
		Key:         "registry",
		Domain:      "registry.example.com",
		MemoryLimit: "512mb",
		VolumeName:  "registry-data",
		ZotConfig:   "{\n  \"distSpecVersion\": \"1.1.1\"\n}",
		Htpasswd:    "hivepaas:$2y$05$abc",
	}
}

func TestAppDocIsBuildable(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	// CheckBuildable is what BuildApp runs first; a document that fails it would
	// fail at provisioning time instead, with the app half created.
	assert.NoError(t, specmodel.CheckBuildable(doc))
	assert.Equal(t, "registry", doc.App)
	assert.Equal(t, registryImage, doc.Deployment.Source.ImageSource.Image)
}

func TestAppDocMountsTheVolume(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	mnt, ok := doc.Deployment.Storage.Mounts[registryRootDir]
	assert.True(t, ok, "the images have to be on the volume")
	assert.Equal(t, "registry-data", mnt.Source)
}

// With S3 there is nothing local worth keeping: zot rebuilds its cache from the
// bucket, and a mount would pin the app to a node for no reason.
func TestAppDocWithoutAVolumeHasNoMount(t *testing.T) {
	in := docInput()
	in.VolumeName = ""

	doc, err := renderAppDoc(in)
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	assert.Empty(t, doc.Deployment.Storage.Mounts)
}

func TestAppDocCarriesTheConfigAndTheAccount(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	settings, _ := doc.Settings["configFiles"].(map[string]any)
	assert.Contains(t, settings, "config.json")

	secrets, _ := doc.Settings["secrets"].(map[string]any)
	assert.Contains(t, secrets, "ZOT_HTPASSWD")

	routing, _ := doc.Settings["routing"].(map[string]any)
	assert.Equal(t, 5000, routing["port"])
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestAppDoc -v`
Expected: FAIL - `undefined: renderAppDoc`.

- [ ] **Step 3: Write the document template**

`app.yaml.tmpl` - the same shape a template's `app:` block has, because that is what an
`AppDoc` is:

```yaml
app: {{ .Key }}
name: {{ .Name }}
deployment:
  source:
    activeMethod: image
    imageSource: {image: {{ .Image }}}
{{- if .VolumeName }}
  storage:
    mounts:
      {{ .RootDir }}: {type: volume, source: {{ .VolumeName }}}
{{- end }}
  container:
    healthcheck:
      # One static binary, no shell: there is nothing in the image to run a check with.
      enabled: false
  resources:
    limits: {memory: {{ .MemoryLimit }}}
settings:
  kind:
    category: webapp
    engine: zot
    webapp: {}
  secrets:
    ZOT_HTPASSWD:
      value: |
{{ .Htpasswd | indent 8 }}
      swarmRef:
        file: {name: {{ .ConfDir }}/htpasswd, mode: 444}
  configFiles:
    config.json:
      content: |
{{ .ZotConfig | indent 8 }}
      swarmRef:
        file: {name: {{ .ConfDir }}/config.json, mode: 444}
  routing:
    port: 5000
    exposePublicly: true
    domains:
      - {domain: {{ .Domain }}, enabled: true, forceHttps: true}
```

- [ ] **Step 4: Write the renderer**

`appdoc.go`:

```go
package registryserviceimpl

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

//go:embed app.yaml.tmpl
var appDocTemplate string

// appDocInput is everything the document needs that is decided elsewhere.
type appDocInput struct {
	Name        string
	Key         string
	Domain      string
	MemoryLimit string
	// VolumeName is empty for a registry on S3, which mounts nothing.
	VolumeName string
	ZotConfig  string
	Htpasswd   string
}

// renderAppDoc produces the document BuildApp turns into the app's settings.
//
// It goes through YAML rather than building specmodel structs directly because
// AppDoc.Settings is an untyped tree - the settings blocks are decoded from it -
// so the typed route would mean hand-writing the same maps with none of the
// readability. This is also the exact path a template takes, which means the
// registry cannot drift away from what templates can express.
func renderAppDoc(in appDocInput) (*specmodel.AppDoc, error) {
	tmpl, err := template.New("registry-app").Funcs(template.FuncMap{
		"indent": indentBlock,
	}).Parse(appDocTemplate)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		appDocInput
		Image   string
		RootDir string
		ConfDir string
	}{appDocInput: in, Image: registryImage, RootDir: registryRootDir, ConfDir: registryConfDir})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	doc := &specmodel.AppDoc{}
	decoder := yaml.NewDecoder(bytes.NewReader(buf.Bytes()))
	decoder.KnownFields(true)
	if err = decoder.Decode(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return doc, nil
}

// indentBlock puts a multi-line value under a YAML block scalar. Every line is
// indented, including the last, and a trailing newline is dropped so that the
// block does not end in a blank line the decoder would keep.
func indentBlock(spaces int, value string) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}
```

Note the argument order: `{{ .Htpasswd | indent 8 }}` calls `indentBlock(8, htpasswd)`, which
is why `spaces` comes first.

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestAppDoc -v`
Expected: PASS, all four.

- [ ] **Step 6: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/registryservice
git commit -m "$(cat <<'EOF'
feat(registry): render the registry app's spec document

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: The service, its dependencies and what it refuses

**Files:**
- Create: `hivepaas_app/service/registryservice/service.go`
- Create: `hivepaas_app/service/registryservice/types.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/service.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/validate.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/validate_test.go`
- Modify: `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: Task 1's entity.
- Produces: `registryservice.Service` with `Apply`, `Status`, `RotateCredential`, `ProbeDomain`,
  `CheckPush`; `registryservice.SettingApplyReq{Setting *entity.Setting}`,
  `SettingApplyResp{App *entity.App}`, `Status`, `RotateCredentialResp`, `DomainProbe`,
  `PushCheckResult`; `registryserviceimpl.New(...) registryservice.Service`.
- Naming note: the action the spec calls "Test a large push" is `CheckPush` in Go, and its file
  is `push_check.go`. A file called `test_push.go` would be fine but `push_test.go` would be
  compiled as a test file, and the near-miss is not worth the risk.

- [ ] **Step 1: Write the failing test**

`validate_test.go`:

```go
package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

func validSettings() *entity.RegistrySettings {
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

func TestValidateAcceptsAWorkingConfiguration(t *testing.T) {
	assert.NoError(t, validateSettings(validSettings(), nil))
}

// Nothing below runs when it is off, so nothing below has to be answerable yet:
// an operator filling the form in over two sittings is not an error.
func TestValidateIgnoresEverythingWhenDisabled(t *testing.T) {
	cfg := validSettings()
	cfg.Enabled = false
	cfg.Domain = ""
	cfg.Storage.Volume = entity.ObjectID{}

	assert.NoError(t, validateSettings(cfg, nil))
}

func TestValidateRefusals(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*entity.RegistrySettings)
		detail string
	}{
		{
			name:   "no domain",
			mutate: func(c *entity.RegistrySettings) { c.Domain = "" },
			detail: "domain",
		},
		{
			name:   "no volume",
			mutate: func(c *entity.RegistrySettings) { c.Storage.Volume = entity.ObjectID{} },
			detail: "volume",
		},
		{
			name: "no bucket",
			mutate: func(c *entity.RegistrySettings) {
				c.Storage = entity.RegistryStorage{Type: base.RegistryStorageTypeS3}
			},
			detail: "cloud storage",
		},
		{
			name:   "keeping nothing",
			mutate: func(c *entity.RegistrySettings) { c.Cleanup.KeepLast = 0 },
			detail: "keepLast",
		},
		{
			name:   "keeping no days",
			mutate: func(c *entity.RegistrySettings) { c.Cleanup.KeepDays = 0 },
			detail: "keepDays",
		},
		{
			name:   "memory below what zot needs",
			mutate: func(c *entity.RegistrySettings) { c.MemoryLimit = 64 * unit.MB },
			detail: "memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSettings()
			tt.mutate(cfg)

			err := validateSettings(cfg, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.detail)
		})
	}
}

// Images do not move from a volume into a bucket, and a registry that silently
// forgot everything it held is worse than a refusal.
func TestValidateRefusesChangingStorageAfterProvisioning(t *testing.T) {
	current := validSettings()
	current.AppID = "app-1"

	next := validSettings()
	next.AppID = "app-1"
	next.Storage = entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}

	err := validateSettings(next, current)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage")
}

// The same check must not fire before there is anything to lose.
func TestValidateAllowsChangingStorageBeforeProvisioning(t *testing.T) {
	current := validSettings()
	next := validSettings()
	next.Storage = entity.RegistryStorage{
		Type: base.RegistryStorageTypeS3, CloudStorage: entity.ObjectID{ID: "cs-1"},
	}

	assert.NoError(t, validateSettings(next, current))
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestValidate -v`
Expected: FAIL - `undefined: validateSettings`.

- [ ] **Step 3: Name the errors**

This repo names its errors per feature rather than reaching for a generic one, so the
messages and the HTTP statuses come out right. Add to
`hivepaas_app/hperrors/errors_registry.go`, which already holds the errors for reading an
image registry:

```go
	ErrRegistryNotConfigured    = NewErr(ErrPreconditionFailed, "ERR_REGISTRY_NOT_CONFIGURED")
	ErrRegistrySettingsInvalid  = NewErr(ErrValueInvalid, "ERR_REGISTRY_SETTINGS_INVALID")
	ErrRegistryStorageImmutable = NewErr(ErrValueInvalid, "ERR_REGISTRY_STORAGE_IMMUTABLE")
	ErrRegistryUnreachable      = NewErr(ErrUnavailable, "ERR_REGISTRY_UNREACHABLE")
```

`ErrValueInvalid` and `ErrUnavailable` both sit under `ErrPreconditionFailed`, which is what
gives a refused save a 4xx instead of a 500.

- [ ] **Step 4: Write the interface and its types**

`service.go`:

```go
// Package registryservice runs the registry HivePaaS provisions for itself: the
// app, the credential HivePaaS pushes with, and the questions the dashboard asks
// about both.
package registryservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Apply makes the cluster match the stored configuration: it provisions the
	// registry on the first save and reconciles it on the rest. It is
	// idempotent, so a failed save leaves work the next save retries.
	Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)

	// Status says what is running and what the registry holds, for the settings
	// screen. It never fails the screen: a registry that cannot be reached comes
	// back as a status saying so.
	Status(ctx context.Context, db database.IDB, setting *entity.Setting) (*Status, error)

	// RotateCredential issues a new password, keeping the old one working until
	// the grace period ends, because a service that is not redeployed still
	// presents the credential it was deployed with.
	RotateCredential(ctx context.Context, db database.IDB, req *RotateCredentialReq) (
		*RotateCredentialResp, error)

	// ProbeDomain reports what can be seen from outside about the registry's
	// address - a proxy in front of it, above all. It is advisory.
	ProbeDomain(ctx context.Context, domain string) (*DomainProbe, error)

	// CheckPush uploads a large blob through the public domain and throws it
	// away, which is the only way to find a body limit before a build does.
	CheckPush(ctx context.Context, db database.IDB, req *PushCheckReq) (*PushCheckResult, error)

	// Validate refuses a configuration before it is written. The usecase calls it
	// in PrepareUpdate, so a bad save is a validation error rather than a stored
	// configuration Apply then fails on.
	Validate(next, current *entity.RegistrySettings) error
}
```

`types.go`:

```go
package registryservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type SettingApplyReq struct {
	// Setting is the row that was just saved. Apply loads it itself when nil.
	Setting *entity.Setting
	// TriggerUserID is who asked, recorded on the first deployment the way the
	// template usecase records it. Empty when nothing asked - a reconcile that
	// runs on start-up, for instance.
	TriggerUserID string
}

type SettingApplyResp struct {
	// App is the registry's app, set whenever the registry is enabled.
	App *entity.App
}

type Status struct {
	// Provisioned says the app exists. Everything below is only meaningful then.
	Provisioned bool
	AppID       string
	// Reachable says the registry answered its own API. Unreachable is normal
	// while it restarts after a save.
	Reachable bool
	// Unreachable carries why, for the screen to show.
	Unreachable string
	// Repositories and StoredBytes come from zot's search extension, and are
	// zero when it could not be asked.
	Repositories int
	StoredBytes  int64
	// CredentialRotatedAt is when the password last changed, empty if never.
	CredentialRotatedAt time.Time
}

type RotateCredentialReq struct {
	Setting *entity.Setting
}

type RotateCredentialResp struct {
	// GraceEndsAt is when the previous password stops working.
	GraceEndsAt time.Time
}

type DomainProbe struct {
	// Proxied says a proxy answered instead of the registry.
	Proxied bool
	// Evidence names what said so: a header, or addresses that are not the
	// cluster's. It is what the dashboard shows; it is never a refusal.
	Evidence []string
	// Reached says the probe got an answer at all.
	Reached bool
}

type PushCheckReq struct {
	Setting *entity.Setting
	// Bytes is how large a blob to send. Zero means the default of 150 MB, which
	// is above the limit a proxy on a free plan imposes.
	Bytes int64
}

type PushCheckResult struct {
	OK bool
	// StatusCode is what the registry, or whatever is in front of it, answered.
	StatusCode int
	// Detail is the sentence the dashboard shows.
	Detail string
	// Elapsed is how long the upload took, which is the other thing an operator
	// wants to know about a proxy.
	Elapsed time.Duration
}
```

- [ ] **Step 5: Write the service struct and the validation**

`registryserviceimpl/service.go`:

```go
package registryserviceimpl

import (
	"net/http"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
)

const httpTimeout = 30 * time.Second

type service struct {
	appRepo         repository.AppRepo
	projectRepo     repository.ProjectRepo
	settingRepo     repository.SettingRepo
	provisionSvc    appprovisionservice.Service
	specService     specservice.Service
	logger          logging.Logger

	// httpClient talks to the registry itself - status, the domain probe and the
	// push check. A field so tests can hand it a transport pointed at a stub.
	httpClient *http.Client
}

// New builds the registry service. fx wires the arguments from the provider list
// in registry/provides.go; adding a parameter here needs no other change.
func New(
	appRepo repository.AppRepo,
	projectRepo repository.ProjectRepo,
	settingRepo repository.SettingRepo,
	provisionSvc appprovisionservice.Service,
	specService specservice.Service,
	logger logging.Logger,
) registryservice.Service {
	return &service{
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		settingRepo:  settingRepo,
		provisionSvc: provisionSvc,
		specService:  specService,
		logger:       logger,
		httpClient:   &http.Client{Timeout: httpTimeout},
	}
}
```

`registryserviceimpl/validate.go`:

```go
package registryserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// validateSettings refuses a configuration that could not produce a working
// registry, and one that would throw away a working one.
//
// current is what is stored today, nil when nothing is. It is only needed for the
// questions that are about a change rather than about a value.
func validateSettings(cfg, current *entity.RegistrySettings) error {
	if cfg == nil {
		return hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}
	// Nothing below is deployed while it is off, so nothing below has to be
	// answerable yet.
	if !cfg.Enabled {
		return nil
	}

	if cfg.Domain == "" {
		return invalid("a domain is required: images are named after it, and docker only " +
			"speaks to a registry over HTTPS at a name")
	}
	switch cfg.Storage.Type {
	case base.RegistryStorageTypeVolume:
		if cfg.Storage.Volume.ID == "" {
			return invalid("choose the volume the images are kept on")
		}
	case base.RegistryStorageTypeS3:
		if cfg.Storage.CloudStorage.ID == "" {
			return invalid("choose the cloud storage the images are kept in")
		}
	default:
		return invalid("unknown storage type %q", cfg.Storage.Type)
	}
	if cfg.Cleanup.Enabled {
		if cfg.Cleanup.KeepLast < 1 {
			return invalid("keepLast must keep at least one build of every app")
		}
		if cfg.Cleanup.KeepDays < 1 {
			return invalid("keepDays must be at least one day")
		}
	}
	if cfg.MemoryLimit < entity.MinRegistryMemoryLimit {
		return invalid("the memory limit must be at least 256MB")
	}

	// Once the app exists the storage is what holds its images. Nothing copies
	// them anywhere, so changing it is refused rather than obeyed.
	if current != nil && current.AppID != "" {
		if cfg.Storage.Type != current.Storage.Type ||
			cfg.Storage.InUse().ID != current.Storage.InUse().ID {
			return invalid("the registry's storage cannot be changed once it holds images: " +
				"provision a new registry instead")
		}
	}
	return nil
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).WithExtraDetail(format, args...)
}
```

- [ ] **Step 6: Run the tests**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestValidate -v`
Expected: PASS, including every subtest of `TestValidateRefusals`.

- [ ] **Step 7: Wire it into fx**

In `hivepaas_app/registry/provides.go`, beside `loggingserviceimpl.New`:

```go
	registryserviceimpl.New,
```

Run: `go build ./... && go test ./hivepaas_app/registry/...`
Expected: PASS - `wiring_test.go` builds the graph, so a missing dependency fails here.

- [ ] **Step 8: Commit**

```bash
golangci-lint run ./...
git add hivepaas_app/service/registryservice hivepaas_app/registry/provides.go
git commit -m "$(cat <<'EOF'
feat(registry): add the registry service and its validation

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: The credential

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/credential.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/credential_test.go`

**Interfaces:**
- Consumes: Task 1's entity, `entity.RegistryAuth`, `entity.NewEncryptedField`.
- Produces: `registryUsername` constant (`"hivepaas"`), `generatePassword() (string, error)`,
  `htpasswdLine(user, password string) (string, error)`,
  `htpasswdContent(lines ...string) string`,
  `newRegistryAuthSetting(id, domain, password string, timeNow time.Time) (*entity.Setting, error)`,
  `updateRegistryAuthSetting(setting *entity.Setting, domain, password string, timeNow time.Time) error`.

**A note on where the tests stop.** This package has no mock repositories - the service tests
in `loggingserviceimpl` test pure functions and nothing else - so the rule here is the same:
every decision lives in a function that takes values and returns values, and the orchestration
that touches the database is kept thin enough to read. Task 6 says how it is verified.

- [ ] **Step 1: Write the failing test**

`credential_test.go`:

```go
package registryserviceimpl

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestGeneratedPasswordsAreLongAndDifferent(t *testing.T) {
	first, err := generatePassword()
	if err != nil {
		t.Fatalf("generatePassword: %v", err)
	}
	second, err := generatePassword()
	if err != nil {
		t.Fatalf("generatePassword: %v", err)
	}

	assert.GreaterOrEqual(t, len(first), 32)
	assert.NotEqual(t, first, second)
}

// zot checks passwords against bcrypt hashes. A line it cannot parse is a
// registry nobody can log in to, and nothing else would say so.
func TestHtpasswdLineIsBcryptAndVerifies(t *testing.T) {
	line, err := htpasswdLine(registryUsername, "s3cret-password")
	if err != nil {
		t.Fatalf("htpasswdLine: %v", err)
	}

	user, hash, found := strings.Cut(line, ":")
	assert.True(t, found)
	assert.Equal(t, "hivepaas", user)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("s3cret-password")))
}

// Rotation keeps the old line beside the new one, so a service that has not been
// redeployed can still pull. Both have to survive into the file.
func TestHtpasswdContentKeepsEveryLine(t *testing.T) {
	content := htpasswdContent("hivepaas:$2y$10$new", "hivepaas:$2y$10$old")

	assert.Equal(t, "hivepaas:$2y$10$new\nhivepaas:$2y$10$old\n", content)
	assert.Equal(t, "hivepaas:$2y$10$only\n", htpasswdContent("hivepaas:$2y$10$only", ""))
}

func TestNewRegistryAuthSettingIsGlobalAndReadable(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)

	setting, err := newRegistryAuthSetting("auth-1", "registry.example.com", "pw", now)
	if err != nil {
		t.Fatalf("newRegistryAuthSetting: %v", err)
	}

	assert.Equal(t, base.SettingTypeRegistryAuth, setting.Type)
	assert.Equal(t, base.ObjectScopeGlobal, setting.Scope)
	assert.True(t, setting.Inheritable, "every project has to be able to push to it")

	auth := setting.MustAsRegistryAuth()
	assert.Equal(t, "registry.example.com", auth.Address)
	assert.Equal(t, "hivepaas", auth.Username)
	assert.False(t, auth.Readonly, "HivePaaS pushes with this credential")

	plain, err := auth.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "pw", plain)
}

// A domain change has to reach the credential, or builds keep pushing to the old
// address with no sign of why.
func TestUpdateRegistryAuthSettingRewritesTheAddress(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	setting, err := newRegistryAuthSetting("auth-1", "old.example.com", "pw", now)
	if err != nil {
		t.Fatalf("newRegistryAuthSetting: %v", err)
	}

	later := now.Add(time.Hour)
	if err = updateRegistryAuthSetting(setting, "new.example.com", "pw2", later); err != nil {
		t.Fatalf("updateRegistryAuthSetting: %v", err)
	}

	auth := setting.MustAsRegistryAuth()
	assert.Equal(t, "new.example.com", auth.Address)
	plain, _ := auth.Password.GetPlain()
	assert.Equal(t, "pw2", plain)
	assert.Equal(t, later, setting.UpdatedAt)
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run 'TestGenerated|TestHtpasswd|TestNewRegistryAuth|TestUpdateRegistryAuth' -v`
Expected: FAIL - `undefined: generatePassword`.

- [ ] **Step 3: Write the credential helpers**

`credential.go`:

```go
package registryserviceimpl

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// registryUsername is fixed. It is the namespace every image HivePaaS pushes
	// lands in - <domain>/hivepaas/<app key> - so it is not something to change
	// later without renaming every image in the registry.
	registryUsername = "hivepaas"

	// passwordBytes is what is drawn from the random source; base64 makes it 32
	// characters, which is more than anything checking it will ever need.
	passwordBytes = 24
)

// generatePassword returns the registry account's password. It is generated
// rather than asked for: nobody types it, the dashboard shows it as a link to the
// registry auth setting, and every consumer reads it from there.
func generatePassword() (string, error) {
	buf := make([]byte, passwordBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", hperrors.Wrap(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// htpasswdLine writes the line zot checks a password against. HivePaaS hashes it
// itself, so nothing asks an operator to run htpasswd by hand.
func htpasswdLine(user, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return fmt.Sprintf("%s:%s", user, hash), nil
}

// htpasswdContent joins the lines into the file. Empty lines are dropped, which
// is what makes "the previous password, if there is one" expressible without the
// caller branching.
func htpasswdContent(lines ...string) string {
	var out strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

// newRegistryAuthSetting builds the credential every push and pull uses.
//
// It is an ordinary registry-auth setting, global and inheritable, because that
// is what makes it appear in an app's deployment settings beside Docker Hub and
// anything else the operator added: nothing downstream needs to know that this
// one was created by HivePaaS.
func newRegistryAuthSetting(id, domain, password string, timeNow time.Time) (*entity.Setting, error) {
	setting := &entity.Setting{
		ID:          id,
		Scope:       base.ObjectScopeGlobal,
		Type:        base.SettingTypeRegistryAuth,
		Name:        "System registry",
		Status:      base.SettingStatusActive,
		Inheritable: true,
		Version:     entity.CurrentRegistryAuthVersion,
		UpdateVer:   1,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	err := setting.SetData(&entity.RegistryAuth{
		Username: registryUsername,
		Password: entity.NewEncryptedField(password),
		Address:  domain,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

// updateRegistryAuthSetting rewrites the address, and the password when one is
// given. An empty password keeps the stored one, which is what every save that is
// not a rotation does.
func updateRegistryAuthSetting(setting *entity.Setting, domain, password string, timeNow time.Time) error {
	auth, err := setting.AsRegistryAuth()
	if err != nil {
		return hperrors.Wrap(err)
	}

	auth.Address = domain
	auth.Username = registryUsername
	if password != "" {
		auth.Password = entity.NewEncryptedField(password)
	}
	if err = setting.SetData(auth); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	return nil
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/service/registryservice/... -run 'TestGenerated|TestHtpasswd|TestNewRegistryAuth|TestUpdateRegistryAuth' -v`
Expected: PASS, all five.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/registryservice
git commit -m "$(cat <<'EOF'
feat(registry): create the managed credential the registry is pushed to with

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Apply - provision, then reconcile

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/apply.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/apply_test.go`
- Modify: `hivepaas_app/service/registryservice/registryserviceimpl/service.go` (two more
  dependencies)

**Interfaces:**
- Consumes: `renderZotConfig`, `zotConfigInput`, `zotS3Input` (Task 2); `renderAppDoc`,
  `appDocInput` (Task 3); `validateSettings` (Task 4); the credential helpers (Task 5);
  `appprovisionservice.ProvisionApp`, `specservice.BuildApp`,
  `clustersecretservice.UpdateConfigForApp` and `UpdateSecretForApp`.
- Produces: `(*service).Apply`, and the pure planner
  `planAppDoc(cfg *entity.RegistrySettings, in planInput) (appDocInput, error)` with
  `planInput{VolumeName string, S3 *zotS3Input, Htpasswd string}`.

- [ ] **Step 1: Write the failing test**

`apply_test.go` - the planner is where the decisions are, so that is what is tested:

```go
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
		VolumeName: "registry-data",
		Htpasswd:   "hivepaas:$2y$10$hash\n",
	})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	assert.Equal(t, base.HivepaasRegistryKey, got.Key)
	assert.Equal(t, "registry.example.com", got.Domain)
	assert.Equal(t, "registry-data", got.VolumeName)
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

	assert.Empty(t, got.VolumeName)
	assert.Contains(t, got.ZotConfig, "\"name\": \"s3\"")
	assert.True(t, strings.Contains(got.ZotConfig, "\"dedupe\": false"))
}

// The document the plan produces has to be one BuildApp will accept, or the
// failure lands halfway through provisioning instead of here.
func TestPlanProducesABuildableDocument(t *testing.T) {
	got, err := planAppDoc(planSettings(), planInput{VolumeName: "registry-data", Htpasswd: "x:y\n"})
	if err != nil {
		t.Fatalf("planAppDoc: %v", err)
	}

	_, err = renderAppDoc(got)
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestPlan -v`
Expected: FAIL - `undefined: planAppDoc`.

- [ ] **Step 3: Add the two dependencies the orchestration needs**

In `registryserviceimpl/service.go`, add to the struct and to `New`, in both cases in the same
order as the existing fields:

```go
	clusterSecretService clustersecretservice.Service
	volumeService        volumeservice.Service
```

`clusterSecretService` is how a changed configuration reaches a running app - it replaces the
swarm config and updates the service. `volumeService` resolves a cluster-volume setting to the
name a mount uses.

- [ ] **Step 4: Write Apply and its planner**

`apply.go`:

```go
package registryserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
)

// planInput is what the plan needs that the setting only points at: the volume's
// name rather than its id, the bucket's credentials rather than its id, and the
// htpasswd file, which is derived from a password nothing else may see.
type planInput struct {
	VolumeName string
	S3         *zotS3Input
	Htpasswd   string
}

// planAppDoc turns the setting into the document the app is built from. It is
// separate from Apply, and pure, because this is where every decision is: what
// gets mounted, what the configuration says, how big the app is allowed to be.
func planAppDoc(cfg *entity.RegistrySettings, in planInput) (appDocInput, error) {
	zotConfig, err := renderZotConfig(cfg, zotConfigInput{S3: in.S3})
	if err != nil {
		return appDocInput{}, hperrors.Wrap(err)
	}

	volumeName := in.VolumeName
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		volumeName = ""
	}

	return appDocInput{
		Name:        "Registry",
		Key:         base.HivepaasRegistryKey,
		Domain:      cfg.Domain,
		// DataSize.String already writes "512mb", which is the spelling every
		// size in a spec document uses.
		MemoryLimit: cfg.MemoryLimit.String(),
		VolumeName:  volumeName,
		ZotConfig:   string(zotConfig),
		Htpasswd:    in.Htpasswd,
	}, nil
}

// Apply makes the cluster match the stored configuration.
//
// It runs after every save and does nothing that is already done, so a save that
// fails half way leaves work the next save retries rather than a broken registry.
func (s *service) Apply(
	ctx context.Context,
	db database.IDB,
	req *registryservice.SettingApplyReq,
) (_ *registryservice.SettingApplyResp, err error) {
	setting := req.Setting
	if setting == nil {
		setting, err = s.loadSetting(ctx, db)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if setting == nil {
			// Never configured, which is the default: there is nothing to run.
			return &registryservice.SettingApplyResp{}, nil
		}
	}

	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled {
		// Switching it off removes nothing: the images are on a volume or in a
		// bucket, and deleting them is done in the app's own screen.
		return &registryservice.SettingApplyResp{}, nil
	}

	credential, password, err := s.ensureCredential(ctx, db, setting, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	line, err := htpasswdLine(registryUsername, password)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	in, err := s.resolveStorage(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	in.Htpasswd = htpasswdContent(line, "")

	plan, err := planAppDoc(cfg, in)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	app, err := s.loadApp(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app == nil {
		app, err = s.provision(ctx, db, plan, req.TriggerUserID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	} else if err = s.reconcile(ctx, db, app, plan); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The ids are how the next Apply finds this work again.
	cfg.AppID, cfg.RegistryAuthID = app.ID, credential.ID
	if err = setting.SetData(cfg); err != nil {
		return nil, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	err = s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registryservice.SettingApplyResp{App: app}, nil
}

// provision creates the app the first time, in the hidden hivepaas project, the
// same way the template usecase creates one: the document becomes the app's
// settings, and the settings become the service.
func (s *service) provision(
	ctx context.Context,
	db database.IDB,
	plan appDocInput,
	triggerUserID string,
) (*entity.App, error) {
	doc, err := renderAppDoc(plan)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	project, projectEnv, err := s.rootProjectEnv(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	resp, err := s.provisionSvc.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
		ProjectID:    project.ID,
		ProjectEnvID: projectEnv.ID,
		AppID:        gofn.Must(ulid.NewStringULID()),
		Name:         plan.Name,
		Status:       base.AppStatusActive,
		Note:         "The registry HivePaaS runs for itself. It is configured in system settings.",
		Configure: func(ctx context.Context, db database.IDB, app *entity.App,
			spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
			built, err := s.specService.BuildApp(ctx, db, &specservice.BuildAppReq{
				App: app, Doc: doc, Spec: spec, TimeNow: timeNow,
			})
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			return built.Settings, nil
		},
		Deployment: &appprovisionservice.FirstDeployment{
			Source:   base.DeploymentTriggerSourceAPI,
			SourceID: triggerUserID,
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.App, nil
}

// reconcile brings a registry that already exists in line with a saved change.
//
// Only the configuration file and the account can change without provisioning
// again - the storage is refused by validateSettings, and the domain is routing,
// which the app's own settings carry. Both are swarm objects, which are
// immutable, so each change is a new object and a service update: the registry
// restarts for a few seconds.
func (s *service) reconcile(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	plan appDocInput,
) error {
	if err := s.applyConfigFile(ctx, db, app, plan.ZotConfig); err != nil {
		return hperrors.Wrap(err)
	}
	if err := s.applyHtpasswd(ctx, db, app, plan.Htpasswd); err != nil {
		return hperrors.Wrap(err)
	}
	return s.applyRouting(ctx, db, app, plan.Domain)
}

// loadApp returns the registry's app, or nil when it was never provisioned. It
// looks the app up by key rather than by the stored id, so that an id left over
// from an app somebody deleted does not stop the registry from being created
// again.
func (s *service) loadApp(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	*entity.App, error) {
	apps, _, err := s.appRepo.List(ctx, db, "", nil,
		bunex.SelectJoin("JOIN projects ON projects.id = app.project_id"),
		bunex.SelectWhere("projects.key = ?", base.HivepaasProjectKey),
		bunex.SelectWhere("app.key = ?", base.HivepaasRegistryKey),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, nil
	}
	return apps[0], nil
}
```

The helpers `loadSetting`, `ensureCredential`, `resolveStorage`, `rootProjectEnv`,
`applyConfigFile`, `applyHtpasswd` and `applyRouting` are the thin database-touching part:

- `loadSetting` is `settingRepo.GetSingle` with `entity.NewObjectScopeGlobal()` and
  `base.SettingTypeRegistry`, returning `nil` on `hperrors.ErrNotFound`.
- `ensureCredential` returns the existing registry-auth setting named by
  `cfg.RegistryAuthID` - rewriting its address through `updateRegistryAuthSetting` when the
  domain changed - or creates one with `generatePassword` and `newRegistryAuthSetting`, and
  upserts it. It returns the setting and the plaintext password, because the htpasswd file
  needs the plaintext and nothing else in the flow has it.
- `resolveStorage` reads the cluster-volume setting and returns its name, or reads the
  cloud-storage setting and fills a `zotS3Input` from it.
- `rootProjectEnv` loads the `hivepaas` project with its `ProjectEnvs` relation and picks the
  one keyed `default`, which is the environment `project_sync.go` gives a system app.
- `applyConfigFile` finds the app's `config-file` setting named `config.json`, returns early
  when the content is identical, and otherwise writes the new content and calls
  `clusterSecretService.UpdateConfigForApp(ctx, db, app, old, new)`.
- `applyHtpasswd` does the same for the `secret` setting named `ZOT_HTPASSWD` through
  `UpdateSecretForApp`.
- `applyRouting` rewrites the app's `app-routing` setting's first domain when the saved domain
  differs, and upserts it.

- [ ] **Step 5: Run the tests and the graph**

Run: `go test ./hivepaas_app/service/registryservice/... -v && go test ./hivepaas_app/registry/...`
Expected: PASS. The wiring test is what catches a dependency added to `New` without a provider.

- [ ] **Step 6: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/registryservice
git commit -m "$(cat <<'EOF'
feat(registry): provision the registry app and reconcile it on later saves

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Status

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/status.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/status_test.go`

**Interfaces:**
- Consumes: Task 4's `registryservice.Status`, Task 6's `loadApp`.
- Produces: `(*service).Status`, and the pure `parseSearchResponse(body []byte) (int, int64, error)`.

zot's search extension answers this GraphQL query, authenticated, which is where the numbers
on the settings screen come from:

```graphql
{ RepoListWithNewestImage(requestedPage: {limit: 100, offset: 0, sortBy: UPDATE_TIME}) {
    Results { Name Size } } }
```

- [ ] **Step 1: Write the failing test**

`status_test.go`:

```go
package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSearchResponseSumsWhatIsStored(t *testing.T) {
	body := []byte(`{"data":{"RepoListWithNewestImage":{"Results":[
		{"Name":"hivepaas/api","Size":"24148826"},
		{"Name":"hivepaas/web","Size":"4139411"}]}}}`)

	repos, bytes, err := parseSearchResponse(body)

	assert.NoError(t, err)
	assert.Equal(t, 2, repos)
	assert.Equal(t, int64(28288237), bytes)
}

// An empty registry is a normal state, not a failure to read one.
func TestParseSearchResponseOnAnEmptyRegistry(t *testing.T) {
	repos, bytes, err := parseSearchResponse([]byte(`{"data":{"RepoListWithNewestImage":{"Results":[]}}}`))

	assert.NoError(t, err)
	assert.Zero(t, repos)
	assert.Zero(t, bytes)
}

// The search extension answers GraphQL errors with 200, so the body is the only
// place a failure shows up.
func TestParseSearchResponseReportsGraphQLErrors(t *testing.T) {
	_, _, err := parseSearchResponse([]byte(`{"errors":[{"message":"Cannot query field \"Name\""}]}`))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot query field")
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestParseSearch -v`
Expected: FAIL - `undefined: parseSearchResponse`.

- [ ] **Step 3: Write the status**

`status.go`:

```go
package registryserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

const searchQuery = `{"query":"{ RepoListWithNewestImage(requestedPage:{limit:100,offset:0,` +
	`sortBy:UPDATE_TIME}) { Results { Name Size } } }"}`

// Status answers the settings screen. It never fails the screen: a registry that
// cannot be reached is a status saying so, because "still restarting after a
// save" is the most common reason and it is not an error.
func (s *service) Status(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
) (*registryservice.Status, error) {
	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	status := &registryservice.Status{}
	app, err := s.loadApp(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app == nil {
		return status, nil
	}
	status.Provisioned, status.AppID = true, app.ID

	repos, stored, err := s.askRegistry(ctx, db, cfg)
	if err != nil {
		status.Unreachable = err.Error()
		return status, nil
	}
	status.Reachable, status.Repositories, status.StoredBytes = true, repos, stored
	return status, nil
}

// askRegistry reads what the registry holds, through its public domain and with
// the managed credential - the same path a build takes, so a status that works
// says more than one read from inside the cluster would.
func (s *service) askRegistry(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	int, int64, error) {
	user, password, err := s.credentialOf(ctx, db, cfg)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}

	url := fmt.Sprintf("https://%s/v2/_zot/ext/search", cfg.Domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(searchQuery)))
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).
			WithExtraDetail("the registry answered %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSearchResponseBytes))
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	return parseSearchResponse(body)
}

// parseSearchResponse reads zot's search answer. Sizes come back as strings
// because they are 64-bit, which JSON numbers in JavaScript are not.
func parseSearchResponse(body []byte) (int, int64, error) {
	var payload struct {
		Data struct {
			RepoList struct {
				Results []struct {
					Name string `json:"Name"`
					Size string `json:"Size"`
				} `json:"Results"`
			} `json:"RepoListWithNewestImage"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	if len(payload.Errors) > 0 {
		return 0, 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).
			WithExtraDetail("%s", payload.Errors[0].Message)
	}

	var total int64
	for _, repo := range payload.Data.RepoList.Results {
		size, err := strconv.ParseInt(repo.Size, 10, 64)
		if err != nil {
			continue // a size nobody can read is not a reason to show nothing
		}
		total += size
	}
	return len(payload.Data.RepoList.Results), total, nil
}
```

Add `maxSearchResponseBytes = 1 << 20` to the constants in `config.go`, and the `io` import.
`credentialOf` loads the registry-auth setting by `cfg.RegistryAuthID` and returns its username
and decrypted password.

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestParseSearch -v`
Expected: PASS, all three.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/registryservice
git commit -m "$(cat <<'EOF'
feat(registry): report what the registry holds

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: The endpoints

**Files:**
- Create: `hivepaas_app/usecase/systemsettings/registryuc/uc.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/settings_get.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/settings_update.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/registrydto/settings_get.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/registrydto/settings_update.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/registrydto/settings_update_test.go`
- Create: `hivepaas_app/interface/api/handler/systemsettingshandler/sys_registry.go`
- Modify: `hivepaas_app/interface/api/handler/basesettinghandler/handler.go`,
  `hivepaas_app/interface/api/server/router_system.go`,
  `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: `registryservice.Service` (Tasks 4-7), `settings.BaseUC`,
  `settings.UpdateUniqueSettingReq`, `settings.GetUniqueSettingReq`.
- Produces: `registryuc.UC` with `GetRegistrySettings` and `UpdateRegistrySettings`;
  `registrydto.GetRegistrySettingsReq/Resp`, `RegistrySettingsResp`,
  `UpdateRegistrySettingsReq/Resp`, `UpdateSettingsBaseReq.ToEntity()`.
- The wire shape, which Task 11 consumes:

```jsonc
{
  "enabled": true,
  "type": "zot",
  "managed": true,
  "domain": "registry.example.com",
  "storage": {"type": "volume", "volume": {"id": "..."}, "cloudStorage": {"id": ""}},
  "cleanup": {"enabled": true, "mode": "policy", "keepLast": 10, "keepDays": 30},
  "memoryLimit": "512mb",
  "app": {"id": "..."},            // response only
  "registryAuth": {"id": "..."},   // response only
  "registryStatus": {              // response only
    "provisioned": true, "appId": "...", "reachable": true, "unreachable": "",
    "repositories": 12, "storedBytes": 8123456789
  }
}
```

- [ ] **Step 1: Write the failing test**

`registrydto/settings_update_test.go`:

```go
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
		MemoryLimit: "512mb",
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
	assert.Equal(t, unit.DataSize(512*unit.MB), got.MemoryLimit)
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

// The ids the server wrote are not the client's to send back.
func TestToEntityIgnoresServerOwnedIDs(t *testing.T) {
	req := &registrydto.UpdateSettingsBaseReq{
		Enabled: true, Domain: "registry.example.com", MemoryLimit: "512mb",
		Storage: registrydto.StorageReq{
			Type: base.RegistryStorageTypeVolume, Volume: &registrydto.ObjectIDReq{ID: "vol-1"},
		},
		Cleanup: registrydto.CleanupReq{Enabled: true, KeepLast: 10, KeepDays: 30},
	}

	got := req.ToEntity()

	assert.Empty(t, got.AppID)
	assert.Empty(t, got.RegistryAuthID)
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/usecase/systemsettings/registryuc/... -v`
Expected: FAIL - the package does not exist.

- [ ] **Step 3: Write the DTOs**

`registrydto/settings_update.go` carries `UpdateRegistrySettingsReq` (embedding
`settings.UpdateUniqueSettingReq` and `*UpdateSettingsBaseReq`), the `StorageReq`,
`CleanupReq` and `ObjectIDReq` types, `NewUpdateRegistrySettingsReq()`, `Validate()` built
from `vld` validators the way `loggingdto` does - domain at most 255 characters, `keepLast`
between 1 and 1000, `keepDays` between 1 and 3650, `memoryLimit` matching
`^$|^\d+(b|kb|mb|gb|tb)$` - and:

```go
func (req *UpdateSettingsBaseReq) ToEntity() *entity.RegistrySettings {
	if req == nil {
		return nil
	}
	return &entity.RegistrySettings{
		Enabled: req.Enabled,
		// Type and Managed are not the client's to choose while zot is the only
		// registry HivePaaS provisions. They are written here so that the stored
		// row is complete, and so that a second type later is a wire change
		// rather than a migration.
		Type:        base.RegistryTypeZot,
		Managed:     true,
		Domain:      strings.TrimSpace(req.Domain),
		Storage:     req.Storage.ToEntity(),
		Cleanup:     req.Cleanup.ToEntity(),
		MemoryLimit: unit.MustParseDataSizeString(req.MemoryLimit),
	}
}
```

`registrydto/settings_get.go` carries `GetRegistrySettingsReq/Resp`, `RegistrySettingsResp`
(embedding `*settings.BaseSettingResp`), `RegistryStatusResp`, and
`TransformRegistrySettings(input *RegistrySettingsTransformationInput) (*RegistrySettingsResp, error)`,
following `loggingdto.TransformLoggingSettings` field for field. The password is never in this
response: the dashboard links to the registry-auth setting instead, which has its own reveal
flow.

- [ ] **Step 4: Write the usecase**

`registryuc/uc.go` mirrors `logginguc/uc.go` with `registryService registryservice.Service`
in place of the logging one. `settings_get.go` mirrors `logginguc/settings_get.go`, calling
`uc.registryService.Status` when a setting exists. `settings_update.go` mirrors
`logginguc/settings_update.go`: `UpdateUniqueSetting` with `Load`, `PrepareUpdate` and
`AfterPersisting`, where

- `PrepareUpdate` runs `validateSettings` through a new service method
  `registryService.Validate(next, current)` - exported on the interface so the usecase can
  refuse before anything is written - and carries `AppID` and `RegistryAuthID` over from the
  stored row, because they are the server's and the client never sends them;
- `AfterPersisting` calls `registryService.Apply` with the persisted setting. Apply is
  idempotent, so a failure there leaves a stored configuration the next save retries.

- [ ] **Step 5: Write the handler and the routes**

`sys_registry.go` mirrors `sys_logging.go` with `base.ResourceTypeRegistry` and
`h.RegistryUC`. Add `RegistryUC *registryuc.UC` to `basesettinghandler.Handler` and its `New`,
`registryuc.New` to `provides.go`, and the group to `router_system.go` beside logging's:

```go
	// Registry settings
	{
		registryGroup := systemSettingGroup.Group("/registry")
		registryGroup.GET("", systemSettingsHandler.GetRegistrySettings)
		registryGroup.PUT("", systemSettingsHandler.UpdateRegistrySettings)
	}
```

- [ ] **Step 6: Run the tests and regenerate the API document**

```bash
go test ./hivepaas_app/usecase/systemsettings/registryuc/... ./hivepaas_app/registry/... -v
make gen-swag
```
Expected: PASS, and `docs/openapi/swagger.json` gains the two operations.

- [ ] **Step 7: Commit**

```bash
go build ./... && golangci-lint run ./... && go test ./...
git add hivepaas_app docs/openapi
git commit -m "$(cat <<'EOF'
feat(registry): add the registry settings endpoints

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: The domain probe and the push check

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/probe.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/probe_test.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/push_check.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/push_check_test.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/domain_probe.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/push_check.go`
- Modify: `registrydto` (the two action DTOs), `sys_registry.go`, `router_system.go`

**Interfaces:**
- Produces: `(*service).ProbeDomain`, `(*service).CheckPush`, and the pure
  `proxyEvidence(header http.Header) []string`; endpoints
  `POST /system/settings/registry/probe-domain` and `POST /system/settings/registry/push-check`.

- [ ] **Step 1: Write the failing test**

`probe_test.go`:

```go
package registryserviceimpl

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Registry traffic through Cloudflare's proxy leaves the cluster, comes back, and
// meets an upload limit on the way. These headers are what says so from outside.
func TestProxyEvidenceFromCloudflare(t *testing.T) {
	header := http.Header{}
	header.Set("cf-ray", "8d2f00000000-SIN")
	header.Set("server", "cloudflare")

	evidence := proxyEvidence(header)

	assert.Len(t, evidence, 2)
	assert.Contains(t, evidence[0], "cf-ray")
	assert.Contains(t, evidence[1], "cloudflare")
}

func TestProxyEvidenceFromAGenericProxy(t *testing.T) {
	header := http.Header{}
	header.Set("via", "1.1 varnish")

	assert.Len(t, proxyEvidence(header), 1)
}

// A registry answering for itself is the state an operator is aiming for, and it
// must not be reported as a warning.
func TestProxyEvidenceFromZotItself(t *testing.T) {
	header := http.Header{}
	header.Set("content-type", "application/json")
	header.Set("docker-distribution-api-version", "registry/2.0")

	assert.Empty(t, proxyEvidence(header))
}
```

`push_check_test.go` uses `httptest` to stand in for the registry, because the decision being
tested is what each answer means:

```go
package registryserviceimpl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushCheckReadsTheAnswer(t *testing.T) {
	tests := []struct {
		name   string
		status int
		ok     bool
		detail string
	}{
		{name: "accepted", status: http.StatusAccepted, ok: true, detail: "went through"},
		{name: "too large", status: http.StatusRequestEntityTooLarge, ok: false, detail: "body limit"},
		{name: "unauthorized", status: http.StatusUnauthorized, ok: false, detail: "credential"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			result := readPushCheckAnswer(tt.status)

			assert.Equal(t, tt.ok, result.OK)
			assert.Contains(t, result.Detail, tt.detail)
		})
	}
}
```

- [ ] **Step 2: Run the tests to watch them fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run 'TestProxy|TestPushCheck' -v`
Expected: FAIL - `undefined: proxyEvidence`.

- [ ] **Step 3: Write the probe**

`probe.go`:

```go
package registryserviceimpl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

// ProbeDomain asks the registry's own address what answers there.
//
// It is advisory and says so: HivePaaS cannot see somebody's DNS settings, and a
// load balancer in front of the cluster is not a proxy in the sense that matters
// here. What it can do is name what it saw, early, instead of leaving an operator
// to discover it from a build that dies on a 100 MB layer.
func (s *service) ProbeDomain(ctx context.Context, domain string) (*registryservice.DomainProbe, error) {
	if strings.TrimSpace(domain) == "" {
		return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
			WithExtraDetail("a domain is required")
	}

	url := fmt.Sprintf("https://%s/v2/", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Not reachable is not the same as not proxied, and saying so is the
		// honest answer while the registry is still starting.
		return &registryservice.DomainProbe{Reached: false}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	evidence := proxyEvidence(resp.Header)
	return &registryservice.DomainProbe{
		Reached:  true,
		Proxied:  len(evidence) > 0,
		Evidence: evidence,
	}, nil
}

// proxyEvidence names the headers that say something answered instead of the
// registry. It returns sentences rather than header names, because they are shown
// to a person who is deciding what to change in their DNS.
func proxyEvidence(header http.Header) []string {
	var evidence []string
	if ray := header.Get("cf-ray"); ray != "" {
		evidence = append(evidence, fmt.Sprintf("a cf-ray header (%s), which only Cloudflare sends", ray))
	}
	if server := header.Get("server"); strings.EqualFold(server, "cloudflare") {
		evidence = append(evidence, "a server header saying cloudflare")
	}
	if via := header.Get("via"); via != "" {
		evidence = append(evidence, fmt.Sprintf("a via header (%s), so something is relaying this", via))
	}
	return evidence
}
```

- [ ] **Step 4: Write the push check**

`push_check.go` uploads to `/v2/hivepaas-selftest/blobs/uploads/` and cancels the upload with
a `DELETE` on the location it was given. Whatever survives a cancelled upload is an unfinished
upload, which zot's garbage collector removes on its next pass, so nothing has to be cleaned
up by hand.

```go
const (
	// defaultPushCheckBytes is above the 100 MB a proxy on a free plan allows per
	// request, which is the limit worth finding out about before a build does.
	defaultPushCheckBytes = 150 << 20
	pushCheckRepo         = "hivepaas-selftest"
	pushCheckTimeout      = 10 * time.Minute
)

// readPushCheckAnswer turns a status code into the sentence the dashboard shows.
func readPushCheckAnswer(status int) *registryservice.PushCheckResult {
	switch status {
	case http.StatusAccepted, http.StatusCreated, http.StatusNoContent:
		return &registryservice.PushCheckResult{
			OK: true, StatusCode: status,
			Detail: "A 150 MB upload went through: nothing in front of the registry is limiting request bodies.",
		}
	case http.StatusRequestEntityTooLarge:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail: "Something in front of the registry refused the upload with a body limit. " +
				"Set this domain to DNS-only so registry traffic reaches the cluster directly.",
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail:     "The registry refused the credential. Rotating it and redeploying may be needed.",
		}
	default:
		return &registryservice.PushCheckResult{
			StatusCode: status,
			Detail:     fmt.Sprintf("The upload was answered with %d.", status),
		}
	}
}
```

`CheckPush` itself opens a `POST` to start the upload, `PATCH`es `req.Bytes` of zeros from an
`io.LimitReader(zeroReader{}, n)` so nothing is held in memory, reads the answer through
`readPushCheckAnswer`, measures the elapsed time, and sends the cancelling `DELETE` whatever
happened. It uses a client with `pushCheckTimeout` rather than the service's 30-second one.

- [ ] **Step 5: Add the endpoints**

Two usecase methods (`ProbeDomain`, `CheckPush`) that load the setting, check the same
permission the update does, and delegate; two handler methods; and the routes:

```go
		registryGroup.POST("/probe-domain", systemSettingsHandler.ProbeRegistryDomain)
		registryGroup.POST("/push-check", systemSettingsHandler.CheckRegistryPush)
```

- [ ] **Step 6: Run the tests, regenerate, commit**

```bash
go test ./hivepaas_app/... && make gen-swag && go build ./... && golangci-lint run ./...
git add hivepaas_app docs/openapi
git commit -m "$(cat <<'EOF'
feat(registry): detect a proxy in front of the registry and check a large push

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: Rotating the credential

**Files:**
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/rotate.go`
- Create: `hivepaas_app/service/registryservice/registryserviceimpl/rotate_test.go`
- Create: `hivepaas_app/usecase/systemsettings/registryuc/credential_rotate.go`
- Modify: `registrydto`, `sys_registry.go`, `router_system.go`,
  `hivepaas_app/entity/setting_registry.go` (two more stored fields)

**Interfaces:**
- Produces: `(*service).RotateCredential`, the pure
  `rotatedHtpasswd(newLine, previousLine string, previousStillValid bool) string`, the
  endpoint `POST /system/settings/registry/rotate-credential`, and two fields on
  `RegistrySettings`:

```go
	// CredentialRotatedAt is when the password last changed, and GraceEndsAt is
	// when the one before it stops working. A service that has not been
	// redeployed still presents the credential it was deployed with, so the
	// previous password has to keep working for long enough to redeploy.
	CredentialRotatedAt time.Time `json:"credentialRotatedAt,omitzero"`
	CredentialGraceEnds time.Time `json:"credentialGraceEnds,omitzero"`
```

- [ ] **Step 1: Write the failing test**

`rotate_test.go`:

```go
package registryserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Both lines have to be in the file during the grace period: a service that was
// not redeployed still presents the old password when a node reschedules it.
func TestRotatedHtpasswdKeepsThePreviousLine(t *testing.T) {
	content := rotatedHtpasswd("hivepaas:$2y$10$new", "hivepaas:$2y$10$old", true)

	assert.Equal(t, 2, strings.Count(content, "\n"))
	assert.Contains(t, content, "$2y$10$new")
	assert.Contains(t, content, "$2y$10$old")
}

// Once the grace period is over the old line goes, which is the whole point of
// having one.
func TestRotatedHtpasswdDropsAnExpiredLine(t *testing.T) {
	content := rotatedHtpasswd("hivepaas:$2y$10$new", "hivepaas:$2y$10$old", false)

	assert.Equal(t, "hivepaas:$2y$10$new\n", content)
}

func TestRotatedHtpasswdOnAFirstRotation(t *testing.T) {
	assert.Equal(t, "hivepaas:$2y$10$new\n", rotatedHtpasswd("hivepaas:$2y$10$new", "", true))
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/registryservice/... -run TestRotated -v`
Expected: FAIL - `undefined: rotatedHtpasswd`.

- [ ] **Step 3: Write the rotation**

```go
const credentialGracePeriod = 14 * 24 * time.Hour

// rotatedHtpasswd is the file during and after a rotation. zot accepts several
// lines for one account, so the previous password keeps working until the grace
// period ends without anything special being done to support it.
func rotatedHtpasswd(newLine, previousLine string, previousStillValid bool) string {
	if !previousStillValid {
		return htpasswdContent(newLine)
	}
	return htpasswdContent(newLine, previousLine)
}
```

`RotateCredential` generates a password, writes it into the registry-auth setting through
`updateRegistryAuthSetting`, builds the new line, keeps the line that is in the secret today as
the previous one, writes the file through `applyHtpasswd`, and stores `CredentialRotatedAt`
and `CredentialGraceEnds = now + credentialGracePeriod` in the registry setting. The next
`Apply` after the grace period ends rebuilds the file without the old line, which is why
`Apply` reads those two fields when it builds `in.Htpasswd`:

```go
	previous := ""
	if !cfg.CredentialGraceEnds.IsZero() && timeutil.NowUTC().Before(cfg.CredentialGraceEnds) {
		previous = s.previousHtpasswdLine(ctx, db, app)
	}
	in.Htpasswd = rotatedHtpasswd(line, previous, previous != "")
```

- [ ] **Step 4: Add the endpoint**

`POST /system/settings/registry/rotate-credential`, returning
`{"data": {"graceEndsAt": "2026-10-05T10:00:00Z"}}`.

- [ ] **Step 5: Run everything, regenerate, commit**

```bash
go test ./... && make gen-swag && go build ./... && golangci-lint run ./...
git add hivepaas_app docs/openapi
git commit -m "$(cat <<'EOF'
feat(registry): rotate the registry credential with a grace period

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 11: The dashboard's data layer

**Files (all under `hivepaas-dashboard/src/application/modules/system-settings`):**
- Create: `domain/hivepaas-registry-settings.entity.ts`
- Create: `api/services/hivepaas-registry-settings-services/hivepaas-registry-settings.api.contracts.ts`
- Create: `api/services/hivepaas-registry-settings-services/hivepaas-registry-settings.api.validator.ts`
- Create: `api/services/hivepaas-registry-settings-services/hivepaas-registry-settings.api.ts`
- Create: `api/services/hivepaas-registry-settings-services/index.ts`
- Create: `api/hooks/use-hivepaas-registry-settings.api.ts`
- Create: `data/queries/hivepaas-registry-settings.queries.ts`
- Create: `data/commands/hivepaas-registry-settings.commands.ts`
- Modify: the `index.ts` barrels beside each of them

**Interfaces:**
- Consumes: the wire shape in Task 8, plus the three action endpoints from Tasks 9 and 10.
- Produces: `HivePaaSRegistrySettings` (domain entity), `useHivePaaSRegistrySettingsQuery`,
  `useUpdateHivePaaSRegistrySettingsCommand`, `useProbeRegistryDomainCommand`,
  `useCheckRegistryPushCommand`, `useRotateRegistryCredentialCommand`.

- [ ] **Step 1: Copy the logging feature's data layer file by file**

Every file in this task has a sibling one directory over
(`hivepaas-logging-settings.*`). Copy each, rename the symbols, and replace the shape. The
files are small and mechanical; the only decisions are below.

- [ ] **Step 2: Write the contracts**

`hivepaas-registry-settings.api.contracts.ts` declares the request and response types from the
JSON in Task 8. `storage.volume` and `storage.cloudStorage` are objects on the wire in both
directions - `{id}` going out, the whole setting coming back - which is the same asymmetry
logging has for its volume, and the form mappers are where it is resolved.

- [ ] **Step 3: Write the validator**

`hivepaas-registry-settings.api.validator.ts` parses the response with zod, the way logging's
does. Parse `registryStatus` as optional: it is absent until the registry is provisioned.

- [ ] **Step 4: Write the queries and commands**

One query (`GET`), four commands (`PUT`, probe, push check, rotate). The push check command
needs a longer timeout than the default: it uploads 150 MB and waits for the answer.

- [ ] **Step 5: Verify**

```bash
cd ../hivepaas-dashboard && npm run typecheck && npm run lint
```
Expected: clean.

- [ ] **Step 6: Commit** (in the dashboard repo)

```bash
git add src/application/modules/system-settings
git commit -m "$(cat <<'EOF'
feat(registry): add the registry settings data layer

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 12: The dashboard's page

**Files:**
- Create: `layouts/registry/layout/registry-layout.com.tsx` and its `index.ts`
- Create: `routes/registry/route/system-settings-registry.route.com.tsx` and its `index.ts`
- Create: `routes/registry/schemas/hivepaas-registry-settings.schema.ts` and its `index.ts`
- Create: `routes/registry/form/hivepaas-registry-settings.form.com.tsx`,
  `hivepaas-registry-settings.form-mappers.ts` and its `index.ts`
- Create: `routes/registry/building-blocks/{storage-section,cleanup-section,registry-status-section,rotate-credential.dialog}.com.tsx`
  and its `index.ts`
- Create: `routes/registry/index.ts`
- Modify: `src/application/shared/constants/route.constants.ts`,
  `src/application/shared/layouts/module/module-sidebar/module-sidebar.com.tsx`,
  `src/application/modules/system-settings/system-settings.router.tsx`,
  `src/application/modules/system-settings/routes/index.ts`,
  `.../layouts/index.ts`

**Interfaces:**
- Consumes: Task 11's hooks; the shared `OptionCardGroup` from
  `~/projects/module-shared/components/option-card-group`.

- [ ] **Step 1: Add the route**

In `route.constants.ts`, beside `logging`:

```ts
        registry: {
            $pattern: "system/registry",
            $route: "/system/registry/configuration/",

            configuration: {
                $pattern: "system/registry/configuration",
                $route: "/system/registry/configuration/",
            },
        },
```

In `module-sidebar.com.tsx`, after the Logging item in the same group:

```tsx
                    {
                        title: "Registry",
                        icon: Container,
                        route: ROUTE.systemSettings.registry.configuration.$route,
                        pattern: ROUTE.systemSettings.registry.$pattern,
                    },
```

`Container` comes from `lucide-react`, beside the other icons imported there. Then add the
branch to `system-settings.router.tsx`, copying logging's and swapping the names.

- [ ] **Step 2: Write the form schema**

`hivepaas-registry-settings.schema.ts`:

```ts
import { z } from "zod";

export const REGISTRY_STORAGE_TYPES = ["volume", "s3"] as const;

export const HivePaaSRegistrySettingsFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string().trim(),
        storageType: z.enum(REGISTRY_STORAGE_TYPES),
        // The selected setting's id, because that is what a Select holds. On the
        // wire it is an object either way: { id } in a request, the whole setting
        // in a response. The form mappers convert at both ends.
        volumeId: z.string(),
        cloudStorageId: z.string(),
        cleanupEnabled: z.boolean(),
        keepLast: z.number().int().min(1).max(1000),
        keepDays: z.number().int().min(1).max(3650),
        memoryLimit: z
            .string()
            .trim()
            .regex(/^$|^\d+(b|kb|mb|gb|tb)$/i, "Use a size such as 1gb or 512mb"),
    })
    .superRefine((value, ctx) => {
        if (!value.enabled) return;

        // Images are named after it and docker only speaks to a registry over
        // HTTPS at a name, so there is nothing to provision without one.
        if (!value.domain) {
            ctx.addIssue({ code: "custom", path: ["domain"], message: "A domain is required" });
        }
        if (value.storageType === "volume" && !value.volumeId) {
            ctx.addIssue({ code: "custom", path: ["volumeId"], message: "Choose a volume" });
        }
        if (value.storageType === "s3" && !value.cloudStorageId) {
            ctx.addIssue({ code: "custom", path: ["cloudStorageId"], message: "Choose a cloud storage" });
        }
    });

export type HivePaaSRegistrySettingsFormValues = z.output<typeof HivePaaSRegistrySettingsFormSchema>;
```

- [ ] **Step 3: Write the storage section**

`storage-section.com.tsx` uses the shared `OptionCardGroup` - the same component App Kind's
Category, Deployment's method, Routing's protocol and Availability's service mode use - with
two cards, and renders the volume or the cloud-storage picker beneath the selected one:

```tsx
const STORAGE_OPTIONS: OptionCard<RegistryStorageType>[] = [
    {
        value: "volume",
        label: "Local volume",
        description:
            "Images are kept on one node's disk, and the registry runs on that node. Layers shared between apps are stored once.",
        icon: HardDrive,
    },
    {
        value: "s3",
        label: "S3 storage",
        description:
            "Images are kept in a bucket, so the registry can run on any node. A layer two apps share is stored twice.",
        icon: Cloud,
    },
];
```

Both cards carry `disabled` once `registryStatus.provisioned` is true, with the reason
underneath: *"The storage cannot be changed once the registry holds images. Provision a new
registry instead."* Disabled, not hidden - an operator asking "can I move this to S3?" should
find the answer where they looked.

- [ ] **Step 4: Write the cleanup section**

Two numbers and a switch, plus the sentence that reads them back, because two numbers in a
form are not a policy anybody can picture:

```tsx
const summary = cleanupEnabled
    ? `Keeps the last ${keepLast} ${keepLast === 1 ? "build" : "builds"} of every app, and everything from the past ${keepDays} ${keepDays === 1 ? "day" : "days"}. Anything else is removed, and its disk space comes back within a couple of hours.`
    : "Nothing is removed. The registry grows until the disk or the bucket does.";
```

- [ ] **Step 5: Write the status section and the actions**

`registry-status-section.com.tsx` shows, once provisioned: a link to the app, the number of
repositories and what they hold (`formatBytes(storedBytes)`), a link to the credential's
registry-auth setting, and the three actions.

- **Check the domain** runs the probe and shows the evidence sentences, or "Nothing is
  answering at that address yet" when `reached` is false.
- **Test a large push** runs the push check. It takes a while, so the button shows a spinner
  and the result stays on screen.
- **Rotate password** opens a warning modal in the house style - title, `<Separator />`, then
  the consequence - saying that apps which are not redeployed keep working until the grace
  period ends, and that they have to be redeployed before it does.

The DNS-only sentence sits under the domain field itself, not behind a probe: it is what the
operator needs while they are typing, not after something fails.

- [ ] **Step 6: Verify**

```bash
npm run typecheck && npm run lint && npm run build
```
Expected: clean.

- [ ] **Step 7: Commit**

```bash
git add src/application
git commit -m "$(cat <<'EOF'
feat(registry): add the registry settings page

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 13: End to end, on a real cluster

Nothing above proves that zot accepts what this code writes. The spike proved it for a
configuration written by hand; this task proves it for the configuration the code generates.

**Files:** none. This is a verification task, and what it produces is either a green run or a
bug to fix in the task that owns it.

- [ ] **Step 1: Provision on a volume**

With the backend and dashboard running against the development cluster: open System Settings ->
Registry, give it a domain that resolves to the cluster, choose a volume, save. Expect the app
`registry` to appear in the `hivepaas` project and reach a running state.

- [ ] **Step 2: Check the registry answers**

```bash
curl -su hivepaas:<password from the registry auth setting> https://<domain>/v2/
```
Expected: `200`. Anonymously: `401`.

- [ ] **Step 3: Push through it**

Point an app's deployment settings at the new registry auth, build it, and watch the push
succeed. Then confirm the tag:

```bash
curl -su hivepaas:<password> https://<domain>/v2/hivepaas/<app key>/tags/list
```
Expected: the commit's short hash.

- [ ] **Step 4: Redeploy from it**

Redeploy that app and watch the task pull the image from the system registry, which is what
proves the credential reaches swarm.

- [ ] **Step 5: Check the cleanup is configured**

```bash
docker config inspect $(docker service inspect hivepaas_registry \
  --format '{{range .Spec.TaskTemplate.ContainerSpec.Configs}}{{.ConfigName}} {{end}}' | tr ' ' '\n' | head -1) \
  --format '{{json .Spec.Data}}' | base64 -d | jq .storage.retention
```
Expected: the two numbers from the form, as `mostRecentlyPushedCount` and the two `*Within`
windows.

- [ ] **Step 6: Rotate, then check the old credential still works**

Rotate the password in the dashboard, then `curl -u hivepaas:<the old password> https://<domain>/v2/`.
Expected: `200`, because the grace period has not ended. The new one works too.

- [ ] **Step 7: Provision the S3 variant**

On a scratch installation, or after deleting the app and the setting: choose S3 with a
cloud-storage setting pointed at a bucket, and repeat steps 2 and 3. Expect no volume, no
placement constraint on the service, and `dedupe: false` in the rendered configuration.

---

## Notes from the self-review

- **Spec coverage.** §4 is Task 1, §5 Tasks 5-6, §6 Task 2, §7 Task 2's retention rules,
  §8 Task 9, §9 Task 8's wire format plus what already exists downstream, §10 Task 4's
  validation and Task 6's reconcile, §11 Task 10, §12 Task 7's status, §13 Tasks 11-12,
  §14 the tests inside each task plus Task 13.
- **Two fields the spec does not name** were added in Task 10 - `CredentialRotatedAt` and
  `CredentialGraceEnds` - because §11 describes a grace period and nothing in §4 could hold
  it. They are additive, and the spec's §11 is what they implement.
- **One deliberate deviation:** the spec writes `VolumeID string` / `CloudStorageID string`;
  Task 1 stores `entity.ObjectID` instead, matching every other setting that references
  another one. Reviewers should accept or reject that in Task 1, before anything is built on
  it.
- **Not in this plan, on purpose:** the keep-set cleanup mode (§7's extension point), a page
  listing what each repository holds beyond the totals, `distribution` as a second type, and
  per-project credentials. They are §15 of the spec.
