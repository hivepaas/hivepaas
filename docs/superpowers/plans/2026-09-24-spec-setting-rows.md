# Spec Import: Setting Rows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every exported setting carries its row - name, kind, status, inheritable, default,
version, expiry, ref id - and `BuildApp` in import mode builds every app-scope setting type from
it, migrating a setting written by an older HivePaaS and refusing one from a newer.

**Architecture:** Export writes the row beside the data under a reserved key, `setting`, as it
writes `id` beside a collection entry. Import mode stops routing settings through the template
builders - those check what a template may ask for, and refuse an exported secret whose name
differs from its key - and builds every settings block, and the deployment source, with one
generic row builder: the row from `setting`, the data as `entity.Setting.Migrate` brings it to
the current version, re-encoded so secrets are encrypted at rest.

**Tech Stack:** Go 1.27, testify.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §4, §6 (`SETTING_VERSION_NEWER`), §10.

## Global Constraints

- Before calling a task done: `go build ./...`; `golangci-lint run ./...` over the whole repo,
  committing only when it prints `0 issues.`; `go test ./hivepaas_app/service/specservice/...`
  whole. The last task runs `go test ./...`.
- The row travels as `setting: {name, kind, status, inheritable, default, version, expireAt, refId}`,
  every field `omitempty`, in every setting body export writes: singletons, collection entries,
  `deployment.source`. No exported setting type may have a top-level data field named `setting`
  in any case - a test holds it, as it holds `id`.
- Templates are untouched: `CheckBuildable` refuses a `setting` key, and the template builders
  run as before.
- In import mode a setting's version comes from its row; `Migrate` brings an older one to the
  current version and refuses a newer one with `ErrDataVerNewerThanSystemVer`. A row without a
  version - a bundle from before this plan - is migrated from version 0.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: Export writes each setting's row

**Files:** `specmodel/document.go` (`SettingMetaKey`, `SettingMeta`); `specserviceimpl/assemble.go`
(`assembleSettings`); test `specserviceimpl/export_ids_test.go`.

- [ ] **Step 1: Failing tests.** Append to `export_ids_test.go`:

```go
// A setting is more than its data: the row says what it is called, what kind it
// is, whether it is on, whether the scopes below see it, whether it is the
// default of its kind, and which version of its data it holds.
func TestExportWritesTheRowOfEverySetting(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	global := &specmodel.GlobalDoc{}
	readDoc(t, path, "global.yaml", global)
	cert := entry(t, global.Settings, "sslCerts", "localhost")
	assert.Equal(t, map[string]any{"name": "localhost", "kind": "self-signed", "status": "active", "version": 1},
		cert[specmodel.SettingMetaKey])

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing, ok := env.Apps["backend"].Settings["routing"].(map[string]any)
	if assert.True(t, ok) {
		assert.Equal(t, map[string]any{"status": "active", "version": 1}, routing[specmodel.SettingMetaKey],
			"a singleton has no name, and the row still says the rest")
	}
}

func TestNoExportedSettingTypeHasATopLevelSettingField(t *testing.T) {
	for _, typ := range entity.AllParsedSettingTypes() {
		decision := entity.SpecExportDecision(&entity.Setting{
			Type: typ, Scope: base.ObjectScopeProject, ObjectID: "prj_1",
		})
		if !decision.Export {
			continue
		}
		data, err := (&entity.Setting{Type: typ}).Parse()
		if !assert.NoError(t, err, "%v", typ) {
			continue
		}
		for _, key := range jsonKeys(reflect.TypeOf(data)) {
			assert.False(t, strings.EqualFold(key, specmodel.SettingMetaKey),
				"%v has a top-level %q, which the row export writes would overwrite", typ, key)
		}
	}
}
```

  The fixture's settings set no `Version`; give `cert` and `routing` `Version: 1` in
  `exportFixture` (`export_test.go`) so the row has one to write.

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run 'TestExportWritesTheRow|TestNoExportedSettingType' -v` -
  expect `undefined: specmodel.SettingMetaKey`.

- [ ] **Step 3: Implement.** In `specmodel/document.go`, after `CollectionEntryIDKey`:

```go
// SettingMetaKey is where export writes a setting's row, beside its data: the
// name and kind a collection is keyed by, and the flags the data does not carry.
// No exported setting type may have a top-level field of that name -
// TestNoExportedSettingTypeHasATopLevelSettingField holds that.
const SettingMetaKey = "setting"

// SettingMeta is the row of a setting, as a bundle carries it.
type SettingMeta struct {
	Name        string    `yaml:"name,omitempty"        json:"name,omitempty"`
	Kind        string    `yaml:"kind,omitempty"        json:"kind,omitempty"`
	Status      string    `yaml:"status,omitempty"      json:"status,omitempty"`
	Inheritable bool      `yaml:"inheritable,omitempty" json:"inheritable,omitempty"`
	Default     bool      `yaml:"default,omitempty"     json:"default,omitempty"`
	Version     int       `yaml:"version,omitempty"     json:"version,omitempty"`
	ExpireAt    time.Time `yaml:"expireAt,omitzero"     json:"expireAt,omitzero"`
	RefID       string    `yaml:"refId,omitempty"       json:"refId,omitempty"`
}
```

  (import `"time"`). In `specserviceimpl/assemble.go`, after each `renderSetting` call in
  `assembleSettings` - the singleton branch and the collection branch - write the row:

```go
			body[specmodel.SettingMetaKey] = settingMeta(setting)
```

  (the singleton branch's setting is `group[0]`), and add:

```go
// settingMeta is a setting's row as a bundle carries it. It is a map rather than
// the struct so that it sits in the body like the data beside it, and so that
// only what is set is written.
func settingMeta(setting *entity.Setting) map[string]any {
	meta := map[string]any{"status": string(setting.Status), "version": setting.Version}
	for key, value := range map[string]string{
		"name": setting.Name, "kind": setting.Kind, "refId": setting.RefID,
	} {
		if value != "" {
			meta[key] = value
		}
	}
	if setting.Inheritable {
		meta["inheritable"] = true
	}
	if setting.Default {
		meta["default"] = true
	}
	if !setting.ExpireAt.IsZero() {
		meta["expireAt"] = setting.ExpireAt.UTC().Format(time.RFC3339)
	}
	if setting.Version == 0 {
		delete(meta, "version")
	}
	return meta
}
```

- [ ] **Step 4: Run** the whole spec package - expect PASS, determinism included.
- [ ] **Step 5: Lint, commit** `feat(spec): write each setting's row beside its data`.

---

### Task 2: Import mode builds every setting from its row

**Files:** `specmodel/buildable.go` (`BlockSettings`, `ImportOnlyBlocks`, `ImportBlocks`);
`specmodel/importable_test.go` (`TestImportBlocks`); `specserviceimpl/build_settings.go`
(`buildImportedSettings`, `addImportedSetting`); `specserviceimpl/build_deployment.go`
(`buildSource`); `specserviceimpl/build.go` (registry); tests
`specserviceimpl/build_import_test.go`.

- [ ] **Step 1: Failing tests.** In `specmodel/importable_test.go`, `TestImportBlocks` becomes:

```go
func TestImportBlocks(t *testing.T) {
	assert.Equal(t, []Block{
		BlockDeploymentStorage, BlockContainer, BlockDeploymentResources, BlockDeploymentNetworks,
		BlockDeploymentService, BlockSettings,
	}, ImportBlocks(decodeDoc(t, "deployment:\n  container: {}\nsettings:\n  kind: {category: webapp}\n")))
	assert.Equal(t, []Block{BlockSettings},
		ImportBlocks(decodeDoc(t, "settings:\n  routing: {port: 80}\n")), "an app never deployed")
	assert.Empty(t, ImportBlocks(&AppDoc{}))
}
```

  Append to `specserviceimpl/build_import_test.go`:

```go
// A setting export wrote comes back as the row it was, whatever its type: the
// name and kind a collection is keyed by, its flags, and its data at the
// version this installation reads.
func TestBuildAppImportsEverySettingFromItsRow(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, `
settings:
  features: {setting: {status: disabled, version: 1}}
  schedJobs:
    nightly@backup:
      setting: {name: nightly, kind: backup, status: active, version: 1}
      id: 01JOLDID
  secrets:
    db-password: {setting: {name: db-password, version: 1}, key: DB_PASSWORD, value: hunter2}
`)
	req.Import = true

	resp, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	rows := map[base.SettingType]*entity.Setting{}
	for _, row := range resp.Settings {
		rows[row.Type] = row
	}
	if job := rows[base.SettingTypeSchedJob]; assert.NotNil(t, job) {
		assert.Equal(t, "nightly", job.Name)
		assert.Equal(t, "backup", job.Kind)
		assert.NotEqual(t, "01JOLDID", job.ID, "a new row, which import matches to an existing one itself")
	}
	if features := rows[base.SettingTypeAppFeatures]; assert.NotNil(t, features) {
		assert.Equal(t, base.SettingStatusDisabled, features.Status)
	}
	if secret := rows[base.SettingTypeSecret]; assert.NotNil(t, secret) {
		assert.Equal(t, "db-password", secret.Name, "a secret's name need not be its key")
		assert.NotContains(t, secret.Data, "hunter2", "the value is encrypted at rest")
	}
}

// A setting from a newer HivePaaS cannot be read backwards.
func TestBuildAppRefusesASettingFromANewerVersion(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, "settings:\n  features: {setting: {version: 999}}\n")
	req.Import = true

	_, err := svc.BuildApp(context.Background(), nil, req)

	assert.ErrorIs(t, err, hperrors.ErrDataVerNewerThanSystemVer)
}
```

  (imports gain `"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"`.)

- [ ] **Step 2: Run** the whole spec package - expect `undefined: BlockSettings`.

- [ ] **Step 3: Implement.** In `specmodel/buildable.go`, add `BlockSettings Block = "settings"` to the
  import-only constants and to `ImportOnlyBlocks` and `buildOrder` (last), and in `ImportBlocks`
  replace `blocks = append(blocks, presentSettingsBlocks(doc)...)` with:

```go
	if len(doc.Settings) > 0 {
		blocks = append(blocks, BlockSettings)
	}
```

  and reword its comment: settings are built as one block in import mode, from their rows.

  In `specserviceimpl/build.go`, the registry gains `specmodel.BlockSettings: s.buildImportedSettings,`.

  In `specserviceimpl/build_settings.go`, replace `entryBody` with:

```go
// buildImportedSettings builds every settings block of an exported document. The
// template builders are not used: they check what a template may ask for, and
// an export is what an installation already held.
func (s *service) buildImportedSettings(_ context.Context, state *buildState) error {
	for _, name := range slices.Sorted(maps.Keys(state.req.Doc.Settings)) {
		body := state.req.Doc.Settings[name]
		if typ, ok := specmodel.SingletonTypeOf(name); ok {
			if err := state.addImportedSetting(typ, "", body); err != nil {
				return err
			}
			continue
		}
		typ, ok := specmodel.CollectionTypeOf(name)
		if !ok {
			return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("settings.%s", name)
		}
		entries, _ := body.(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(entries)) {
			if err := state.addImportedSetting(typ, key, entries[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

// addImportedSetting builds one setting of an exported document: the row from
// what export wrote beside the data, and the data brought to the version this
// installation reads - refused when it is newer. key is a collection entry's
// key, the name when the row names none.
func (state *buildState) addImportedSetting(typ base.SettingType, key string, body any) error {
	block := specmodel.Block("settings." + string(typ))
	fields, ok := body.(map[string]any)
	if !ok {
		return invalidBlock(block, "%s is not a mapping", key)
	}
	fields = maps.Clone(fields)
	meta := specmodel.SettingMeta{}
	if raw, found := fields[specmodel.SettingMetaKey]; found {
		if err := decodeBlock(block, raw, &meta); err != nil {
			return err
		}
	}
	delete(fields, specmodel.SettingMetaKey)
	delete(fields, specmodel.CollectionEntryIDKey)
	data, err := json.Marshal(fields)
	if err != nil {
		return invalidBlock(block, "%s", err.Error())
	}

	setting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		ObjectID:    state.req.App.ID,
		RefID:       meta.RefID,
		Type:        typ,
		Kind:        meta.Kind,
		Status:      gofn.Coalesce(base.SettingStatus(meta.Status), base.SettingStatusActive),
		Name:        gofn.Coalesce(meta.Name, key),
		Data:        string(data),
		Inheritable: meta.Inheritable,
		Default:     meta.Default,
		Version:     meta.Version,
		UpdateVer:   1,
		CreatedAt:   state.req.TimeNow,
		UpdatedAt:   state.req.TimeNow,
		ExpireAt:    meta.ExpireAt,
	}
	if _, err = setting.Migrate(); err != nil {
		return hperrors.Wrap(err)
	}
	// Written again through the entity, so a secret that travelled in the clear is
	// encrypted at rest like any other.
	parsed, err := setting.Parse()
	if err != nil {
		return invalidBlock(block, "%s: %s", key, err.Error())
	}
	if err = setting.SetData(parsed); err != nil {
		return hperrors.Wrap(err)
	}
	state.settings = append(state.settings, setting)
	return nil
}
```

  In `specmodel/singleton.go` add the reverse lookups:

```go
// SingletonTypeOf is the singleton type a block name belongs to.
func SingletonTypeOf(block string) (base.SettingType, bool) {
	for typ, name := range singletonBlockNames {
		if name == block {
			return typ, true
		}
	}
	return "", false
}

// CollectionTypeOf is the collection type a block name belongs to.
func CollectionTypeOf(block string) (base.SettingType, bool) {
	for typ, name := range collectionBlockNames {
		if name == block {
			return typ, true
		}
	}
	return "", false
}
```

  In `build_deployment.go`, `buildSource` in import mode builds the app-deployment setting from its
  row:

```go
func (s *service) buildSource(_ context.Context, state *buildState) error {
	if state.req.Import {
		return state.addImportedSetting(base.SettingTypeAppDeployment, "", map[string]any(state.req.Doc.Deployment.Source))
	}
```

  and the template check keeps its original form (drop the `!state.req.Import &&` added by the
  builders plan). Remove the calls to `entryBody` from `buildSecrets` and `buildConfigFiles`.

- [ ] **Step 4: Run** the whole spec package - expect PASS, `TestBuildAppImportsAnExportedDocument`
  included.
- [ ] **Step 5: Lint, commit** `feat(spec): build every imported setting from its row`.

---

### Task 3: Record, verify

- [ ] **Step 1:** In the import spec, correction 1 gains a sentence: export also writes each
  setting's row under `setting`, which the export spec's `is_default` required and the code did
  not do. §14 item 3 names this plan for the app-scope setting types.
- [ ] **Step 2:** `go build ./... && golangci-lint run ./... && go test ./...`.
- [ ] **Step 3:** Commit `docs(spec): record the setting row`, and report.
