# Setting Mounts - Permissions and Surfaces Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** People can create, edit, disable and delete setting mounts through the API behind §7's gate. A path is never claimed twice between entries, secrets and config files. Entries travel through export and import, and clones and previews leave them behind.

**Architecture:**
- **Pure rules** live in `service/settingmountservice/checks.go` and are unit-tested without a database:
  - what an entry must be;
  - which (source, sensitive part) grants it hands out;
  - whether an update widens them;
  - whether a path is free.
- **The engine** gains two reads:
  - `ClaimedPaths`: the paths an app's secrets, config files and entries hold;
  - `EntryStates`: why an entry is not mounted, for the screens of plan 4.
- **The API.** A new settings use case, `settingmountuc`, follows `configfileuc`. It is served by the generic settings handlers under `/apps/{appID}/setting-mounts`, plus a `sources` endpoint that lists the parts registry.
- **The gate:**
  - `AuthorizeSecretReveal` runs on `uc.DB`, outside the save's transaction, so a denial is recorded even though the save rolls back;
  - import asks a non-recording `MayRevealSecrets` at validate and the recording gate at apply.
- **Both directions.** Secrets and config files check paths against entries. `clustersecretservice` lets an ordinary file take its path back from a mount, as the engine already does the other way.

**Tech Stack:** Go, gin, bun, testify.

**Spec:** `docs/superpowers/specs/2026-09-25-setting-mounts-design.md` §1, §6, §7, §9, §13 plan 2. Plan 1 (`docs/superpowers/plans/2026-09-25-setting-mounts-engine.md`) is merged.

## Global Constraints

- **Entry key** (= `Setting.Name`): `^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$`, never `tls` (`settingmountservice.ValidEntryKey`).
- **Source types and parts:** `settingmountservice.PartsOf`. Sensitive parts are `privateKey`, `password` and `htpasswd`, whatever the source type.
- **Paths** (`settingmountservice.ValidPath`):
  - absolute and clean;
  - not `/`;
  - not under `/run/secrets/tls`;
  - unique among the app's active entries, secrets and config files.
- **Gate (§7):**
  - `AuthorizeSecretReveal` is asked only when the set of (source, sensitive part) pairs an app mounts grows;
  - removing a part, disabling an entry or deleting one needs Write on the app only;
  - enabling a disabled entry that has sensitive parts is growth.
- **Import (§9):**
  - an entry that grows the set for a caller who may not reveal is skipped, with issue code `SETTING_MOUNT_NOT_PERMITTED` (severity skipped);
  - the rest of the app is imported.
- **Routes:** `GET|POST /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts`, `GET|PUT|DELETE .../setting-mounts/{itemID}`, `PUT .../setting-mounts/{itemID}/status`, `GET .../setting-mounts/sources`.
- **Wire format (the dashboard's in plan 4):**
  - an entry is `{id, name, status, updateVer, ..., source: BaseSettingResp, files: [{part, path, uid, gid, mode, sensitive}], state: {reason, mounted: [path]}}`;
  - create and update take `{name, source: {id}, files: [{part, path, uid, gid, mode}], updateVer}`;
  - `sources` returns `{data: [{type, parts: [{name, required, sensitive}]}], mayMountSensitive}`.
- **Gates:** `go build ./...`, `golangci-lint run ./...` (whole repo), `go test ./...`, `make gen-swag` (DTOs change; commit `docs/openapi/swagger.json`).
- **Git:**
  - branch `feat/setting-mounts-surfaces` off `main`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, delete the branch, do not push;
  - stage only files this plan names.

## Review Focus

1. **An update that keeps its grants asks nothing.** Renaming an entry, moving a path or changing a mode keeps its (source, part) pairs, and must not ask for Reveal Secrets. *Test: Task 1 `TestWidensOnlyByNewPairs`.*
2. **Enabling a disabled entry with a private key asks.** Its pairs were not mounted while it was disabled. *Test: Task 3 `TestStatusGrantsCountOnlyWhenActive`.*
3. **A secret and an entry naming one file differently still collide.** `db_password` and `/run/secrets/db_password` are the same file. *Tests: Task 4 `TestASecretNamedRelativelyCollidesWithAnEntry`, Task 5 `TestMakeRoomComparesResolvedTargets`.*
4. **Re-importing an unchanged app asks nothing and writes nothing**, although its entry has a sensitive part. *Test: Task 6 `TestPlanDoesNotAskForAnEntryThatGrantsNothingNew`.*
5. **A refused entry is not written, even for an app being created.** A created node writes every setting it holds. *Test: Task 6 `TestApplyLeavesARefusedEntryOut`.*

---

### Task 1: The rules, `ClaimedPaths`, `EntryStates`, `MayRevealSecrets`

**Files:**
- Create: `hivepaas_app/service/settingmountservice/checks.go`, test `checks_test.go`
- Modify: `hivepaas_app/service/settingmountservice/service.go` (two methods, `EntryState`, reason constants)
- Modify: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/resolve.go` (reasons), `service.go` (seam `loadClaimants`)
- Create: `settingmountserviceimpl/claimed.go`, test `claimed_test.go`; add a test to `resolve_test.go`
- Modify: `hivepaas_app/hperrors/errors_settings.go`, `hivepaas_app/pkg/translation/messages/en/errors.settings.en.toml`
- Modify: `hivepaas_app/permission/manager.go`, `hivepaas_app/permission/permissionimpl/secret_reveal.go`, test `secret_reveal_test.go`

**Interfaces:**
- Produces, in package `settingmountservice`:
  - `SensitivePart(name string) bool`;
  - `type Grant struct{ Source string \`json:"source"\`; Part string \`json:"part"\` }`;
  - `Grants(mount *entity.AppSettingMount) []Grant`, sorted and unique;
  - `Widens(before, after []Grant) []Grant`: the pairs of `after` missing from `before`;
  - `CheckEntry(key string, mount *entity.AppSettingMount, sourceType base.SettingType) error`;
  - `CheckPathsFree(claimed map[string]string, paths ...string) error`;
  - `SecretFileTarget(secret *entity.Secret) string` and `ConfigFileTarget(configFile *entity.ConfigFile) string`, each `""` for a setting that mounts no file;
  - reason constants `ReasonEntryDisabled`, `ReasonKeyInvalid`, `ReasonSourceUnavailable`, `ReasonSourceIncomplete`, `ReasonPathsTaken`;
  - `type EntryState struct{ Reason string \`json:"reason,omitempty"\`; Mounted []string \`json:"mounted"\` }`.
- Produces, on `Service`:
  - `ClaimedPaths(ctx, db, appID, exceptSettingID string) (map[string]string, error)`, mapping a path to its owner: `"secret DB_PASSWORD"`, `"config file app.conf"` or `"setting mount cert"`;
  - `EntryStates(ctx, db, app *entity.App) (map[string]*EntryState, error)`, keyed by the entry's setting id.
- Produces: `permission.Manager.MayRevealSecrets(ctx, db, auth) (bool, error)`.
- Produces errors, all with code-named params:
  - `ErrSettingMountKeyInvalid` (Name);
  - `ErrSettingMountNoFiles`;
  - `ErrSettingMountSourceUnsupported` (Type);
  - `ErrSettingMountPartInvalid` (Part, Type);
  - `ErrSettingMountPathInvalid` (Path);
  - `ErrSettingMountPathTaken` (Path, Owner).

- [ ] **Step 1: Branch.** `git checkout main && git checkout -b feat/setting-mounts-surfaces`

- [ ] **Step 2: Write the failing tests.** `hivepaas_app/service/settingmountservice/checks_test.go`:

```go
package settingmountservice

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func mountOf(source string, parts ...string) *entity.AppSettingMount {
	m := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		m.Files = append(m.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/app/" + part})
	}
	return m
}

func TestSensitiveByNameWhateverTheType(t *testing.T) {
	for name, want := range map[string]bool{
		"privateKey": true, "password": true, "htpasswd": true,
		"certificate": false, "caCertificate": false, "publicKey": false, "username": false, "other": false,
	} {
		assert.Equal(t, want, SensitivePart(name), name)
	}
}

func TestGrantsAreTheSensitivePairsOnce(t *testing.T) {
	assert.Equal(t, []Grant{{Source: "cert_1", Part: "privateKey"}},
		Grants(mountOf("cert_1", "certificate", "privateKey", "privateKey")))
	assert.Empty(t, Grants(mountOf("cert_1", "certificate")))
	assert.Empty(t, Grants(nil))
}

// Renaming an entry, moving a path or changing a mode keeps its pairs: nothing
// new is handed out, so nothing is asked.
func TestWidensOnlyByNewPairs(t *testing.T) {
	before := Grants(mountOf("cert_1", "certificate", "privateKey"))

	assert.Empty(t, Widens(before, Grants(mountOf("cert_1", "privateKey"))), "a part removed")
	assert.Empty(t, Widens(before, before), "unchanged")
	assert.Equal(t, []Grant{{Source: "cert_2", Part: "privateKey"}},
		Widens(before, Grants(mountOf("cert_2", "privateKey"))), "another source")
	assert.Equal(t, []Grant{{Source: "auth_1", Part: "htpasswd"}},
		Widens(nil, Grants(mountOf("auth_1", "username", "htpasswd"))), "a new entry")
}

func TestCheckEntry(t *testing.T) {
	valid := mountOf("cert_1", "certificate", "privateKey")
	assert.NoError(t, CheckEntry("cert", valid, base.SettingTypeSSLCert))

	dup := mountOf("cert_1", "certificate", "certificate")
	samePath := mountOf("cert_1", "certificate", "privateKey")
	samePath.Files[1].Path = samePath.Files[0].Path
	underTLS := mountOf("cert_1", "certificate")
	underTLS.Files[0].Path = "/run/secrets/tls/cert.pem"
	for name, tc := range map[string]struct {
		key   string
		mount *entity.AppSettingMount
		typ   base.SettingType
		want  error
	}{
		"reserved key":       {"tls", valid, base.SettingTypeSSLCert, hperrors.ErrSettingMountKeyInvalid},
		"upper case key":     {"Cert", valid, base.SettingTypeSSLCert, hperrors.ErrSettingMountKeyInvalid},
		"no files":           {"cert", mountOf("cert_1"), base.SettingTypeSSLCert, hperrors.ErrSettingMountNoFiles},
		"not a source":       {"cert", valid, base.SettingTypeSecret, hperrors.ErrSettingMountSourceUnsupported},
		"part of other type": {"cert", mountOf("cert_1", "htpasswd"), base.SettingTypeSSLCert, hperrors.ErrSettingMountPartInvalid},
		"part twice":         {"cert", dup, base.SettingTypeSSLCert, hperrors.ErrSettingMountPartInvalid},
		"path twice":         {"cert", samePath, base.SettingTypeSSLCert, hperrors.ErrSettingMountPathInvalid},
		"path under tls":     {"cert", underTLS, base.SettingTypeSSLCert, hperrors.ErrSettingMountPathInvalid},
	} {
		err := CheckEntry(tc.key, tc.mount, tc.typ)
		assert.True(t, errors.Is(err, tc.want), "%s: %v", name, err)
	}
}

func TestCheckPathsFree(t *testing.T) {
	claimed := map[string]string{"/run/secrets/db_password": "secret DB_PASSWORD"}
	assert.NoError(t, CheckPathsFree(claimed, "/etc/app/key.pem"))
	assert.True(t, errors.Is(CheckPathsFree(claimed, "/run/secrets/db_password"), hperrors.ErrSettingMountPathTaken))
}

func TestFileTargetsOfSecretsAndConfigFiles(t *testing.T) {
	assert.Equal(t, "/run/secrets/db_password", SecretFileTarget(&entity.Secret{
		SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "db_password"}}}))
	assert.Empty(t, SecretFileTarget(&entity.Secret{}), "read through the environment, no file")
	assert.Equal(t, "/etc/app.conf", ConfigFileTarget(&entity.ConfigFile{
		SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app.conf"}}}))
}
```

`settingmountserviceimpl/claimed_test.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func claimant(t *testing.T, id string, typ base.SettingType, name string, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: typ, Name: name, ObjectID: testApp.ID, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func TestClaimedPathsAreEveryFileOfTheApp(t *testing.T) {
	useDataKey(t)
	svc := &service{}
	claimants := []*entity.Setting{
		claimant(t, "s1", base.SettingTypeSecret, "DB_PASSWORD", &entity.Secret{Key: "DB_PASSWORD",
			Value:    entity.NewEncryptedField("x"),
			SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "db_password"}}}),
		claimant(t, "s2", base.SettingTypeSecret, "API_TOKEN", &entity.Secret{Key: "API_TOKEN",
			Value: entity.NewEncryptedField("y")}),
		claimant(t, "c1", base.SettingTypeConfigFile, "app.conf", &entity.ConfigFile{Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "app.conf"}}}),
		claimant(t, "m1", base.SettingTypeAppSettingMount, "cert", certFiles("cert_1")),
	}
	svc.loadClaimants = func(context.Context, database.IDB, string) ([]*entity.Setting, error) {
		return claimants, nil
	}

	claimed, err := svc.ClaimedPaths(context.Background(), nil, testApp.ID, "m1")

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"/run/secrets/db_password": "secret DB_PASSWORD",
		"/app.conf":                "config file app.conf",
	}, claimed, "the entry being saved is left out, and a secret without a file claims nothing")
}
```

Add to `resolve_test.go`:

```go
// Why an entry is not in the container is what the screen says.
func TestEntryStatesSayWhyNothingIsMounted(t *testing.T) {
	mounted := entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))
	disabled := entry(t, "b", base.SettingStatusDisabled, certFiles("cert_1"))
	missing := entry(t, "c", base.SettingStatusActive, certFiles("cert_gone"))
	incomplete := entry(t, "d", base.SettingStatusActive, certFiles("cert_2"))
	shadowed := entry(t, "e", base.SettingStatusActive, certFiles("cert_1"))
	svc := fixture(t, []*entity.Setting{mounted, disabled, missing, incomplete, shadowed},
		certSource(t, "cert_1", "CERT", "KEY"), certSource(t, "cert_2", "", ""))

	states, err := svc.EntryStates(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, &settingmountservice.EntryState{Mounted: []string{"/etc/app/tls/cert.pem", "/etc/app/tls/key.pem"}},
		states[mounted.ID])
	assert.Equal(t, settingmountservice.ReasonEntryDisabled, states[disabled.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonSourceUnavailable, states[missing.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonSourceIncomplete, states[incomplete.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonPathsTaken, states[shadowed.ID].Reason)
}
```

`permissionimpl/secret_reveal_test.go`, appended:

```go
// Import's validate asks without recording: nothing is revealed until apply.
func TestMayRevealSecretsAnswersWithoutRecording(t *testing.T) {
	enableReveal(t, true)
	mgr, audit := newRevealManager(nil)

	allowed, err := mgr.MayRevealSecrets(context.Background(), nil, adminAuth())
	assert.NoError(t, err)
	assert.True(t, allowed)
	allowed, err = mgr.MayRevealSecrets(context.Background(), nil, plainAuth())
	assert.NoError(t, err)
	assert.False(t, allowed, "a denial is an answer, not an error")

	enableReveal(t, false)
	allowed, err = mgr.MayRevealSecrets(context.Background(), nil, adminAuth())
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Empty(t, audit.entries)
}
```

That test file uses `t.Fatalf`, not `assert`. Add the `github.com/stretchr/testify/assert` import, or write the checks with `if ... { t.Fatalf(...) }` to match the file.

- [ ] **Step 3: Run** `go test ./hivepaas_app/service/settingmountservice/... ./hivepaas_app/permission/...`. It should FAIL to compile.

- [ ] **Step 4: Errors.** In `hperrors/errors_settings.go`, add a block:

```go
// Errors for setting mounts
var (
	ErrSettingMountKeyInvalid        = NewErr(ErrArgumentInvalid, "ERR_SETTING_MOUNT_KEY_INVALID")
	ErrSettingMountNoFiles           = NewErr(ErrArgumentInvalid, "ERR_SETTING_MOUNT_NO_FILES")
	ErrSettingMountSourceUnsupported = NewErr(ErrArgumentInvalid, "ERR_SETTING_MOUNT_SOURCE_UNSUPPORTED")
	ErrSettingMountPartInvalid       = NewErr(ErrArgumentInvalid, "ERR_SETTING_MOUNT_PART_INVALID")
	ErrSettingMountPathInvalid       = NewErr(ErrArgumentInvalid, "ERR_SETTING_MOUNT_PATH_INVALID")
	// ErrSettingMountPathTaken refuses a file at a path another file of the app
	// has: Docker would refuse the service, or one file would hide the other.
	ErrSettingMountPathTaken = NewErr(ErrConflict, "ERR_SETTING_MOUNT_PATH_TAKEN")
)
```

In `errors.settings.en.toml`:

```toml
ERR_SETTING_MOUNT_KEY_INVALID = "'{{.Name}}' cannot name a setting mount: use lowercase letters, digits and hyphens, at most 20, and not 'tls'"
ERR_SETTING_MOUNT_NO_FILES = "A setting mount needs at least one file"
ERR_SETTING_MOUNT_SOURCE_UNSUPPORTED = "Settings of type '{{.Type}}' cannot be mounted"
ERR_SETTING_MOUNT_PART_INVALID = "'{{.Part}}' is not a part a '{{.Type}}' setting offers, or it is listed twice"
ERR_SETTING_MOUNT_PATH_INVALID = "'{{.Path}}' cannot hold a mounted file: use an absolute path, once, outside /run/secrets/tls"
ERR_SETTING_MOUNT_PATH_TAKEN = "'{{.Path}}' is already used by the app's {{.Owner}}"
```

- [ ] **Step 5: `checks.go`:**

```go
package settingmountservice

import (
	"cmp"
	"path"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// SensitivePart reports whether a part of that name is sensitive in any source
// type. Import reads an entry before it knows its source's type, and the
// registry names parts the same way throughout, so the name is enough.
func SensitivePart(name string) bool {
	for _, typ := range SourceTypes() {
		if part := PartOf(typ, name); part != nil && part.Sensitive {
			return true
		}
	}
	return false
}

// Grant is one sensitive part of one source that an entry hands to its app:
// what §7's gate is about.
type Grant struct {
	Source string `json:"source"`
	Part   string `json:"part"`
}

// Grants are the sensitive pairs an entry hands out, sorted, each once.
func Grants(mount *entity.AppSettingMount) []Grant {
	if mount == nil {
		return nil
	}
	var grants []Grant
	for _, f := range mount.Files {
		if f != nil && SensitivePart(f.Part) {
			grants = append(grants, Grant{Source: mount.Source.ID, Part: f.Part})
		}
	}
	slices.SortFunc(grants, func(a, b Grant) int {
		return cmp.Or(strings.Compare(a.Source, b.Source), strings.Compare(a.Part, b.Part))
	})
	return slices.Compact(grants)
}

// Widens is what after hands out that before did not: the pairs that take the
// Reveal Secrets permission.
func Widens(before, after []Grant) []Grant {
	var added []Grant
	for _, grant := range after {
		if !slices.Contains(before, grant) {
			added = append(added, grant)
		}
	}
	return added
}

// CheckEntry refuses an entry wrong in itself, or for its source's type.
func CheckEntry(key string, mount *entity.AppSettingMount, sourceType base.SettingType) error {
	if !ValidEntryKey(key) {
		return hperrors.Wrap(hperrors.ErrSettingMountKeyInvalid).WithParam("Name", key)
	}
	if mount == nil || len(mount.Files) == 0 {
		return hperrors.Wrap(hperrors.ErrSettingMountNoFiles)
	}
	if !IsSourceType(sourceType) {
		return hperrors.Wrap(hperrors.ErrSettingMountSourceUnsupported).WithParam("Type", string(sourceType))
	}
	parts, paths := map[string]bool{}, map[string]bool{}
	for _, f := range mount.Files {
		if f == nil || PartOf(sourceType, f.Part) == nil || parts[f.Part] {
			part := ""
			if f != nil {
				part = f.Part
			}
			return hperrors.Wrap(hperrors.ErrSettingMountPartInvalid).
				WithParam("Part", part).WithParam("Type", string(sourceType))
		}
		if !ValidPath(f.Path) || paths[f.Path] {
			return hperrors.Wrap(hperrors.ErrSettingMountPathInvalid).WithParam("Path", f.Path)
		}
		parts[f.Part], paths[f.Path] = true, true
	}
	return nil
}

// CheckPathsFree refuses a path another file of the app already has. claimed is
// ClaimedPaths: a path, and what has it.
func CheckPathsFree(claimed map[string]string, paths ...string) error {
	for _, p := range paths {
		if owner, taken := claimed[path.Clean(p)]; taken {
			return hperrors.Wrap(hperrors.ErrSettingMountPathTaken).WithParam("Path", p).WithParam("Owner", owner)
		}
	}
	return nil
}

// SecretFileTarget is where a secret's file lands in the app's containers, empty
// for a secret read through the environment only.
func SecretFileTarget(secret *entity.Secret) string {
	if secret == nil || secret.SwarmRef == nil || secret.SwarmRef.File == nil || secret.SwarmRef.File.Name == "" {
		return ""
	}
	return SecretTarget(secret.SwarmRef.File.Name)
}

// ConfigFileTarget is SecretFileTarget for a config file.
func ConfigFileTarget(configFile *entity.ConfigFile) string {
	if configFile == nil || configFile.SwarmRef == nil || configFile.SwarmRef.File == nil ||
		configFile.SwarmRef.File.Name == "" {
		return ""
	}
	return ConfigTarget(configFile.SwarmRef.File.Name)
}
```

Check `hperrors` `WithParam` chaining: it returns `HPError`, and `Wrap` returns `HPError`, as `NewNotFound` shows.

- [ ] **Step 6: The interface.** In `service.go`, add to `Service`:

```go
	// ClaimedPaths are the paths the app's active secrets, config files and
	// entries give files to, each with what gives it, leaving out the setting
	// exceptSettingID - the one being saved.
	ClaimedPaths(ctx context.Context, db database.IDB, appID, exceptSettingID string) (map[string]string, error)
	// EntryStates says, for each of the app's entries by setting id, what of it
	// is mounted, and why nothing is when nothing is.
	EntryStates(ctx context.Context, db database.IDB, app *entity.App) (map[string]*EntryState, error)
```

And after `File`:

```go
// Why nothing of an entry is mounted.
const (
	ReasonEntryDisabled     = "entry-disabled"
	ReasonKeyInvalid        = "key-invalid"
	ReasonSourceUnavailable = "source-unavailable" // missing, disabled, or not visible from the app
	ReasonSourceIncomplete  = "source-incomplete"  // a required part is empty: a certificate not obtained yet
	ReasonPathsTaken        = "paths-taken"        // every path is another entry's
)

// EntryState is what of an entry is mounted.
type EntryState struct {
	// Reason is why nothing is mounted; empty when something is.
	Reason  string   `json:"reason,omitempty"`
	Mounted []string `json:"mounted"`
}
```

- [ ] **Step 7: Reasons in `resolve.go`.** Change `Resolve` into a thin wrapper over `resolve`, which also returns the states:

```go
func (s *service) Resolve(ctx context.Context, db database.IDB, app *entity.App) ([]*settingmountservice.File, error) {
	files, _, err := s.resolve(ctx, db, app)
	return files, err
}

func (s *service) EntryStates(
	ctx context.Context, db database.IDB, app *entity.App,
) (map[string]*settingmountservice.EntryState, error) {
	_, states, err := s.resolve(ctx, db, app)
	return states, err
}
```

In `resolve`:
- Every entry the loader returns gets a state.
- A disabled entry gets `ReasonEntryDisabled` and an invalid key `ReasonKeyInvalid`. Both are left out of the rest.
- `entryFiles` returns `(files, reason, err)`:
  - `ReasonSourceUnavailable` for a missing, inactive or non-source setting;
  - `ReasonSourceIncomplete` when `!Usable`.
- After claiming, an entry keeps `Mounted`, the paths it got. One that had files and got none has `ReasonPathsTaken`.
- `Mounted` is never nil: `[]string{}` when empty, so the JSON is `[]`.

- [ ] **Step 8: `claimed.go`**, and the seam. In `service.go`, add the field `loadClaimants func(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error)` and set it in `New` to `s.loadClaimantsFromRepo`:

```go
package settingmountserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

func (s *service) ClaimedPaths(
	ctx context.Context, db database.IDB, appID, exceptSettingID string,
) (map[string]string, error) {
	settings, err := s.loadClaimants(ctx, db, appID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	claimed := map[string]string{}
	for _, setting := range settings {
		if setting.ID == exceptSettingID || setting.Status != base.SettingStatusActive {
			continue
		}
		switch setting.Type { //nolint:exhaustive
		case base.SettingTypeSecret:
			secret, err := setting.AsSecret()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if target := settingmountservice.SecretFileTarget(secret); target != "" {
				claimed[target] = "secret " + setting.Name
			}
		case base.SettingTypeConfigFile:
			configFile, err := setting.AsConfigFile()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if target := settingmountservice.ConfigFileTarget(configFile); target != "" {
				claimed[target] = "config file " + setting.Name
			}
		case base.SettingTypeAppSettingMount:
			mount, err := setting.AsAppSettingMount()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			for _, f := range mount.Files {
				if f != nil {
					claimed[f.Path] = "setting mount " + setting.Name
				}
			}
		}
	}
	return claimed, nil
}

// loadClaimantsFromRepo is the app's own settings that can give it files.
func (s *service) loadClaimantsFromRepo(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.type IN (?)", bunex.List([]base.SettingType{
			base.SettingTypeSecret, base.SettingTypeConfigFile, base.SettingTypeAppSettingMount})),
	)
	return settings, hperrors.Wrap(err)
}
```

- [ ] **Step 9: `MayRevealSecrets`.** In `permission/manager.go`, beside `AuthorizeSecretReveal`:

```go
	// MayRevealSecrets answers whether the caller may see a stored secret in the
	// clear, and records nothing: for checks that reveal nothing yet, such as an
	// import's validate.
	MayRevealSecrets(ctx context.Context, db database.IDB, auth *basedto.Auth) (bool, error)
```

In `secret_reveal.go`:

```go
func (p *manager) MayRevealSecrets(ctx context.Context, db database.IDB, auth *basedto.Auth) (bool, error) {
	allowed, err := p.canRevealSecrets(ctx, db, auth, "")
	if errors.Is(err, hperrors.ErrRevealSecretsDisabled) ||
		errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets) {
		return false, nil
	}
	return allowed, err
}
```

- [ ] **Step 10: Run** `go test ./hivepaas_app/service/settingmountservice/... ./hivepaas_app/permission/... ./hivepaas_app/pkg/translation/...`, which should PASS. Then run `go build ./...` and `golangci-lint run ./hivepaas_app/...`.

- [ ] **Step 11: Commit**

```bash
git add hivepaas_app/service/settingmountservice/ hivepaas_app/hperrors/errors_settings.go \
  hivepaas_app/pkg/translation/messages/en/errors.settings.en.toml hivepaas_app/permission/
git commit -m "feat(settingmounts): the rules an entry keeps, the paths an app claims, and why an entry is not mounted

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: What a clone and a preview copy

**Files:**
- Modify: `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_2_settings.go` (`onCloneSettingDefault`)
- Modify: `hivepaas_app/service/apppreviewservice/apppreviewserviceimpl/create_preview_settings.go`, `clone_db_apps_settings.go`
- Test: `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_2_settings_test.go` (create)

**Interfaces:** none new.

The entries are not copied. The clone runs as a task without the caller's session, so it cannot pass §7's gate. The copy resolves its own entries, which it has none of, and the references a copied service spec held are already cleared (`clone_3_swarm_service.go:49-50`). This is §6's "resolved on its own".

- [ ] **Step 1: Write the failing test.**

```go
package appcloneserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
)

// A clone copies what its settings ask for, and never an app's setting mounts:
// handing a private key to a copy takes the Reveal Secrets permission, and a
// clone runs as a task with nobody's session to ask it of.
func TestACloneLeavesSettingMountsBehind(t *testing.T) {
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{CloneSettings: &entity.AppCloneSettings{
		CloneEnvVars: true, CloneSecrets: true, CloneConfigFiles: true,
	}}}
	entry := &entity.Setting{ID: "m1", Type: base.SettingTypeAppSettingMount, Name: "cert"}

	got, err := (&service{}).onCloneSettingDefault(entry, data)

	assert.NoError(t, err)
	assert.Nil(t, got)
}
```

- [ ] **Step 2: Run it.** `go test ./hivepaas_app/service/appcloneservice/... -run TestACloneLeavesSettingMountsBehind` should PASS already: the switch's default drops it. That makes this a regression pin. Record it in the ledger as `Ruling: test passes before the change - it pins today's default so a later "copy every app setting" change cannot slip entries through`.

- [ ] **Step 3: Make it explicit.** Add the case above `default:` in all three switches:

```go
	case base.SettingTypeAppSettingMount:
		// Not copied: handing a private key to the copy takes the Reveal Secrets
		// permission, and nobody's session is here to ask it of. The copy
		// resolves its own entries, of which it has none.
		return nil, nil
```

- [ ] **Step 4: Run** `go test ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/service/apppreviewservice/...`. It should PASS.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/appcloneservice/appcloneserviceimpl/ hivepaas_app/service/apppreviewservice/apppreviewserviceimpl/
git commit -m "feat(settingmounts): clones and previews leave setting mounts behind

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The `settingmountuc` use case

**Files:**
- Create: `hivepaas_app/usecase/settings/settingmountuc/`:
  - `uc.go`, `create.go`, `update.go`, `update_status.go`, `delete.go`, `get.go`, `list.go`, `sources.go`;
  - `gate.go`, with test `gate_test.go`.
- Create: `hivepaas_app/usecase/settings/settingmountuc/settingmountdto/`: `create.go`, `update.go`, `update_status.go`, `delete.go`, `get.go`, `list.go`, `sources.go`.
- Modify: `hivepaas_app/registry/provides.go` (`settingmountuc.New`, beside `configfileuc.New`).

**Interfaces:**
- Consumes: `CheckEntry`, `Grants`, `Widens`, `CheckPathsFree`, `ClaimedPaths`, `EntryStates`, `PartsOf` and `SourceTypes` (Task 1); `permission.Manager.AuthorizeSecretReveal` and `MayRevealSecrets`.
- Produces:
  - the methods `(*UC) CreateSettingMount`, `UpdateSettingMount`, `UpdateSettingMountStatus`, `DeleteSettingMount`, `GetSettingMount`, `ListSettingMount`, `ListSettingMountSources`;
  - the DTO constructors `NewCreateSettingMountReq()`, `NewUpdateSettingMountReq()`, `NewUpdateSettingMountStatusReq()`, `NewDeleteSettingMountReq()`, `NewGetSettingMountReq()`, `NewListSettingMountReq()`, `NewListSettingMountSourcesReq()`.

- [ ] **Step 1: Write the failing test** `gate_test.go`. Each use-case flow runs through `settings.BaseUC` and a real transaction, as every settings use case here does. Only the gate's decisions are unit-tested:

```go
package settingmountuc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type fakePermissions struct {
	permission.Manager
	err      error
	subjects []*permission.RevealSubject
}

func (f *fakePermissions) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	f.subjects = append(f.subjects, subject)
	return f.err
}

func entrySetting(t *testing.T, status base.SettingStatus, source string, parts ...string) *entity.Setting {
	t.Helper()
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/" + part})
	}
	setting := &entity.Setting{ID: "m1", Name: "cert", Type: base.SettingTypeAppSettingMount, Status: status}
	assert.NoError(t, setting.SetData(mount))
	return setting
}

// A disabled entry hands nothing out, so enabling one with a private key is
// growth, and asks.
func TestStatusGrantsCountOnlyWhenActive(t *testing.T) {
	disabled := entrySetting(t, base.SettingStatusDisabled, "cert_1", "certificate", "privateKey")
	enabled := entrySetting(t, base.SettingStatusActive, "cert_1", "certificate", "privateKey")

	assert.Empty(t, activeGrants(disabled))
	assert.Len(t, activeGrants(enabled), 1)
}

func TestTheGateIsAskedOnlyForWhatIsAdded(t *testing.T) {
	perms := &fakePermissions{}
	uc := &UC{BaseUC: &settings.BaseUC{PermissionManager: perms}}
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}
	before := entrySetting(t, base.SettingStatusActive, "cert_1", "privateKey")

	assert.NoError(t, uc.authorizeGrants(context.Background(), nil, scope, before, before,
		base.AuditLogSourceAPIUpdate))
	assert.Empty(t, perms.subjects, "unchanged grants ask nothing")

	after := entrySetting(t, base.SettingStatusActive, "cert_2", "privateKey")
	assert.NoError(t, uc.authorizeGrants(context.Background(), nil, scope, before, after,
		base.AuditLogSourceAPIUpdate))
	if assert.Len(t, perms.subjects, 1) {
		subject := perms.subjects[0]
		assert.Equal(t, "app_1", subject.ObjectID)
		assert.Equal(t, base.ResourceTypeSettingMount, subject.ResType)
		assert.Equal(t, "cert", subject.ResName)
		var detail map[string]any
		assert.NoError(t, json.Unmarshal([]byte(subject.Detail), &detail))
		assert.Equal(t, []any{map[string]any{"source": "cert_2", "part": "privateKey"}}, detail["grants"])
	}

	perms.err = hperrors.ErrUserNotHavePermissionOnRevealSecrets
	err := uc.authorizeGrants(context.Background(), nil, scope, nil, after, base.AuditLogSourceAPICreate)
	assert.ErrorIs(t, err, hperrors.ErrUserNotHavePermissionOnRevealSecrets)
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/usecase/settings/settingmountuc/`. It should FAIL to compile.

- [ ] **Step 3: `uc.go` and `gate.go`.**

```go
package settingmountuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeAppSettingMount
	currentSettingVersion = entity.CurrentAppSettingMountVersion
)

type UC struct {
	*settings.BaseUC
	settingMountService settingmountservice.Service
}

func New(baseUC *settings.BaseUC, settingMountService settingmountservice.Service) *UC {
	return &UC{BaseUC: baseUC, settingMountService: settingMountService}
}
```

`gate.go`:

```go
package settingmountuc

import (
	"context"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// activeGrants are what an entry hands its app now: nothing while disabled.
func activeGrants(setting *entity.Setting) []settingmountservice.Grant {
	if setting == nil || setting.Status != base.SettingStatusActive {
		return nil
	}
	mount, err := setting.AsAppSettingMount()
	if err != nil {
		return nil
	}
	return settingmountservice.Grants(mount)
}

// authorizeGrants passes §7's gate for what after hands out that before did not.
// Mounting a private key or a password is revealing it: whoever controls the
// container reads the file.
//
// It is asked on the database, not on the save's transaction: a denial rolls
// the save back, and the record of the attempt has to outlive it.
func (uc *UC) authorizeGrants(
	ctx context.Context, auth *basedto.Auth, scope *entity.ObjectScope,
	before, after *entity.Setting, source base.AuditLogSource,
) error {
	added := settingmountservice.Widens(activeGrants(before), activeGrants(after))
	if len(added) == 0 {
		return nil
	}
	detail, err := json.Marshal(map[string]any{"grants": added})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.PermissionManager.AuthorizeSecretReveal(ctx, uc.DB, auth, &permission.RevealSubject{
		Scope:    base.ObjectScopeApp,
		ObjectID: scope.AppID,
		Source:   source,
		ResType:  base.ResourceTypeSettingMount,
		ResID:    after.ID,
		ResName:  after.Name,
		Detail:   string(detail),
	}))
}
```

The parameter order is `(ctx, auth, scope, before, after, source)`; the test passes `nil` for `auth`. In `base/permission.go`, add `ResourceTypeSettingMount ResourceType = "setting-mount"` in alphabetical place, and add it wherever `ResourceTypeConfigFile` is listed in a list of all resource types (`grep -n "ResourceTypeConfigFile" hivepaas_app/base/*.go`).

- [ ] **Step 4: The DTOs.** `settingmountdto/create.go`:

```go
package settingmountdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	entryKeyMaxLen = 20
	pathMaxLen     = 512
	idMaxLen       = 64
	maxFiles       = 20
)

type SettingMountBaseReq struct {
	// Name is the entry's key.
	Name   string                    `json:"name"`
	Source basedto.ObjectIDReq       `json:"source"`
	Files  []*SettingMountFileReq    `json:"files"`
}

type SettingMountFileReq struct {
	Part string            `json:"part"`
	Path string            `json:"path"`
	UID  string            `json:"uid"`
	GID  string            `json:"gid"`
	Mode fileutil.FileMode `json:"mode"`
}

func (req *SettingMountBaseReq) ToEntity() *entity.AppSettingMount {
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: req.Source.ID}}
	for _, f := range req.Files {
		if f == nil {
			continue
		}
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{
			Part: f.Part, Path: f.Path, UID: f.UID, GID: f.GID, Mode: f.Mode,
		})
	}
	return mount
}

// validate checks the shape; what the source's type allows is the use case's.
func (req *SettingMountBaseReq) validate() (res []vld.Validator) {
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, entryKeyMaxLen, "name")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Source, true, "source")...)
	res = append(res, vld.SliceLen(req.Files, 1, maxFiles).OnError(
		vld.SetField("files", nil), vld.SetCustomKey("ERR_VLD_SLICE_LEN")))
	for _, f := range req.Files {
		if f == nil {
			continue
		}
		res = append(res, basedto.ValidateStr(&f.Path, true, 1, pathMaxLen, "files.path")...)
	}
	return res
}

type CreateSettingMountReq struct {
	settings.CreateSettingReq
	*SettingMountBaseReq
}

func NewCreateSettingMountReq() *CreateSettingMountReq {
	return &CreateSettingMountReq{}
}

func (req *CreateSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	if req.SettingMountBaseReq == nil {
		req.SettingMountBaseReq = &SettingMountBaseReq{}
	}
	validators = append(validators, req.validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSettingMountResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
```

Check the slice-length validator's API in `vendor/github.com/tiendc/go-validator`, and how other DTOs validate a slice's length (`grep -rn "SliceLen" hivepaas_app/usecase | head -3`). Copy that exact form. If no DTO does, drop the `vld.SliceLen` line: `CheckEntry` refuses an entry with no files anyway (`ERR_SETTING_MOUNT_NO_FILES`). `idMaxLen` is unused if `ValidateObjectIDReq` covers the id; leave it out if lint says so.

`update.go`:

```go
type UpdateSettingMountReq struct {
	settings.UpdateSettingReq
	*SettingMountBaseReq
}

func NewUpdateSettingMountReq() *UpdateSettingMountReq {
	return &UpdateSettingMountReq{}
}

func (req *UpdateSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	if req.SettingMountBaseReq == nil {
		req.SettingMountBaseReq = &SettingMountBaseReq{}
	}
	validators = append(validators, req.validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSettingMountResp struct {
	Meta *basedto.Meta `json:"meta"`
}
```

`update_status.go`, `delete.go` and `list.go` are `configfiledto`'s with `ConfigFile` renamed `SettingMount`. Their bodies are in `usecase/settings/configfileuc/configfiledto/update_status.go`, `delete.go` and `list.go`. Copy each file and rename: `UpdateSettingMountStatusReq`/`Resp`, `DeleteSettingMountReq`/`Resp`, `ListSettingMountReq`/`Resp` with `Data []*SettingMountResp`, and the list sorted by `name`.

`get.go`:

```go
type GetSettingMountReq struct {
	settings.GetSettingReq
}

func NewGetSettingMountReq() *GetSettingMountReq {
	return &GetSettingMountReq{}
}

func (req *GetSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSettingMountResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *SettingMountResp `json:"data"`
}

type SettingMountResp struct {
	*settings.BaseSettingResp
	Source *settings.BaseSettingResp       `json:"source"`
	Files  []*SettingMountFileResp         `json:"files"`
	State  *settingmountservice.EntryState `json:"state"`
}

type SettingMountFileResp struct {
	Part      string            `json:"part"`
	Path      string            `json:"path"`
	UID       string            `json:"uid"`
	GID       string            `json:"gid"`
	Mode      fileutil.FileMode `json:"mode"`
	Sensitive bool              `json:"sensitive"`
}

// TransformSettingMount renders an entry with its source and its state. state
// is nil when the app's states could not be read, which the screen shows as
// unknown.
func TransformSettingMount(
	setting *entity.Setting, refObjects *entity.RefObjects, state *settingmountservice.EntryState,
) (*SettingMountResp, error) {
	mount, err := setting.AsAppSettingMount()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &SettingMountResp{State: state}
	if resp.BaseSettingResp, err = settings.TransformSettingBase(setting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if refObjects != nil && refObjects.RefSettings[mount.Source.ID] != nil {
		resp.Source, _ = settings.TransformSettingBase(refObjects.RefSettings[mount.Source.ID])
	}
	if resp.Source == nil {
		resp.Source = settings.NewMissingSetting(mount.Source.ID, "")
	}
	for _, f := range mount.Files {
		resp.Files = append(resp.Files, &SettingMountFileResp{Part: f.Part, Path: f.Path, UID: f.UID,
			GID: f.GID, Mode: f.Mode, Sensitive: settingmountservice.SensitivePart(f.Part)})
	}
	return resp, nil
}
```

Check what `settings.NewMissingSetting` takes (`grep -n "func NewMissingSetting" -A5 hivepaas_app/usecase/settings/*.go`), and pass the type it wants.

`list.go` also holds:

```go
func TransformSettingMounts(
	settings []*entity.Setting, refObjects *entity.RefObjects, states map[string]*settingmountservice.EntryState,
) ([]*SettingMountResp, error) {
	out := make([]*SettingMountResp, 0, len(settings))
	for _, setting := range settings {
		resp, err := TransformSettingMount(setting, refObjects, states[setting.ID])
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out = append(out, resp)
	}
	return out, nil
}
```

The parameter name shadows the `settings` package. Name it `entries`.

`sources.go`:

```go
type ListSettingMountSourcesReq struct {
	Scope *entity.ObjectScope `json:"-"`
}

func NewListSettingMountSourcesReq() *ListSettingMountSourcesReq {
	return &ListSettingMountSourcesReq{}
}

func (req *ListSettingMountSourcesReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ListSettingMountSourcesResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data []*SettingMountSource  `json:"data"`
	// MayMountSensitive says whether the caller may mount a private key or a
	// password: the screen locks those parts, with the reason, when not.
	MayMountSensitive bool `json:"mayMountSensitive"`
}

type SettingMountSource struct {
	Type  base.SettingType     `json:"type"`
	Parts []*SettingMountPart  `json:"parts"`
}

type SettingMountPart struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	Sensitive bool   `json:"sensitive"`
}
```

- [ ] **Step 5: The use case methods.** `create.go`:

```go
func (uc *UC) CreateSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.CreateSettingMountReq,
) (*settingmountdto.CreateSettingMountResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	mount := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: mount.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context, db database.Tx,
			_ *settings.CreateSettingData, pData *settings.PersistingSettingCreationData,
		) error {
			if err := uc.checkEntry(ctx, db, req.Scope, "", req.Name, mount); err != nil {
				return err
			}
			if err := pData.Setting.SetData(mount); err != nil {
				return hperrors.Wrap(err)
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, nil, pData.Setting, base.AuditLogSourceAPICreate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.CreateSettingMountResp{Data: resp.Data}, nil
}

// checkEntry refuses an entry wrong for its source, or with a path another file
// of the app has. The source's existence and visibility from the app were
// checked with the references.
func (uc *UC) checkEntry(
	ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	exceptSettingID, key string, mount *entity.AppSettingMount,
) error {
	if scope == nil || !scope.IsAppScope() {
		return hperrors.NewUnsupported("Setting mount outside an app")
	}
	source, err := uc.SettingRepo.GetByID(ctx, db, scope, "", mount.Source.ID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = settingmountservice.CheckEntry(key, mount, source.Type); err != nil {
		return err
	}
	claimed, err := uc.settingMountService.ClaimedPaths(ctx, db, scope.AppID, exceptSettingID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	paths := make([]string, 0, len(mount.Files))
	for _, f := range mount.Files {
		paths = append(paths, f.Path)
	}
	return settingmountservice.CheckPathsFree(claimed, paths...)
}
```

`update.go`:

```go
func (uc *UC) UpdateSettingMount(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.UpdateSettingMountReq,
) (*settingmountdto.UpdateSettingMountResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	mount := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: mount.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context, db database.Tx,
			data *settings.UpdateSettingData, pData *settings.PersistingSettingData,
		) error {
			if err := uc.checkEntry(ctx, db, req.Scope, data.Setting.ID, req.Name, mount); err != nil {
				return err
			}
			if err := pData.Setting.SetData(mount); err != nil {
				return hperrors.Wrap(err)
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, data.Setting, pData.Setting, base.AuditLogSourceAPIUpdate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.UpdateSettingMountResp{}, nil
}
```

`update_status.go`: enabling rechecks paths, since a secret may have taken one meanwhile, and passes the gate:

```go
func (uc *UC) UpdateSettingMountStatus(
	ctx context.Context, auth *basedto.Auth, req *settingmountdto.UpdateSettingMountStatusReq,
) (*settingmountdto.UpdateSettingMountStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{
		BeforePersisting: func(
			ctx context.Context, db database.Tx,
			data *settings.UpdateSettingStatusData, pData *settings.PersistingSettingStatusData,
		) error {
			if !pData.Setting.IsActive() {
				return nil
			}
			mount, err := pData.Setting.AsAppSettingMount()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if err = uc.checkEntry(ctx, db, req.Scope, pData.Setting.ID, pData.Setting.Name, mount); err != nil {
				return err
			}
			return uc.authorizeGrants(ctx, auth, req.Scope, data.Setting, pData.Setting, base.AuditLogSourceAPIUpdate)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &settingmountdto.UpdateSettingMountStatusResp{}, nil
}
```

`delete.go` calls `uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})`. The event records the refresh (plan 1), and no gate applies.

`get.go` and `list.go` call `GetSetting`/`ListSetting` as `configfileuc` does. They then read the states once, through `uc.settingMountService.EntryStates(ctx, uc.DB, req.Scope.App)`. If `req.Scope.App` is nil there, load it with `uc.AppService` or the app repository the base has, whichever `grep -n "App" hivepaas_app/usecase/settings/base_uc.go` shows. A failure to read states is not a failure of the read: log nothing, pass `nil` states, and the response has `state: null`.

`sources.go`:

```go
func (uc *UC) ListSettingMountSources(
	ctx context.Context, auth *basedto.Auth, _ *settingmountdto.ListSettingMountSourcesReq,
) (*settingmountdto.ListSettingMountSourcesResp, error) {
	mayMount, err := uc.PermissionManager.MayRevealSecrets(ctx, uc.DB, auth)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &settingmountdto.ListSettingMountSourcesResp{MayMountSensitive: mayMount}
	for _, typ := range settingmountservice.SourceTypes() {
		source := &settingmountdto.SettingMountSource{Type: typ}
		for _, part := range settingmountservice.PartsOf(typ) {
			source.Parts = append(source.Parts, &settingmountdto.SettingMountPart{
				Name: part.Name, Required: part.Required, Sensitive: part.Sensitive})
		}
		resp.Data = append(resp.Data, source)
	}
	return resp, nil
}
```

Register `settingmountuc.New` in `registry/provides.go`, beside `configfileuc.New`.

- [ ] **Step 6: Run** `go build ./... && go test ./hivepaas_app/usecase/settings/settingmountuc/... ./hivepaas_app/cmd/...`. It should PASS, and the fx wiring should resolve. Run `golangci-lint run ./hivepaas_app/usecase/settings/...`.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/usecase/settings/settingmountuc/ hivepaas_app/base/permission.go hivepaas_app/registry/provides.go
git commit -m "feat(settingmounts): the use case for entries, behind the Reveal Secrets gate when they hand out more

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: The endpoints, and secrets and config files keeping their paths

**Files:**
- Create: `hivepaas_app/interface/api/handler/appsettingshandler/setting_mount.go`
- Modify: `hivepaas_app/interface/api/handler/basesettinghandler/`:
  - `handler.go`: the `SettingMountUC *settingmountuc.UC` field and `New` parameter;
  - a `case base.ResourceTypeSettingMount:` in `create.go`, `update.go`, `list.go`, `get.go`, `delete.go` and `update_status.go`.
- Modify: `hivepaas_app/interface/api/server/router_apps.go`
- Modify: `hivepaas_app/usecase/settings/base_uc.go` (`SettingMountService` field and `New` parameter), create `usecase/settings/mount_paths.go`, test `mount_paths_test.go`
- Modify: `usecase/settings/secretuc/create.go`, `secretuc/update.go`, `configfileuc/create.go`, `configfileuc/update.go`
- Modify: `docs/openapi/swagger.json` (by `make gen-swag`)

**Interfaces:**
- Consumes: Task 3's use case and DTO constructors; `ClaimedPaths` and `CheckPathsFree`.
- Produces: `(*settings.BaseUC) CheckMountPaths(ctx, db, scope *entity.ObjectScope, exceptSettingID string, paths ...string) error`.

- [ ] **Step 1: Write the failing test** `usecase/settings/mount_paths_test.go`:

```go
package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

type fakeClaims struct {
	settingmountservice.Service
	claimed map[string]string
	except  string
}

func (f *fakeClaims) ClaimedPaths(_ context.Context, _ database.IDB, _, except string) (map[string]string, error) {
	f.except = except
	return f.claimed, nil
}

// db_password and /run/secrets/db_password are one file.
func TestASecretNamedRelativelyCollidesWithAnEntry(t *testing.T) {
	claims := &fakeClaims{claimed: map[string]string{"/run/secrets/db_password": "setting mount cert"}}
	uc := &BaseUC{SettingMountService: claims}
	app := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}

	err := uc.CheckMountPaths(context.Background(), nil, app, "s1",
		settingmountservice.SecretTarget("db_password"))

	assert.True(t, errors.Is(err, hperrors.ErrSettingMountPathTaken), "%v", err)
	assert.Equal(t, "s1", claims.except, "the secret being saved does not collide with itself")
	assert.NoError(t, uc.CheckMountPaths(context.Background(), nil,
		&entity.ObjectScope{ScopeType: base.ObjectScopeProject, ProjectID: "p1"}, "", "/run/secrets/db_password"),
		"a project's secret is no app's file")
	assert.NoError(t, uc.CheckMountPaths(context.Background(), nil, app, "", ""), "no file, nothing to check")
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/usecase/settings/ -run TestASecretNamedRelatively`. It should FAIL to compile.

- [ ] **Step 3: `mount_paths.go`**, plus a `SettingMountService settingmountservice.Service` field and `New` parameter on `BaseUC`:

```go
package settings

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// CheckMountPaths refuses a file of an app at a path one of its other secrets,
// config files or setting mounts has. Only an app's settings are files of a
// container; an empty path is a setting with no file.
func (uc *BaseUC) CheckMountPaths(
	ctx context.Context, db database.IDB, scope *entity.ObjectScope, exceptSettingID string, paths ...string,
) error {
	paths = gofn.ToSliceSkippingZero(paths...)
	if scope == nil || !scope.IsAppScope() || len(paths) == 0 {
		return nil
	}
	claimed, err := uc.SettingMountService.ClaimedPaths(ctx, db, scope.AppID, exceptSettingID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return settingmountservice.CheckPathsFree(claimed, paths...)
}
```

- [ ] **Step 4: Secrets and config files check.** In `secretuc/create.go`, as the first statement of `PrepareCreation`:

```go
			if err := uc.CheckMountPaths(ctx, db, req.Scope, "",
				settingmountservice.SecretFileTarget(secret)); err != nil {
				return err
			}
```

In `secretuc/update.go`, after `updatedSecret` has its final value in `PrepareUpdate`, before `SetData`, check with `data.Setting.ID` as the exception:

```go
			if err = uc.CheckMountPaths(ctx, db, req.Scope, data.Setting.ID,
				settingmountservice.SecretFileTarget(updatedSecret)); err != nil {
				return err
			}
```

Do the same in `configfileuc/create.go` with `ConfigFileTarget(configFile)`, and in `configfileuc/update.go` with `ConfigFileTarget(updatedConfigFile)` and `data.Setting.ID`.

- [ ] **Step 5: Handlers and routes.** Add to each of the six basesettinghandler files, beside the `ResourceTypeConfigFile` case:

```go
	case base.ResourceTypeSettingMount: // create.go
		r := settingmountdto.NewCreateSettingMountReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.CreateSettingMount(reqCtx, auth, r) }
```

```go
	case base.ResourceTypeSettingMount: // update.go
		r := settingmountdto.NewUpdateSettingMountReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.UpdateSettingMount(reqCtx, auth, r) }
```

```go
	case base.ResourceTypeSettingMount: // list.go
		r := settingmountdto.NewListSettingMountReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.ListSettingMount(reqCtx, auth, r) }
```

```go
	case base.ResourceTypeSettingMount: // get.go
		r := settingmountdto.NewGetSettingMountReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.GetSettingMount(reqCtx, auth, r) }
```

```go
	case base.ResourceTypeSettingMount: // delete.go
		r := settingmountdto.NewDeleteSettingMountReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.DeleteSettingMount(reqCtx, auth, r) }
```

```go
	case base.ResourceTypeSettingMount: // update_status.go
		r := settingmountdto.NewUpdateSettingMountStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SettingMountUC.UpdateSettingMountStatus(reqCtx, auth, r) }
```

`appsettingshandler/setting_mount.go`:
- The six generic handlers, `ListSettingMount`, `GetSettingMount`, `CreateSettingMount`, `UpdateSettingMount`, `UpdateSettingMountStatus` and `DeleteSettingMount`, each calling `h.<Verb>Setting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)`.
- Each gets swagger annotations as `config_file.go`'s do, with `@Id` `listAppSettingMount` and so on, and `@Router /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts...`.
- Plus the sources handler:

```go
// ListSettingMountSources Lists the settings an app can mount, and their parts
// @Summary Lists the settings an app can mount, and their parts
// @Description Lists the settings an app can mount, and their parts
// @Tags    app_settings
// @Produce json
// @Id      listAppSettingMountSources
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} settingmountdto.ListSettingMountSourcesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/sources [get]
func (h *Handler) ListSettingMountSources(ctx *gin.Context) {
	auth, _, _, _, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.SettingMountUC.ListSettingMountSources(h.RequestCtx(ctx), auth,
		settingmountdto.NewListSettingMountSourcesReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
```

In `router_apps.go`, after the config-files group:

```go
	{ // Setting mounts
		settingMountGroup := appGroup.Group("/:appID/setting-mounts")
		settingMountGroup.GET("", appSettingsHandler.ListSettingMount)
		settingMountGroup.GET("/sources", appSettingsHandler.ListSettingMountSources)
		settingMountGroup.GET("/:itemID", appSettingsHandler.GetSettingMount)
		settingMountGroup.POST("", appSettingsHandler.CreateSettingMount)
		settingMountGroup.PUT("/:itemID", appSettingsHandler.UpdateSettingMount)
		settingMountGroup.PUT("/:itemID/status", appSettingsHandler.UpdateSettingMountStatus)
		settingMountGroup.DELETE("/:itemID", appSettingsHandler.DeleteSettingMount)
	}
```

- [ ] **Step 6: Run** `go test ./hivepaas_app/usecase/settings/... ./hivepaas_app/cmd/...`, then `go build ./... && golangci-lint run ./... && make gen-swag`. All should PASS or be clean. Check that `docs/openapi/swagger.json` has the seven operations: `grep -c "setting-mounts" docs/openapi/swagger.json` should be 4 or more.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/interface/api/ hivepaas_app/usecase/settings/base_uc.go hivepaas_app/usecase/settings/mount_paths*.go \
  hivepaas_app/usecase/settings/secretuc/create.go hivepaas_app/usecase/settings/secretuc/update.go \
  hivepaas_app/usecase/settings/configfileuc/create.go hivepaas_app/usecase/settings/configfileuc/update.go \
  docs/openapi/swagger.json
git commit -m "feat(settingmounts): the setting-mounts endpoints, and one path for one file across secrets, config files and mounts

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: An ordinary file takes its path back in Docker

**Files:**
- Create: `hivepaas_app/service/clustersecretservice/clustersecretserviceimpl/mount_targets.go`, test `mount_targets_test.go`
- Modify: `secret_create.go` (`addSwarmSecretsToService`), `config_create.go` (`addSwarmConfigsToService`)

**Interfaces:**
- Produces:
  - `makeRoomForSecret(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool)`;
  - `makeRoomForConfig(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool)`;
  - `(s *service) mountedIDs(ctx) (map[string]bool, error)`.

Save-time checks (Task 4) keep new clashes out. What gets past them is:
- a secret enabled again;
- an import;
- data from before this work.

Docker then decides, and the rule is the engine's: an ordinary file wins.

- [ ] **Step 1: Write the failing test.**

```go
package clustersecretserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

// A mount at the path goes; an ordinary file there keeps it; the comparison is
// of where the files land, not of how they are named.
func TestMakeRoomComparesResolvedTargets(t *testing.T) {
	spec := &swarm.ContainerSpec{Secrets: []*swarm.SecretReference{
		{SecretID: "mount_key", File: &swarm.SecretReferenceFileTarget{Name: "/run/secrets/app_key"}},
		{SecretID: "db", File: &swarm.SecretReferenceFileTarget{Name: "db_password"}},
	}}
	mounted := map[string]bool{"mount_key": true}

	assert.False(t, makeRoomForSecret(spec, "/run/secrets/app_key", mounted), "the mount steps aside")
	assert.Len(t, spec.Secrets, 1)
	assert.True(t, makeRoomForSecret(spec, "/run/secrets/db_password", mounted),
		"an ordinary secret already there, named relatively")
	assert.False(t, makeRoomForSecret(spec, "/run/secrets/other", mounted))

	cfg := &swarm.ContainerSpec{Configs: []*swarm.ConfigReference{
		{ConfigID: "mount_cert", File: &swarm.ConfigReferenceFileTarget{Name: "/app.pem"}},
	}}
	assert.False(t, makeRoomForConfig(cfg, "/app.pem", map[string]bool{"mount_cert": true}))
	assert.Empty(t, cfg.Configs)
}
```

- [ ] **Step 2: Run** it. It should FAIL to compile.

- [ ] **Step 3: `mount_targets.go`:**

```go
package clustersecretserviceimpl

import (
	"context"
	"slices"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// mountedIDs are the secrets and configs setting mounts hold, whichever app's.
func (s *service) mountedIDs(ctx context.Context) (map[string]bool, error) {
	ids := map[string]bool{}
	secrets, err := s.dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		docker.FilterAdd(&opts.Filters, "label", settingmountservice.LabelEntry)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, secret := range secrets.Items {
		ids[secret.ID] = true
	}
	configs, err := s.dockerManager.ConfigList(ctx, func(opts *client.ConfigListOptions) {
		docker.FilterAdd(&opts.Filters, "label", settingmountservice.LabelEntry)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, config := range configs.Items {
		ids[config.ID] = true
	}
	return ids, nil
}

// makeRoomForSecret frees target for an ordinary secret: a setting mount there
// steps aside, as it does when the engine applies (an ordinary file wins). It
// reports whether an ordinary secret already has the path.
func makeRoomForSecret(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool) {
	spec.Secrets = slices.DeleteFunc(spec.Secrets, func(ref *swarm.SecretReference) bool {
		if ref.File == nil || settingmountservice.SecretTarget(ref.File.Name) != target {
			return false
		}
		if mounted[ref.SecretID] {
			return true
		}
		taken = true
		return false
	})
	return taken
}

// makeRoomForConfig is makeRoomForSecret for a config.
func makeRoomForConfig(spec *swarm.ContainerSpec, target string, mounted map[string]bool) (taken bool) {
	spec.Configs = slices.DeleteFunc(spec.Configs, func(ref *swarm.ConfigReference) bool {
		if ref.File == nil || settingmountservice.ConfigTarget(ref.File.Name) != target {
			return false
		}
		if mounted[ref.ConfigID] {
			return true
		}
		taken = true
		return false
	})
	return taken
}
```

- [ ] **Step 4: Use them.** In `addSwarmSecretsToService`:
  - Before `ServiceUpdateFunc`, read `mounted, err := s.mountedIDs(ctx)` and return the error wrapped.
  - Inside the loop, replace the `gofn.Find(... sec.File.Name == swarmRef.File.Name)` block with:

```go
				if makeRoomForSecret(containerSpec, settingmountservice.SecretTarget(swarmRef.File.Name), mounted) {
					continue
				}
```

Make the same change in `addSwarmConfigsToService` with `makeRoomForConfig` and `ConfigTarget`. If `gofn` is no longer used in the file, drop its import.

- [ ] **Step 5: Run** `go test ./hivepaas_app/service/clustersecretservice/... ./hivepaas_app/cmd/...`. It should PASS. Then run `go build ./...` and `golangci-lint run ./hivepaas_app/service/clustersecretservice/...`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/clustersecretservice/clustersecretserviceimpl/
git commit -m "fix(secrets): an ordinary secret or config file takes its path back from a setting mount

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: Export and import

**Files:**
- Modify: `hivepaas_app/entity/setting_spec.go`: drop the `skipSpecPolicy` of `SettingTypeAppSettingMount`, and add the type to `registerDefaultSpecPolicies`.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`: reword the reason to `"only an app mounts settings: a scope's entry would mount nothing"`.
- Modify: `hivepaas_app/service/specservice/specmodel/importplan.go`: `CodeSettingMountNotPermitted = "SETTING_MOUNT_NOT_PERMITTED"`.
- Modify: `hivepaas_app/service/specservice/types.go`: `MayMountSecrets` on `ValidateImportReq`.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/`:
  - `import_plan.go`: the planner fields `mountSecretsAllowed *bool`;
  - `import_checks.go`: `checkSettingMounts`, `mayMountSecrets`, `refuseAppSetting`, `entryGrants`, called from `checkPermissions`;
  - `import_write.go`: `writtenNames` honors `refused` for apps.
- Modify: `hivepaas_app/usecase/specuc/import_validate.go`, `import_apply.go`: `importReq(auth, req, record bool)`.
- Test: `export_test.go` (a fixture entry), `import_setting_mounts_test.go` (create)

**Interfaces:**
- Consumes: `settingmountservice.SensitivePart`; `permission.Manager.MayRevealSecrets` and `AuthorizeSecretReveal`.
- Produces: `ValidateImportReq.MayMountSecrets func(ctx context.Context) (bool, error)`.

- [ ] **Step 1: The fixture entry.** In `exportFixture` (`export_test.go`), after `secret`:

```go
	// The backend mounts the certificate the routing uses.
	mountEntry := &entity.Setting{
		ID: "mount_1", Type: base.SettingTypeAppSettingMount, Scope: base.ObjectScopeApp,
		ObjectID: "app_1", Name: "cert", Status: base.SettingStatusActive, Version: entity.CurrentAppSettingMountVersion,
	}
	assert.NoError(t, mountEntry.SetData(&entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "certificate", Path: "/etc/app/tls/cert.pem"}},
	}))
```

Add `mountEntry` to `all`.

- [ ] **Step 2: Write the failing tests** `import_setting_mounts_test.go`:

```go
package specserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const mountChange = "settings.settingMounts/cert"

func backendMount(bundle *specmodel.ImportBundle) map[string]any {
	entries, _ := backendSettings(bundle)["settingMounts"].(map[string]any)
	entry, _ := entries["cert"].(map[string]any)
	return entry
}

func addPrivateKey(bundle *specmodel.ImportBundle) {
	entry := backendMount(bundle)
	files, _ := entry["files"].([]any)
	entry["files"] = append(files, map[string]any{"part": "privateKey", "path": "/etc/app/tls/key.pem"})
}

func TestExportWritesAnEntryWithItsSourceAndFiles(t *testing.T) {
	_, bundle := planFixture(t)

	entry := backendMount(bundle)

	if assert.NotNil(t, entry) {
		assert.NotNil(t, entry["source"])
		assert.Len(t, entry["files"], 1)
	}
}

func planAskingMounts(
	t *testing.T, svc *service, bundle *specmodel.ImportBundle, allowed bool,
) (*specmodel.ImportPlan, *int) {
	t.Helper()
	asked := 0
	out := planWith(t, svc, bundle, &specservice.ValidateImportReq{
		MayMountSecrets: func(context.Context) (bool, error) { asked++; return allowed, nil },
	})
	return out, &asked
}

// The rest of the app is imported; the entry that hands out a private key is
// not, for a caller who may not reveal secrets.
func TestPlanSkipsAMountOfAPrivateKeyForAnOperatorWhoMayNotRevealSecrets(t *testing.T) {
	for allowed, wantChange := range map[bool]bool{false: false, true: true} {
		svc, bundle := planFixture(t)
		addPrivateKey(bundle)

		out, asked := planAskingMounts(t, svc, bundle, allowed)

		backend := node(t, out, backendPath)
		assert.Equal(t, 1, *asked)
		assert.Equal(t, wantChange, contains(backend.Changes, mountChange), "allowed=%v", allowed)
		if !wantChange {
			issues := issuesOf(backend, specmodel.CodeSettingMountNotPermitted)
			if assert.Len(t, issues, 1) {
				assert.Equal(t, specmodel.SeveritySkipped, issues[0].Severity)
			}
		}
	}
}

// The installed entry and the bundle's agree: nothing new is handed out.
func TestPlanDoesNotAskForAnEntryThatGrantsNothingNew(t *testing.T) {
	svc, bundle := planFixture(t)
	backendMount(bundle)["files"] = []any{map[string]any{"part": "certificate", "path": "/srv/cert.pem"}}

	out, asked := planAskingMounts(t, svc, bundle, false)

	assert.Zero(t, *asked, "a path moved; no pair was added")
	assert.Contains(t, node(t, out, backendPath).Changes, mountChange)
}

func TestApplyLeavesARefusedEntryOut(t *testing.T) {
	svc, bundle := planFixture(t)
	addPrivateKey(bundle)
	req := applyReq(t, svc, bundle)
	req.MayMountSecrets = func(context.Context) (bool, error) { return false, nil }
	req.AcceptIssues = true
	req.PlanHash = planWith(t, svc, bundle, &req.ValidateImportReq).PlanHash

	apply(t, svc, bundle, req)

	for _, setting := range persisted(svc).UpsertingSettings {
		assert.NotEqual(t, base.SettingTypeAppSettingMount, setting.Type, "the refused entry is not written")
	}
}

// An app created by the import writes every setting it holds; a refused entry
// is still left out.
func TestApplyLeavesARefusedEntryOutOfANewApp(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, map[string]any{"settingMounts": map[string]any{"key": map[string]any{
		"source": backendMount(bundle)["source"],
		"files":  []any{map[string]any{"part": "privateKey", "path": "/etc/key.pem"}},
	}}})
	req := applyReq(t, svc, bundle)
	req.MayMountSecrets = func(context.Context) (bool, error) { return false, nil }
	req.AcceptIssues = true
	req.PlanHash = planWith(t, svc, bundle, &req.ValidateImportReq).PlanHash

	apply(t, svc, bundle, req)

	provision := svc.appProvisionService.(*fakeProvisionService)
	for _, settings := range provision.settings {
		for _, setting := range settings {
			assert.NotEqual(t, base.SettingTypeAppSettingMount, setting.Type)
		}
	}
	_ = entity.CurrentAppSettingMountVersion
}
```

Check the helpers this relies on:
- `contains`, `issuesOf`, `planWith`, `applyReq`, `persisted`, `addWorker`, `backendPath` and the fake `provision.settings` should all exist in the package's tests (`grep -n "^func contains\|^func issuesOf\|^func planWith\|^func applyReq\|backendPath =" hivepaas_app/service/specservice/specserviceimpl/*_test.go`). Use them as they are.
- If `persisted(svc)` does not show app settings of an existing app, find where the fixture's apply records them (`grep -n "UpsertingSettings\|updatedApps" fakes_apply_test.go`), and assert there.
- Remove the `_ = entity...` line if `entity` is used elsewhere in the file.

- [ ] **Step 3: Run** `go test ./hivepaas_app/service/specservice/...`. The new tests should FAIL, and `TestExportWritesAnEntryWithItsSourceAndFiles` should fail because export skips the type. Some existing tests may now see the fixture's entry. For each that fails, decide whether it enumerates the backend's settings or references; update its expectation to include the entry, and record a `Ruling:` line per test.

- [ ] **Step 4: Implement.**
  - **Export and the reason:** export's default policy, and the reworded import reason for scopes.
  - **Issue code:** `specmodel.CodeSettingMountNotPermitted`.
  - **`types.go`:**

```go
	// MayMountSecrets answers whether this caller may have an app mount a
	// private key or a password - revealing it to whoever controls the app. It is
	// asked at most once, and only when an entry the import writes hands out a
	// (source, sensitive part) pair the installed entry did not. Nil allows it.
	MayMountSecrets func(ctx context.Context) (bool, error)
```

  - **`import_plan.go`:** in `planner`, `mountSecretsAllowed *bool // the answer of MayMountSecrets, once asked`.
  - **`import_checks.go`:**

```go
// checkSettingMounts leaves out of an app the entries that would hand it a
// private key or a password it did not have, for a caller who may not reveal
// secrets. The rest of the app is imported.
func (p *planner) checkSettingMounts(ctx context.Context, node *specmodel.PlanNode) error {
	block := specmodel.CollectionBlockName(base.SettingTypeAppSettingMount)
	entries, _ := p.apps[node.Path].doc.Settings[block].(map[string]any)
	var installed map[string]any
	if current := p.currentApp(node); current != nil {
		installed, _ = current.Settings[block].(map[string]any)
	}
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		name := block + "/" + key
		if !writesBlock(node, "settings."+name) || !entryWidens(entries[key], installed[key]) {
			continue
		}
		allowed, err := p.mayMountSecrets(ctx)
		if err != nil {
			return err
		}
		if !allowed {
			p.refuseAppSetting(node, name, specmodel.Issue{
				Severity: specmodel.SeveritySkipped, Code: specmodel.CodeSettingMountNotPermitted,
				Path: node.Path, Detail: map[string]any{refInSetting: name},
				Action: "not imported: mounting a private key or a password takes the Reveal Secrets permission",
			})
		}
	}
	return nil
}

// mayMountSecrets asks MayMountSecrets once per plan.
func (p *planner) mayMountSecrets(ctx context.Context) (bool, error) {
	if p.mountSecretsAllowed == nil {
		allowed := true
		if p.req.MayMountSecrets != nil {
			var err error
			if allowed, err = p.req.MayMountSecrets(ctx); err != nil {
				return false, hperrors.Wrap(err)
			}
		}
		p.mountSecretsAllowed = &allowed
	}
	return *p.mountSecretsAllowed, nil
}

// entryWidens reports whether a bundle's entry hands out a sensitive part the
// installed one did not: a new entry, another source, a part added, or an entry
// enabled. Sources are compared as the documents write them, since both are
// exports; a bundle from another installation names them differently, and is
// asked about.
func entryWidens(bundleBody, installedBody any) bool {
	source, parts := entryGrants(bundleBody)
	if len(parts) == 0 {
		return false
	}
	installedSource, installedParts := entryGrants(installedBody)
	if !sameBody(source, installedSource) {
		return true
	}
	for _, part := range parts {
		if !slices.Contains(installedParts, part) {
			return true
		}
	}
	return false
}

// entryGrants reads the source and the sensitive parts of an exported entry;
// none for a disabled one.
func entryGrants(body any) (source any, parts []string) {
	entry, _ := body.(map[string]any)
	if entry == nil {
		return nil, nil
	}
	if meta, _ := entry[specmodel.SettingMetaKey].(map[string]any); meta != nil {
		if status, _ := meta["status"].(string); status != "" && status != string(base.SettingStatusActive) {
			return nil, nil
		}
	}
	files, _ := entry["files"].([]any)
	for _, f := range files {
		file, _ := f.(map[string]any)
		if part, _ := file["part"].(string); settingmountservice.SensitivePart(part) {
			parts = append(parts, part)
		}
	}
	return entry["source"], parts
}

// refuseAppSetting is refuse for an app's setting, whose change is named with
// the "settings." prefix app documents use.
func (p *planner) refuseAppSetting(node *specmodel.PlanNode, name string, issue specmodel.Issue) {
	if p.refused == nil {
		p.refused = map[string][]string{}
	}
	p.refused[node.Path] = append(p.refused[node.Path], name)
	node.Changes = slices.DeleteFunc(node.Changes, func(change string) bool { return change == "settings."+name })
	node.Issues = append(node.Issues, issue)
}
```

  - **Calling it:** in `checkPermissions`, after `checkSharedMounts` for a node that is not skipped, call `if err := p.checkSettingMounts(ctx, node); err != nil { return err }`. Check that `sameBody` exists (it is used by `settingsChanges`) and takes `(any, any)`.
  - **`import_write.go` `writtenNames`:**

```go
	if place, isApp := w.p.apps[node.Path]; isApp {
		refused := w.p.refused[node.Path]
		return slices.DeleteFunc(appWrittenNames(node, place.doc), func(name string) bool {
			return slices.Contains(refused, name)
		})
	}
```

  - **Provisioning:** check `import_apply_apps.go` for where a created app's settings reach provisioning. If it takes them from `w.appSettings` (filled by `writeSettings`), the filter above covers it. If it builds them from the document again (`BuildApp` with `Import: true` and `buildImportedSettings`), delete the refused entries from `place.doc.Settings[block]` in `refuseAppSetting` as well:

```go
	if place, isApp := p.apps[node.Path]; isApp {
		block, key, _ := strings.Cut(name, "/")
		if entries, ok := place.doc.Settings[block].(map[string]any); ok {
			delete(entries, key)
		}
	}
```

    `TestApplyLeavesARefusedEntryOutOfANewApp` decides which.
  - **Use case:** in `importReq(auth, req, record bool)`, pass `false` from `ValidateImport` and `true` from `ApplyImport`:

```go
		// Validate asks without recording; apply records, allowed or denied.
		MayMountSecrets: func(ctx context.Context) (bool, error) {
			if !record {
				return uc.permissionManager.MayRevealSecrets(ctx, uc.db, auth)
			}
			err := uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
				Scope:    req.Scope.ScopeType,
				ObjectID: req.Scope.ScopeObjectID(),
				Source:   base.AuditLogSourceAPIAction,
				ResType:  base.ResourceTypeSettingMount,
				ResName:  "configuration spec import (setting mounts)",
			})
			if errors.Is(err, hperrors.ErrRevealSecretsDisabled) ||
				errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets) {
				return false, nil
			}
			if err != nil {
				return false, hperrors.Wrap(err)
			}
			return true, nil
		},
```

- [ ] **Step 5: Run** `go test ./hivepaas_app/service/specservice/... ./hivepaas_app/usecase/specuc/... ./hivepaas_app/entity/...`. It should PASS, including `TestApplyOfAnUnchangedInstallationWritesNothing` with the entry in the fixture.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/entity/setting_spec.go hivepaas_app/service/specservice/ hivepaas_app/usecase/specuc/
git commit -m "feat(settingmounts): export writes entries, and import leaves out those the caller may not reveal

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: Gates, the spec, merge

**Files:**
- Modify: `docs/superpowers/specs/2026-09-25-setting-mounts-design.md`:
  - §6: clone and preview do not copy entries, and why;
  - §1: path uniqueness is checked on both sides, and Docker lets an ordinary file win.

- [ ] **Step 1: Gates.** `go build ./...`, `golangci-lint run ./...`, `go test ./...` and `make gen-swag` should be clean, with no diff after `gen-swag`.
- [ ] **Step 2: Update the spec** as listed. Keep its style.
- [ ] **Step 3: Commit, merge, delete the branch**

```bash
git add docs/superpowers/specs/2026-09-25-setting-mounts-design.md
git commit -m "docs(spec): setting mounts, as plan 2 built them

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main && git merge --no-ff feat/setting-mounts-surfaces -m "Merge branch 'feat/setting-mounts-surfaces'"
go test ./... && git branch -d feat/setting-mounts-surfaces
```

- [ ] **Step 4: Tell the user what to check on the Linux server:**
  1. As an administrator, create an entry mounting a certificate and its key through the API. The files appear in the container, and the audit log has an allowed reveal.
  2. As a member without Can Reveal Secrets, do the same. The request is refused and the audit log has a denied reveal. Mounting only `certificate` works.
  3. A secret with file name `app_key` and an entry at `/run/secrets/app_key` refuse each other with `ERR_SETTING_MOUNT_PATH_TAKEN`.
  4. Export an app with an entry, and import it as that member. The entry with a private key is skipped with `SETTING_MOUNT_NOT_PERMITTED`, and the rest of the app is imported.
