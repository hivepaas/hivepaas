# Setting Mounts, One Way - Backend Sources Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Secrets and config files become sources of setting mounts, and lose their own `swarmRef`. Setting mounts are then the only way a file reaches an app's container. Templates and HivePaaS's own apps keep working.

**Architecture:**
- **The parts registry.** It gains `secret.value` and `config-file.content`, and its one "sensitive" flag becomes two:
  - `Secret`, stored as a Docker secret;
  - `Gated`, asks for Reveal Secrets.
- **What goes.** `SwarmRef` leaves the entities and DTOs. With it goes everything that made Docker objects for secrets and config files: most of `clustersecretservice`, the use cases' Docker calls, provisioning's `applySwarmFiles`, import's `updateSwarmFiles`, and plan 2's `makeRoom`. Provisioning and clones call `settingMountService.Refresh` or `RemoveApp` instead.
- **Templates.** A template's `swarmRef.file` on a secret or config file is read by the builder as shorthand for an `app-setting-mount` entry. Templates may set `inheritable`.
- **HivePaaS's own apps.** `systemappservice` and the registry write settings and entries, then refresh.

**Tech Stack:** Go, testify.

**Spec:** `docs/superpowers/specs/2026-09-25-setting-mounts-one-way-design.md` (§1-§3, plan 1 of §5), amending `2026-09-25-setting-mounts-design.md`.

## Global Constraints

- **Parts:**

  | type | part | required | secret | gated |
  |---|---|---|---|---|
  | `secret` | `value` | yes | yes | |
  | `config-file` | `content` | yes | | |
  | `ssl-cert` | `certificate` | yes | | |
  | | `privateKey` | yes | yes | yes |
  | | `caCertificate` | | | |
  | `ssh-key` | `privateKey` | yes | yes | yes |
  | | `publicKey` | | | |
  | `basic-auth` | `username` | yes | | |
  | | `password` | yes | yes | yes |
  | | `htpasswd` | yes | yes | yes |

- **Base64:** a base64 secret or config file is decoded before its file is written (`Secret.ValueAsBytes`, `ConfigFile.ContentAsBytes`).
- **The gate** counts gated parts only (`GatedPart`, `Grants`).
- **Wire changes** (the dashboard is not built yet):
  - entry files: `{part, path, uid, gid, mode, secret, gated}`, where `secret` and `gated` replace `sensitive`;
  - `sources` parts: `{name, required, secret, gated}`.
- **No migration:** a `swarmRef` in a stored row or an imported bundle is ignored.
- **Template shorthand:**
  - `secrets.<KEY>.swarmRef.file: {name, uid, gid, mode}` and `configFiles.<name>.swarmRef.file` become an entry;
  - the entry key is the setting's key, lowercased, with every character outside `[a-z0-9-]` turned into `-`, runs of `-` collapsed, trimmed of `-` at the ends, and cut to 20 characters;
  - the source is the new setting, and there is one file (part `value` or `content`);
  - `Inheritable` is the setting's own;
  - `secrets.<KEY>.inheritable` and `configFiles.<name>.inheritable` (bool) are allowed in templates.
- **Gates:** `go build ./...`, `golangci-lint run ./...` (whole repo), `go test ./...`, `make gen-swag`.
- **Git:**
  - branch `feat/setting-mounts-one-way`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, delete the branch, do not push;
  - stage only the files named.

## Review Focus

1. **A template secret key like `DB_PASSWORD` or a config name like `config.json` makes a valid entry key**, and two settings whose keys collapse to the same entry key do not collide silently. *Test: Task 3, `TestEntryKeyOfATemplateSetting`, and the build refusing a clash.*
2. **A base64 config file mounts its bytes, not its base64 text.** *Test: Task 1, `TestSecretsAndConfigFilesRenderTheirBytes`.*
3. **A secret mounted from a template is a Docker secret, and mounting it asks nothing.** *Tests: Task 1, the registry row; Task 3, the build test's entry.*
4. **The registry's config and htpasswd still reach its container after a settings change.** *Test: Task 4, the registry apply test asserts a refresh.*
5. **A failed provisioning or clone leaves no mount objects behind.** *Test: Task 2, the provisioning cleanup test asserts `RemoveApp`.*

---

### Task 1: Secrets and config files as sources; the secret/gated split

**Files:**
- Modify: `hivepaas_app/service/settingmountservice/`:
  - `parts.go` (`Part.Secret`, `Part.Gated`, two source types);
  - `checks.go` (`SensitivePart` becomes `GatedPart`, used by `Grants`);
  - `service.go` (`File.Sensitive` becomes `File.Secret`).
- Modify: `settingmountserviceimpl/resolve.go` (`Secret: part.Secret`) and `settingmountserviceimpl/apply.go` (`file.Secret`).
- Modify: `hivepaas_app/usecase/settings/settingmountuc/settingmountdto/get.go` (file resp `Secret`, `Gated`), `sources.go` (part `Secret`, `Gated`), `settingmountuc/sources.go`.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_checks.go` (`GatedPart`).
- Tests: `parts_test.go`, `checks_test.go`, the engine tests using `Sensitive`, and `settingmountuc/gate_test.go` as needed.

**Interfaces:**
- Produces:
  - `Part{Name; Required, Secret, Gated bool; Version int; Inputs []string; Render}`;
  - `GatedPart(name string) bool`;
  - `File.Secret bool`;
  - the source types `base.SettingTypeSecret` (part `value`) and `base.SettingTypeConfigFile` (part `content`).

- [ ] **Step 1: Branch.** `git checkout main && git checkout -b feat/setting-mounts-one-way`

- [ ] **Step 2: Write the failing tests.** In `parts_test.go`:
  - replace `TestSensitivePartsAreTheSecretOnes` with the two tests below;
  - add `TestSecretsAndConfigFilesRenderTheirBytes`:

```go
func TestPartsStoredAsSecretsAndPartsGated(t *testing.T) {
	secret, gated := map[string]bool{}, map[string]bool{}
	for _, typ := range SourceTypes() {
		for _, part := range PartsOf(typ) {
			if part.Secret {
				secret[string(typ)+"/"+part.Name] = true
			}
			if part.Gated {
				gated[string(typ)+"/"+part.Name] = true
			}
		}
	}
	assert.Equal(t, map[string]bool{
		"secret/value": true, "ssl-cert/privateKey": true, "ssh-key/privateKey": true,
		"basic-auth/password": true, "basic-auth/htpasswd": true,
	}, secret)
	assert.Equal(t, map[string]bool{
		"ssl-cert/privateKey": true, "ssh-key/privateKey": true,
		"basic-auth/password": true, "basic-auth/htpasswd": true,
	}, gated, "a secret's value is the app's already: mounting it reveals nothing new")
}

// A base64 setting holds bytes; the file is those bytes.
func TestSecretsAndConfigFilesRenderTheirBytes(t *testing.T) {
	useDataKey(t)
	plainSecret := sourceSetting(t, base.SettingTypeSecret, &entity.Secret{
		Key: "DB_PASSWORD", Value: entity.NewEncryptedField("s3cret")})
	binarySecret := sourceSetting(t, base.SettingTypeSecret, &entity.Secret{
		Key: "KEYSTORE", Value: entity.NewEncryptedField("AAEC"), Base64: true})
	plainConfig := sourceSetting(t, base.SettingTypeConfigFile, &entity.ConfigFile{
		Name: "app.conf", Content: "listen 80"})
	binaryConfig := sourceSetting(t, base.SettingTypeConfigFile, &entity.ConfigFile{
		Name: "blob", Content: "AAEC", Base64: true})

	for _, tc := range []struct {
		setting *entity.Setting
		part    string
		want    []byte
	}{
		{plainSecret, "value", []byte("s3cret")},
		{binarySecret, "value", []byte{0, 1, 2}},
		{plainConfig, "content", []byte("listen 80")},
		{binaryConfig, "content", []byte{0, 1, 2}},
	} {
		values, err := Values(tc.setting)
		assert.NoError(t, err)
		out, err := PartOf(tc.setting.Type, tc.part).RenderFrom(values)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, out, "%s %s", tc.setting.Type, tc.part)
	}
}
```

In `checks_test.go`, rename `TestSensitiveByNameWhateverTheType` to `TestGatedByNameWhateverTheType`, calling `GatedPart`. Add `"value": false, "content": false` to its table.

- [ ] **Step 3: Run** `go test ./hivepaas_app/service/settingmountservice/`. It should FAIL to compile (`part.Secret`, `GatedPart`).

- [ ] **Step 4: Implement.** In `parts.go`:
  - replace `Sensitive bool` in `Part` with:

```go
	// Secret parts are stored as Docker secrets; the others as Docker configs.
	Secret bool
	// Gated parts take the Reveal Secrets permission to mount: the app has no
	// other way to read them. A secret's value is not gated - the app reads it
	// through ${secrets.NAME} already.
	Gated bool
```

  - in the registry, set `Secret: true, Gated: true` where `Sensitive: true` was;
  - add the two source types, with field constants `fieldValue = "value"` and `fieldContent = "content"`:

```go
	base.SettingTypeSecret: {
		parts:  []*Part{{Name: fieldValue, Required: true, Secret: true, Version: 1, Inputs: []string{fieldValue}}},
		values: secretValues,
	},
	base.SettingTypeConfigFile: {
		parts:  []*Part{{Name: fieldContent, Required: true, Version: 1, Inputs: []string{fieldContent}}},
		values: configFileValues,
	},
```

```go
func secretValues(setting *entity.Setting) (map[string]string, error) {
	secret, err := setting.AsSecret()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	value, err := secret.ValueAsBytes()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldValue: string(value)}, nil
}

func configFileValues(setting *entity.Setting) (map[string]string, error) {
	configFile, err := setting.AsConfigFile()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldContent: string(configFile.ContentAsBytes())}, nil
}
```

Check that `ValueAsBytes` and `ContentAsBytes` decode base64 (`sed -n 40,80p hivepaas_app/entity/setting_secret.go hivepaas_app/entity/setting_config_file.go`). If `ContentAsBytes` does not, decode there with `base64.StdEncoding` when `Base64` is set, and say so in the ledger.

In `checks.go`:
- `SensitivePart` becomes `GatedPart`, testing `part.Gated`;
- `Grants` uses `GatedPart`, and its comment says gated;
- the `Grant` comment says "gated part".

In `service.go`, `File.Sensitive` becomes `File.Secret`, with the comment "stored as a Docker secret". Then:
- `resolve.go` sets `Secret: part.Secret`;
- `apply.go` reads `file.Secret`;
- `settingmountdto/get.go`'s `SettingMountFileResp` has `Secret bool \`json:"secret"\`` and `Gated bool \`json:"gated"\``, set from `PartOf(source type, part)` when the source is known, else `GatedPart(part)` for `gated` and `false` for `secret`;
- `sources.go`'s `SettingMountPart` has `Secret` and `Gated` in place of `Sensitive`, set in `settingmountuc/sources.go`;
- `import_checks.go`'s `entryGrants` uses `GatedPart`.

`TransformSettingMount` knows the source's type from `refObjects.RefSettings[mount.Source.ID].Type` when present. Use `settingmountservice.PartOf(type, f.Part)` there.

- [ ] **Step 5: Run** `go build ./... && go test ./hivepaas_app/service/settingmountservice/... ./hivepaas_app/usecase/settings/... ./hivepaas_app/service/specservice/...`. It should PASS once the tests using `Sensitive:` in `resolve_test.go` and `apply_test.go` read `Secret:`. Then run `golangci-lint run ./hivepaas_app/...`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/settingmountservice/ hivepaas_app/usecase/settings/settingmountuc/ \
  hivepaas_app/service/specservice/specserviceimpl/import_checks.go
git commit -m "feat(settingmounts): secrets and config files are sources; stored as secret and gated are two things

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: `swarmRef` leaves secrets and config files

**Files:**
- **Entities:** in `hivepaas_app/entity/setting_secret.go` and `setting_config_file.go`, delete `SwarmRef`, `SwarmSecretRef`, `SwarmConfigRef` and `SwarmRefFileTarget`. Fix `entity/setting_spec_test.go`.
- **DTOs:** in `hivepaas_app/usecase/settings/secretuc/secretdto/{create,get}.go` and `configfileuc/configfiledto/{create,get}.go`, delete the `swarmRef` requests and responses.
- **Use cases:** in `hivepaas_app/usecase/settings/secretuc/{create,update,update_status,delete}.go` and `configfileuc/{create,update,update_status,delete}.go`, delete the Docker calls, the re-persist of `SwarmRef` ids, and the `CheckMountPaths` hooks.
- **Base use case:** in `hivepaas_app/usecase/settings/mount_paths.go` and `base_uc.go`, delete `CheckMountPaths`, `CheckMountPathsAfterLoading`, the test file and the `SettingMountService` field, if nothing else uses them.
- **Engine and helpers:** in `hivepaas_app/service/settingmountservice/`, delete `checks.go`'s `SecretFileTarget` and `ConfigFileTarget` with their test. `settingmountserviceimpl/claimed.go`'s `ClaimedPaths` keeps entries only; update `claimed_test.go`.
- **`clustersecretservice`:** delete the package (`service/clustersecretservice/`) if Task 4's rework leaves no caller. Otherwise keep only what that caller uses. Remove it from `registry/provides.go`.
- **Provisioning:** in `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/`:
  - `apply_config.go`: `applySwarmFiles` becomes `s.settingMountService.Refresh(ctx, db, app)`, and `ApplyAppConfigurationResp.Configs`/`Secrets` go;
  - `provision_apps.go`: the cleanup calls `settingMountService.RemoveApp(ctx, appID)`;
  - `service.go` takes the dependency;
  - fix the tests `apply_config_test.go`, `provision_apps_test.go` and `provision_test.go`;
  - `service/appprovisionservice/service.go`: the response type.
- **Clone:**
  - `clone.go`: the cleanup's `SecretsRemove` and `ConfigsRemove` become `settingMountService.RemoveApp(ctx, destApp.ID)`;
  - `post_clone_configuration.go`: `DestConfig`/`DestSecrets` go;
  - `clone_3_swarm_service.go` keeps clearing `Secrets`/`Configs` on the copy (the clone resolves its own).
- **Deletion:** in `hivepaas_app/service/appservice/appserviceimpl/deletion.go`, delete `getDockerSecretsAndConfigs` and `deleteDockerSecretsAndConfigs`. `settingMountService.RemoveApp` is what removes an app's objects now.
- **Import:** in `hivepaas_app/service/specservice/specserviceimpl/import_phase2.go`, delete `updateSwarmFiles` and its call. The refresh task `write` records covers changed sources. Fix `import_phase2_test.go` and `fakes_apply_test.go`.
- **Also:** the `fakeClusterSecretService` and `clusterSecretService` parameter in `specserviceimpl/service.go` `New` and `export_test.go`, if the service goes.

**Interfaces:**
- Consumes: `settingmountservice.Service.Refresh`, `RemoveApp` (plan 1 of the engine).
- Produces: `appprovisionservice.ApplyAppConfigurationResp` without `Configs`/`Secrets`.

- [ ] **Step 1: Write the failing test.** In `appprovisionserviceimpl/apply_config_test.go`, add a fake `settingmountservice.Service` recording `Refresh` and `RemoveApp`, then:

```go
// The app's files are its setting mounts': configuring a provisioned app
// brings its service to them, in one update.
func TestConfiguringAnAppRefreshesItsSettingMounts(t *testing.T) {
	svc, mounts := newApplyConfigService(t) // the file's existing constructor, given the fake
	app := &entity.App{ID: "app_1", ServiceID: "svc_1"}

	_, err := svc.ApplyAppConfiguration(context.Background(), nil, &appprovisionservice.ApplyAppConfigurationReq{App: app})

	assert.NoError(t, err)
	assert.Equal(t, []string{"app_1"}, mounts.refreshed)
}
```

In `provision_apps_test.go`, the cleanup test asserts `mounts.removed` contains the provisioned app's id, instead of secret and config ids. Read both files first, and build the fake into the constructors they already use (`grep -n "func new\|clusterSecretService" hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/*_test.go`).

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/appprovisionservice/...`. It should FAIL.

- [ ] **Step 3: Remove, and route through the engine.** Delete the fields and types in the entities first; `go build ./...` then lists every use. Deal with each:
  - **The secret use cases.** Delete the `IsAppScope`/`App != nil` Docker blocks and the re-persist. In `secretuc/update.go`, keep the env var rebuild (`secretValueChanged`).
  - **Sources that change.** Settings written through the use cases already record a refresh through their events (engine plan 1). When a secret or config file that an entry mounts changes, the refresh task updates the file.
  - **`clustersecretservice`'s remaining users.** Registry `apply.go` and `systemappservice/secrets.go` are Task 4. For this task, replace their calls with a `TODO`-free minimum that compiles: persist the setting without the Docker call, and note in the ledger that Task 4 adds the refresh. If that leaves the package unused, delete it.
  - **Tests** that asserted `swarmRef` in `buildable_test.go`, `build_test.go` and `import_write_apps_test.go` change with Task 3's shorthand. For this task, make `swarmRef` on a secret or config file in a template build to a setting without it, and let Task 3 add the entry. Keep the buildable check accepting `swarmRef`.
  - **Rows** with a `swarmRef` key parse without error: `encoding/json` ignores unknown fields, and the setting parser uses it. Check this with a test in `entity/setting_secret_test.go`:

```go
// Nothing is carried over from swarmRef: a row that has one reads as a secret
// without it.
func TestASecretRowWithASwarmRefReadsWithoutIt(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeSecret, Data: `{"key":"A","value":"","swarmRef":{"file":{"name":"a"}}}`}
	got, err := setting.AsSecret()
	assert.NoError(t, err)
	assert.Equal(t, "A", got.Key)
}
```

  Use a `value` form the `EncryptedField` unmarshal accepts; check `setting_secret_test.go` or `setting_encryption.go` for how an empty one is written.

- [ ] **Step 4: Run** `go build ./... && go test ./...`, which should PASS. Then run `golangci-lint run ./...` and `make gen-swag`; the DTOs changed.

- [ ] **Step 5: Commit**

```bash
git add -A hivepaas_app docs/openapi/swagger.json
git status --short   # check that only files of this task are staged; unstage anything else
git commit -m "refactor(settings): secrets and config files mount through setting mounts only

swarmRef leaves the secret and config file settings, their API and export.
Nothing makes Docker objects for them any more: an app's files are its
setting mounts'.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The template shorthand, and `inheritable` in templates

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/buildable.go`: `checkSecrets` and `checkConfigFiles` accept `inheritable` (a bool); `checkSwarmRef` stays.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/build_settings.go`: `buildSecrets` and `buildConfigFiles` read `swarmRef` and `inheritable` and make the entry.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/build.go`: `addNamedSetting` returns the setting it adds.
- Create: `hivepaas_app/service/specservice/specserviceimpl/build_mounts.go` (`templateEntryKey`, `addTemplateMount`)
- Test: `build_test.go`, `specmodel/buildable_test.go`

**Interfaces:**
- Produces:
  - `templateEntryKey(name string) string`;
  - `(state *buildState) addNamedSetting(...) (*entity.Setting, error)`;
  - `(state *buildState) addTemplateMount(key string, source *entity.Setting, part string, file map[string]any, inheritable bool) error`.

- [ ] **Step 1: Write the failing tests.** In `build_test.go`, add the two tests below. Find a template-build helper there (`grep -n "^func build\|^func buildDoc\|BuildApp(" hivepaas_app/service/specservice/specserviceimpl/build_test.go`) and use it.

```go
func TestEntryKeyOfATemplateSetting(t *testing.T) {
	for name, want := range map[string]string{
		"DB_PASSWORD": "db-password", "config.json": "config-json", "mosquitto-conf": "mosquitto-conf",
		"__A__": "a", strings.Repeat("x", 30): strings.Repeat("x", 20),
	} {
		assert.Equal(t, want, templateEntryKey(name), name)
	}
}

// A template's swarmRef is shorthand for a setting mount: the setting is built
// without it, and an entry mounts it.
func TestATemplateMountBecomesAnEntry(t *testing.T) {
	settings := buildTemplateSettings(t, map[string]any{
		"configFiles": map[string]any{"config.json": map[string]any{
			"content": "{}", "inheritable": true,
			"swarmRef": map[string]any{"file": map[string]any{"name": "/etc/app/config.json", "mode": 444}},
		}},
		"secrets": map[string]any{"DB_PASSWORD": map[string]any{
			"value":    "s3cret",
			"swarmRef": map[string]any{"file": map[string]any{"name": "/run/secrets/db", "mode": 400}},
		}},
	})

	byName := map[string]*entity.Setting{}
	for _, setting := range settings {
		byName[string(setting.Type)+"/"+setting.Name] = setting
	}
	config, secret := byName["config-file/config.json"], byName["secret/DB_PASSWORD"]
	configMount, secretMount := byName["app-setting-mount/config-json"], byName["app-setting-mount/db-password"]
	if !assert.NotNil(t, config) || !assert.NotNil(t, secret) ||
		!assert.NotNil(t, configMount) || !assert.NotNil(t, secretMount) {
		return
	}
	assert.True(t, config.Inheritable)
	assert.True(t, configMount.Inheritable, "the entry is inheritable with its source")
	assert.False(t, secretMount.Inheritable)
	assert.Equal(t, &entity.AppSettingMount{Source: entity.ObjectID{ID: config.ID}, Files: []*entity.AppSettingMountFile{
		{Part: "content", Path: "/etc/app/config.json", Mode: fileutil.FileMode(0o444)},
	}}, configMount.MustAsAppSettingMount())
	assert.Equal(t, "value", secretMount.MustAsAppSettingMount().Files[0].Part)
}

func TestTwoTemplateSettingsMayNotMountUnderOneEntryKey(t *testing.T) {
	_, err := buildTemplateSettingsErr(t, map[string]any{
		"configFiles": map[string]any{
			"app.conf": map[string]any{"content": "a",
				"swarmRef": map[string]any{"file": map[string]any{"name": "/a"}}},
			"app-conf": map[string]any{"content": "b",
				"swarmRef": map[string]any{"file": map[string]any{"name": "/b"}}},
		},
	})
	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
}
```

The template's `mode: 444` is octal digits written as a YAML number, as in `mosquitto.yaml`. Read it with `fileutil.ParseFileMode(fmt.Sprint(value))` (`"444"` gives `0o444`). Write `buildTemplateSettings` and `buildTemplateSettingsErr` over the existing build helper, returning the built settings and the error. Check the name of the invalid-block error (`grep -n "func invalidBlock" -A4 hivepaas_app/service/specservice/specserviceimpl/*.go`) and use that sentinel.

In `specmodel/buildable_test.go`, a secret and a config file with `inheritable: true` pass `CheckBuildable`, and `inheritable: "yes"` does not.

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/specservice/...`. It should FAIL.

- [ ] **Step 3: Implement.** `build_mounts.go`:

```go
package specserviceimpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const entryKeyMaxLen = 20

var notEntryKeyChars = regexp.MustCompile(`[^a-z0-9]+`)

// templateEntryKey is the key of the entry a template's swarmRef becomes: the
// setting's key, lowercased, anything else made a hyphen.
func templateEntryKey(name string) string {
	key := strings.Trim(notEntryKeyChars.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(key) > entryKeyMaxLen {
		key = strings.Trim(key[:entryKeyMaxLen], "-")
	}
	return key
}

// addTemplateMount makes the entry a template's swarmRef stands for: one file,
// of source's part, where the template put it.
func (state *buildState) addTemplateMount(
	block specmodel.Block, name string, source *entity.Setting, part string, file map[string]any, inheritable bool,
) error {
	key := templateEntryKey(name)
	for _, existing := range state.settings {
		if existing.Type == base.SettingTypeAppSettingMount && existing.Name == key {
			return invalidBlock(block, "%s: its mount would be called %q, as another's is", name, key)
		}
	}
	mountFile := &entity.AppSettingMountFile{Part: part}
	mountFile.Path, _ = file["name"].(string)
	mountFile.UID = fmt.Sprint(gofnOr(file["uid"], ""))
	mountFile.GID = fmt.Sprint(gofnOr(file["gid"], ""))
	if mode, ok := file["mode"]; ok {
		parsed, err := fileutil.ParseFileMode(fmt.Sprint(mode))
		if err != nil {
			return invalidBlock(block, "%s: mode %v is not a file mode", name, mode)
		}
		mountFile.Mode = parsed
	}
	_, err := state.addNamedSetting(base.SettingTypeAppSettingMount, key, entity.CurrentAppSettingMountVersion,
		inheritable, &entity.AppSettingMount{
			Source: entity.ObjectID{ID: source.ID}, Files: []*entity.AppSettingMountFile{mountFile},
		})
	return hperrors.Wrap(err)
}

// gofnOr is value, or or when value is absent.
func gofnOr(value, or any) any {
	if value == nil {
		return or
	}
	return value
}
```

Fold `gofnOr` into two plain `if` statements if lint prefers. `invalidBlock` returns an error already wrapped; check its signature and adapt the call.

In `build.go`, `addNamedSetting` returns `(*entity.Setting, error)`, and `addSetting` discards the setting. In `build_settings.go`, `buildSecrets`, before decoding:

```go
		entry, _ := entries[name].(map[string]any)
		swarmRef, _ := entry["swarmRef"].(map[string]any)
		file, _ := swarmRef["file"].(map[string]any)
		inheritable, _ := entry["inheritable"].(bool)
```

Then, after the key checks:

```go
		setting, err := state.addNamedSetting(base.SettingTypeSecret, secret.Key,
			entity.CurrentSecretVersion, inheritable, secret)
		if err != nil {
			return err
		}
		if file != nil {
			if err = state.addTemplateMount(block, secret.Key, setting, "value", file, inheritable); err != nil {
				return err
			}
		}
```

`decodeBlock` of an entry that still carries `swarmRef` and `inheritable`: check whether it refuses unknown fields (`grep -n "func decodeBlock" -A15 ...`). If it does, delete both keys from a copy of the entry before decoding. `buildConfigFiles` changes the same way, with part `"content"`.

In `buildable.go`, add `case "inheritable":` to both `checkSecrets` and `checkConfigFiles`, returning `unsupported(path + ".inheritable")` when the value is not a bool.

- [ ] **Step 4: Run** `go test ./hivepaas_app/service/specservice/... ./hivepaas_app/service/apptemplateservice/...`. It should PASS, including the linting of every shipped template (`app-templates` is read by `apptemplateservice` tests when present). Then run `golangci-lint run ./hivepaas_app/...`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/specservice/
git commit -m "feat(templates): a template's swarmRef is shorthand for a setting mount, and settings may be inheritable

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: HivePaaS's own apps

**Files:**
- Modify: `hivepaas_app/service/systemappservice/systemappserviceimpl/secrets.go` (+ its service's dependencies, tests)
- Modify: `hivepaas_app/service/registryservice/registryserviceimpl/apply.go` (+ `service.go`, `apply_test.go`)
- Check: `hivepaas_app/service/loggingservice/loggingserviceimpl/appdoc.go`. It builds its app through the builder with `swarmRef`, which Task 3's shorthand serves; `appdoc_test.go` should need no change.

**Interfaces:**
- Consumes: `settingmountservice.Service.Refresh`; `templateEntryKey`'s rule (repeat it here, or export it from `settingmountservice` as `EntryKeyFor(name string) string`, and use it in both places).

- [ ] **Step 1: Write the failing tests.**
  - **Registry.** In `registryserviceimpl/apply_test.go`, give the fixture a fake `settingmountservice.Service` recording `Refresh`. A change of the config or the htpasswd asserts `refreshed == []string{registryApp.ID}`; an unchanged one asserts none. Read the test's fixture first (`grep -n "clusterSecretService\|func new" .../apply_test.go`).
  - **System apps.** In `systemappserviceimpl`, if there is a test for `SyncSecrets` (`grep -rn "SyncSecrets" hivepaas_app/service/systemappservice`), change it to assert:
    - each file gets a secret setting (no `SwarmRef`);
    - each also gets an `app-setting-mount` entry, keyed `EntryKeyFor(file.Key)`, with one `value` file at `file.Path` and mode `0444`;
    - a file no longer wanted has both removed;
    - `Refresh` is called once when anything changed.

    If there is no test, write one with fakes of the setting repo, as the package's other tests do.

- [ ] **Step 2: Run** the two packages' tests. They should FAIL.

- [ ] **Step 3: Implement.**
  - **Registry.**
    - `applyConfigFile` and `applyHtpasswd` persist the changed setting with `persistSetting`, as they do, then call `s.settingMountService.Refresh(ctx, db, app)`.
    - The registry app's entries come from its template (`app.yaml.tmpl`'s `swarmRef`, Task 3).
    - The `clusterSecretService` field and parameter go.
  - **System apps.** `SyncSecrets` writes each secret without `SwarmRef` and upserts its entry beside it:
    - `Name: EntryKeyFor(key)`, `Source: {ID: setting.ID}`;
    - one file `{Part: "value", Path: file.Path, Mode: secretFileMode}`.

    A removed secret's entry is removed with it. `sameSecretFile` compares the value and the entry's path. When anything changed, it calls `s.settingMountService.Refresh(ctx, db, app)` once, after persisting. The `clusterSecretService` dependency goes.
  - **`clustersecretservice`.** If nothing uses it now (`git grep -n clustersecretservice hivepaas_app`), delete the package and its line in `registry/provides.go`.

- [ ] **Step 4: Run** `go build ./... && go test ./...`, which should PASS, and `golangci-lint run ./...`.

- [ ] **Step 5: Commit**

```bash
git add -A hivepaas_app
git status --short
git commit -m "refactor(systemapps): HivePaaS's own apps mount their files through setting mounts

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: Gates, the spec, merge

- [ ] **Step 1: Check for leftovers.** Run `git grep -n "SwarmRef\|swarmRef\|clustersecretservice" hivepaas_app`. Only the template shorthand should remain: `buildable.go`, `build_settings.go`, `build_mounts.go`, their tests, the registry template, logging's `appdoc.go`, and template-linting tests.
- [ ] **Step 2: Gates.** `go build ./...`, `golangci-lint run ./...`, `go test ./...` and `make gen-swag` should all be clean, with no diff after gen-swag.
- [ ] **Step 3: Update the spec.** In the one-way spec, note in §5 that plan 1 is built, and name any rule the plans ruled on: the entry-key derivation, and a key clash refused.
- [ ] **Step 4: Commit and merge.**

```bash
git add docs/superpowers/specs/2026-09-25-setting-mounts-one-way-design.md
git commit -m "docs(spec): setting mounts one way, as plan 1 built it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main && git merge --no-ff feat/setting-mounts-one-way -m "Merge branch 'feat/setting-mounts-one-way'"
go test ./... && git branch -d feat/setting-mounts-one-way
```

- [ ] **Step 5: Tell the user what to check on Linux,** after rebuilding the images:
  1. The registry still serves pushes and pulls. Its container has `config.json` and `htpasswd` where they were.
  2. A template app with a mounted config file (mosquitto) has its file. The app's Setting Mounts list (API) shows the entry.
  3. A secret or config file created through the API has no mount fields. Mounting one is done with an entry.
