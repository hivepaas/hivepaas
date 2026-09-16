# Configuration Spec Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Export HivePaaS configuration at any scope to a portable, human-readable `.tar.gz` bundle of YAML that can later be imported to provision the same configuration.

**Architecture:** A new `specservice` walks scopes, reads settings from the database and app configuration from the Docker Swarm service directly (never through usecase DTOs), applies a per-setting-type export policy registered beside the existing parser registry, and serializes to a deterministic YAML bundle. Secrets follow a three-mode switch reusing the existing reveal-and-audit path; `encrypted` mode wraps the whole bundle with `age`, the same way `sysbackupservice` does. Import is not built here, but the format is designed against a written import contract.

**Tech Stack:** Go 1.24+, `gopkg.in/yaml.v3`, `filearchiver` (tar.gz), `filippo.io/age`, `bun` (via existing repositories), Docker Swarm API types (`github.com/moby/moby/api/types/swarm`), testify for tests.

**Spec:** `docs/superpowers/specs/2026-09-16-config-spec-export-design.md`

## Global Constraints

- **Layering.** `ARCHITECTURE.md` places `dto` above `service`. `specservice` and `specmodel` MUST NOT import anything under `hivepaas_app/usecase/`. Test files may.
- **Export is read-only.** No task in this plan writes to the database or to Docker.
- **Every exported setting type needs a registered policy.** A type with no registered `SpecPolicy` is skipped and reported, never exported blindly.
- **Skipped setting types are exactly three:** `api-key`, `backup-snapshot`, `app`.
- **Strip only what is regenerated or transient.** System-specific values are kept. The complete strip list is five rules; see Task 5 and Task 6.
- **Determinism.** Two exports of an unchanged system must produce byte-identical YAML.
- **Error codes.** New errors go in `hperrors` as `NewErr(<base>, "ERR_…")`. The repo lints this with `tools/errcodelint`.
- **Formatting and linting.** `make fmt` then `make lint` must pass before every commit. `make test` runs the suite.
- **Spelling.** The repo's Go code is linted by `misspell` for US spelling. Docs are not.

---

## File Structure

**New packages**

| path | responsibility |
|---|---|
| `hivepaas_app/service/specservice/specmodel/` | the durable spec contract. Depends on `entity`, `base`, docker types. Nothing above. |
| ├ `manifest.go`, `issue.go` | manifest, report, severity, issue codes (Task 1) |
| ├ `key.go` | collection key derivation (Task 2) |
| ├ `deployment.go` | the five Swarm-derived blocks (Task 6) |
| ├ `singleton.go` | which types are singletons, and every block name (Task 7) |
| └ `document.go` | `GlobalDoc`, `ProjectDoc`, `EnvDoc`, `AppDoc`, `ExternalRef`, `Bundle` (Tasks 7-8) |
| `hivepaas_app/service/specservice/service.go` | the `Service` interface |
| `hivepaas_app/service/specservice/specserviceimpl/` | the work: `swarm_labels.go` (Task 5), `swarm_map.go` (Task 6), `assemble.go` (Task 7), `walk.go` (Task 8), `secrets.go` (Task 9), `serialize.go` (Task 10), `bundle.go` (Task 11) |
| `hivepaas_app/usecase/specuc/` + `specuc/specdto/` | the export operation: permission, audit, download response |
| `hivepaas_app/interface/api/handler/spechandler/` | transport |

**Modified**

| path | change |
|---|---|
| `hivepaas_app/entity/setting_spec.go` (new) | `SpecPolicy` interface + registry, beside `registerSettingParser` |
| `hivepaas_app/entity/setting_refmap.go` (new) | `RemapRefs` — the value-driven reference walk |
| `hivepaas_app/hperrors/errors_spec.go` (new) | spec error codes |
| `hivepaas_app/registry/provides.go` | wire the service, usecase and handler |
| `hivepaas_app/interface/api/server/router_*.go` | register the export routes |

---

## Task 1: Spec model foundations

**Files:**
- Create: `hivepaas_app/service/specservice/specmodel/manifest.go`
- Create: `hivepaas_app/service/specservice/specmodel/issue.go`
- Create: `hivepaas_app/hperrors/errors_spec.go`
- Test: `hivepaas_app/service/specservice/specmodel/manifest_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `specmodel.APIVersion`, `specmodel.KindSpec`, `specmodel.SecretsMode` (+ the three constants), `specmodel.Manifest`, `specmodel.Severity` (+ three constants), `specmodel.Issue`, `specmodel.Report`. `hperrors.ErrSpecAPIVersionUnsupported`, `ErrSpecSettingTypeUnclassified`, `ErrSpecMountTargetDuplicated`, `ErrSpecSecretsModeInvalid`.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specmodel/manifest_test.go
package specmodel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestManifestMarshalsWithStableFieldOrder(t *testing.T) {
	m := &Manifest{
		APIVersion:        APIVersion,
		Kind:              KindSpec,
		ExportedAt:        time.Date(2026, 9, 16, 10, 15, 0, 0, time.UTC),
		SourceAppVersion:  "v0.1.0",
		SourceVersionCode: "v000001",
		Scope:             "global",
		SecretsMode:       SecretsModeNone,
		Files:             []string{"global.yaml", "projects/project_a/project.yaml"},
	}

	out, err := yaml.Marshal(m)
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: hivepaas.com/v1
kind: Spec
exportedAt: 2026-09-16T10:15:00Z
sourceAppVersion: v0.1.0
sourceVersionCode: v000001
scope: global
secretsMode: none
files:
    - global.yaml
    - projects/project_a/project.yaml
`, string(out))
}

func TestSecretsModeIsValid(t *testing.T) {
	assert.True(t, SecretsModeNone.IsValid())
	assert.True(t, SecretsModeEncrypted.IsValid())
	assert.True(t, SecretsModePlaintext.IsValid())
	assert.False(t, SecretsMode("").IsValid())
	assert.False(t, SecretsMode("clear").IsValid())
}

func TestReportCountsBySeverity(t *testing.T) {
	r := &Report{}
	r.Add(Issue{Severity: SeverityFixable, Code: "REF_NOT_FOUND", Path: "a"})
	r.Add(Issue{Severity: SeveritySkipped, Code: "TYPE_SKIPPED", Path: "b"})
	r.Add(Issue{Severity: SeverityFixable, Code: "REF_NOT_FOUND", Path: "c"})

	assert.Equal(t, 3, len(r.Issues))
	assert.Equal(t, 2, r.CountBySeverity(SeverityFixable))
	assert.Equal(t, 1, r.CountBySeverity(SeveritySkipped))
	assert.Equal(t, 0, r.CountBySeverity(SeverityBlocked))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run TestManifest -v`
Expected: FAIL — the package does not compile, `undefined: Manifest`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specmodel/manifest.go

// Package specmodel holds the spec format itself - the manifest, the report and
// the types every payload file is built from.
//
// It is the durable half of the feature. A published bundle is read back by
// versions of HivePaaS that do not exist yet, so the shape here changes only
// with apiVersion, and it deliberately shares no types with the dashboard DTOs,
// which are free to move whenever the UI does.
package specmodel

import "time"

// APIVersion is the format version every bundle and payload file carries.
const APIVersion = "hivepaas.com/v1"

// Kind distinguishes a configuration-only bundle from anything added later - a
// snapshot carrying volume data, for one.
type Kind string

const KindSpec Kind = "Spec"

// SecretsMode decides what happens to the values settings keep encrypted.
type SecretsMode string

const (
	// SecretsModeNone omits every secret. It needs no capability and leaks
	// nothing, which is why it is the default.
	SecretsModeNone SecretsMode = "none"
	// SecretsModeEncrypted includes secrets and wraps the whole bundle with age
	// under a passphrase the operator supplies.
	SecretsModeEncrypted SecretsMode = "encrypted"
	// SecretsModePlaintext includes secrets in the clear.
	SecretsModePlaintext SecretsMode = "plaintext"
)

func (m SecretsMode) IsValid() bool {
	switch m {
	case SecretsModeNone, SecretsModeEncrypted, SecretsModePlaintext:
		return true
	default:
		return false
	}
}

// RevealsSecrets reports whether this mode decrypts anything, and therefore
// whether the export has to pass the secret-reveal capability gate.
func (m SecretsMode) RevealsSecrets() bool {
	return m == SecretsModeEncrypted || m == SecretsModePlaintext
}

// Manifest is spec.yaml at the root of every bundle.
type Manifest struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       Kind      `yaml:"kind"`
	ExportedAt time.Time `yaml:"exportedAt"`
	// SourceAppVersion is informational - which build produced the bundle.
	SourceAppVersion string `yaml:"sourceAppVersion"`
	// SourceVersionCode is what compatibility is judged on.
	SourceVersionCode string      `yaml:"sourceVersionCode"`
	Scope             string      `yaml:"scope"`
	SecretsMode       SecretsMode `yaml:"secretsMode"`
	// Files is every payload file in the bundle, in bundle order. It is also the
	// surface a partial import selects from.
	Files []string `yaml:"files"`
}
```

```go
// hivepaas_app/service/specservice/specmodel/issue.go
package specmodel

// Severity says what an issue does to the object it belongs to.
type Severity string

const (
	// SeverityFixable clears the offending part and keeps going.
	SeverityFixable Severity = "fixable"
	// SeveritySkipped drops one object and keeps the rest.
	SeveritySkipped Severity = "skipped"
	// SeverityBlocked stops everything.
	SeverityBlocked Severity = "blocked"
)

// Issue codes. Export emits the first three; the rest are the import contract's,
// declared here so both halves name the same things.
const (
	CodeTypeUnclassified   = "TYPE_UNCLASSIFIED"
	CodeTypeSkipped        = "TYPE_SKIPPED"
	CodePreviewAppSkipped  = "PREVIEW_APP_SKIPPED"
	CodeRefNotFound        = "REF_NOT_FOUND"
	CodeRefNotSelected     = "REF_NOT_SELECTED"
	CodeMountTargetDup     = "MOUNT_TARGET_DUPLICATED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Issue is one thing that did not go cleanly, written to be both read by a
// person and acted on by a UI.
type Issue struct {
	Severity Severity       `yaml:"severity"`
	Code     string         `yaml:"code"`
	Path     string         `yaml:"path"`
	Detail   map[string]any `yaml:"detail,omitempty"`
	// AvailableIn names the bundle file holding what the issue could not reach,
	// which is what separates "go and create this" from "select one more file".
	AvailableIn string `yaml:"availableIn,omitempty"`
	Action      string `yaml:"action,omitempty"`
	Hint        string `yaml:"hint,omitempty"`
}

// Report collects issues for one export or import run.
type Report struct {
	Issues []Issue `yaml:"issues,omitempty"`
}

func (r *Report) Add(issue Issue) {
	r.Issues = append(r.Issues, issue)
}

func (r *Report) CountBySeverity(s Severity) int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == s {
			count++
		}
	}
	return count
}

func (r *Report) HasBlocked() bool {
	return r.CountBySeverity(SeverityBlocked) > 0
}
```

```go
// hivepaas_app/hperrors/errors_spec.go
package hperrors

// Errors for configuration spec export and import
var (
	ErrSpecAPIVersionUnsupported   = NewErr(ErrNotAllowed, "ERR_SPEC_API_VERSION_UNSUPPORTED")
	ErrSpecSettingTypeUnclassified = NewErr(ErrInternal, "ERR_SPEC_SETTING_TYPE_UNCLASSIFIED")
	ErrSpecMountTargetDuplicated   = NewErr(ErrDataInvalid, "ERR_SPEC_MOUNT_TARGET_DUPLICATED")
	ErrSpecSecretsModeInvalid      = NewErr(ErrArgumentInvalid, "ERR_SPEC_SECRETS_MODE_INVALID")
	ErrSpecPassphraseRequired      = NewErr(ErrPreconditionRequired, "ERR_SPEC_PASSPHRASE_REQUIRED")
)
```

Before writing the file, confirm the base errors exist:
`grep -n "ErrDataInvalid\|ErrArgumentInvalid\|ErrPreconditionRequired\|ErrNotAllowed\|ErrInternal" hivepaas_app/hperrors/*.go`
If a base error has a different name, use the one that is there rather than adding a new base.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -v`
Expected: PASS, three tests.

If `TestManifestMarshalsWithStableFieldOrder` fails on indentation, adjust the expected string to what `yaml.v3` actually emits rather than fighting the encoder — the point of the test is field *order*, which follows struct order.

- [ ] **Step 5: Add the translations for the new error codes**

Each `ERR_…` code needs an entry in the translation TOMLs. Find them and follow the existing shape:

Run: `grep -rn "ERR_PROJECT_NOT_FOUND" --include="*.toml" .`

Add one line per new code to the same files, with an English message such as
`ERR_SPEC_MOUNT_TARGET_DUPLICATED = "Two volume mounts use the same target path '{{.Target}}'"`.

- [ ] **Step 6: Verify lint and commit**

```bash
make fmt && make lint
go test ./hivepaas_app/service/specservice/... ./hivepaas_app/hperrors/...
git add hivepaas_app/service/specservice/specmodel hivepaas_app/hperrors/errors_spec.go
git commit -m "feat(spec): add spec model foundations and error codes"
```

---

## Task 2: Setting key derivation

**Files:**
- Create: `hivepaas_app/service/specservice/specmodel/key.go`
- Test: `hivepaas_app/service/specservice/specmodel/key_test.go`

**Interfaces:**
- Consumes: `entity.Setting`.
- Produces: `specmodel.DeriveSettingKeys(settings []*entity.Setting) map[string]string` — setting ID to spec key.

**Why this is not `slugify`.** Slugifying destroys the distinctions that make two settings different. Measured on the five certificates in a real database: `*.dev.hivepaas.com` and `dev.hivepaas.com` both slugify to `dev_hivepaas_com`, and `*.localhost` and `localhost` both to `localhost`. Keys live inside YAML, which quotes anything, so the name is used verbatim.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specmodel/key_test.go
package specmodel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func cert(id, name, kind string, createdAt time.Time) *entity.Setting {
	return &entity.Setting{
		ID: id, Name: name, Kind: kind, Type: base.SettingTypeSSLCert,
		Scope: base.ObjectScopeGlobal, CreatedAt: createdAt,
	}
}

// The five certificates seeded into a real development installation. Three have
// unique names and must keep them untouched; only the genuinely ambiguous pair
// gains a suffix.
func TestDeriveSettingKeysOnRealCertificateNames(t *testing.T) {
	t0 := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 9, 15, 14, 27, 54, 0, time.UTC)

	keys := DeriveSettingKeys([]*entity.Setting{
		cert("01JAB9XED0GTXBSQDFVYAJ8WM2", "*.dev.hivepaas.com", "self-signed", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM1", "*.localhost", "letsencrypt", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM4", "dev.hivepaas.com", "self-signed", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM3", "localhost", "letsencrypt", t0),
		cert("01M2JQF72SWV5NCVATKD1WV536", "localhost", "self-signed", t1),
	})

	assert.Equal(t, "*.dev.hivepaas.com", keys["01JAB9XED0GTXBSQDFVYAJ8WM2"])
	assert.Equal(t, "*.localhost", keys["01JAB9XED0GTXBSQDFVYAJ8WM1"])
	assert.Equal(t, "dev.hivepaas.com", keys["01JAB9XED0GTXBSQDFVYAJ8WM4"])
	assert.Equal(t, "localhost@letsencrypt", keys["01JAB9XED0GTXBSQDFVYAJ8WM3"])
	assert.Equal(t, "localhost@self-signed", keys["01M2JQF72SWV5NCVATKD1WV536"])
}

// ssh-key, notification and basic-auth all carry an empty kind, so a duplicate
// name there cannot be broken by kind and falls through to an index.
func TestDeriveSettingKeysFallsThroughToIndexWhenKindIsEmpty(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(id, name string, at time.Time) *entity.Setting {
		return &entity.Setting{ID: id, Name: name, Type: base.SettingTypeSSHKey, CreatedAt: at}
	}

	keys := DeriveSettingKeys([]*entity.Setting{
		mk("k2", "deploy", t0.Add(time.Hour)),
		mk("k1", "deploy", t0),
		mk("k3", "deploy", t0.Add(2*time.Hour)),
	})

	// Ordered by created_at, so k1 is the unsuffixed one.
	assert.Equal(t, "deploy", keys["k1"])
	assert.Equal(t, "deploy#2", keys["k2"])
	assert.Equal(t, "deploy#3", keys["k3"])
}

// Two settings of different types never collide with each other.
func TestDeriveSettingKeysIsScopedByType(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	keys := DeriveSettingKeys([]*entity.Setting{
		{ID: "a", Name: "main", Type: base.SettingTypeSSHKey, CreatedAt: t0},
		{ID: "b", Name: "main", Type: base.SettingTypeSecret, CreatedAt: t0},
	})
	assert.Equal(t, "main", keys["a"])
	assert.Equal(t, "main", keys["b"])
}

// The same input must always produce the same keys, whatever order the database
// returned the rows in.
func TestDeriveSettingKeysIsDeterministic(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := &entity.Setting{ID: "z", Name: "dup", Type: base.SettingTypeSSHKey, CreatedAt: t0}
	b := &entity.Setting{ID: "y", Name: "dup", Type: base.SettingTypeSSHKey, CreatedAt: t0}

	forward := DeriveSettingKeys([]*entity.Setting{a, b})
	reversed := DeriveSettingKeys([]*entity.Setting{b, a})
	assert.Equal(t, forward, reversed)
	// Same created_at, so the id breaks the tie: "y" sorts before "z".
	assert.Equal(t, "dup", forward["y"])
	assert.Equal(t, "dup#2", forward["z"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run TestDeriveSettingKeys -v`
Expected: FAIL — `undefined: DeriveSettingKeys`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specmodel/key.go
package specmodel

import (
	"fmt"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// DeriveSettingKeys assigns every setting the key a spec refers to it by,
// returning setting id -> key.
//
// The name is used verbatim rather than slugified. Slugifying collapses
// distinctions that matter: `*.localhost` and `localhost` are different
// certificates and both slugify to `localhost`. Keys only ever appear inside
// YAML, which quotes whatever it has to.
//
// Duplicates are broken by kind first, because kind is already part of identity
// in this system - one default is allowed per (scope, type, kind) - and only
// then by an index.
//
// Ordering is by created_at then id so that two exports of unchanged data
// produce the same keys regardless of the order rows came back in.
func DeriveSettingKeys(settings []*entity.Setting) map[string]string {
	ordered := make([]*entity.Setting, len(settings))
	copy(ordered, settings)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
		}
		return ordered[i].ID < ordered[j].ID
	})

	nameCount := make(map[string]int, len(ordered))
	for _, s := range ordered {
		nameCount[string(s.Type)+"\x00"+s.Name]++
	}

	keys := make(map[string]string, len(ordered))
	taken := make(map[string]bool, len(ordered))

	for _, s := range ordered {
		candidate := s.Name
		if nameCount[string(s.Type)+"\x00"+s.Name] > 1 && s.Kind != "" {
			candidate = s.Name + "@" + s.Kind
		}

		key := candidate
		for i := 2; taken[string(s.Type)+"\x00"+key]; i++ {
			key = fmt.Sprintf("%s#%d", candidate, i)
		}

		taken[string(s.Type)+"\x00"+key] = true
		keys[s.ID] = key
	}
	return keys
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -v`
Expected: PASS, all key tests plus Task 1's.

- [ ] **Step 5: Prove the tests have teeth**

Temporarily change `candidate = s.Name + "@" + s.Kind` to `candidate = s.Name`.
Run the tests: `TestDeriveSettingKeysOnRealCertificateNames` must fail, because the two `localhost` certificates would become `localhost` and `localhost#2` instead of being distinguished by kind.
Revert the change and re-run to confirm green.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specmodel/key.go hivepaas_app/service/specservice/specmodel/key_test.go
git commit -m "feat(spec): derive setting keys from names without slugifying"
```

---

## Task 3: Reference remapping

**Files:**
- Create: `hivepaas_app/entity/setting_refmap.go`
- Test: `hivepaas_app/entity/setting_refmap_test.go`

**Interfaces:**
- Consumes: `entity.SettingData`, `entity.RefObjectIDs`.
- Produces: `entity.RemapRefs(data SettingData, mapping map[string]string) error`.

**Why by value and not by type.** References live in four different container shapes — `ObjectID`, `ObjectValue`, `ObjectIDSlice`, and bespoke structs with a bare `ID string` (`SystemBackupCloudStorage`, `SchedJobCommandOutputFileStorage`). A reflection walk keyed on the *type* `ObjectID` silently misses the fourth, and a missed reference becomes a dangling pointer with no error. `GetRefObjectIDs()` already reports exactly which identifier *values* are references, so the walk matches on value and the shape stops mattering. After remapping, calling `GetRefObjectIDs()` again and comparing proves nothing was missed.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/entity/setting_refmap_test.go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemapRefsRewritesObjectIDFields(t *testing.T) {
	data := &SSLCert{
		Domain:       "example.com",
		Provider:     ObjectID{ID: "old-provider"},
		AcmeProvider: ObjectID{ID: "old-acme"},
	}
	err := RemapRefs(data, map[string]string{
		"old-provider": "new-provider",
		"old-acme":     "new-acme",
	})
	assert.NoError(t, err)
	assert.Equal(t, "new-provider", data.Provider.ID)
	assert.Equal(t, "new-acme", data.AcmeProvider.ID)
	assert.Equal(t, "example.com", data.Domain, "non-reference fields must not move")
}

// SystemBackupCloudStorage is a bespoke struct holding a bare `ID string`, not
// an ObjectID. A type-keyed walk misses it; this is the regression guard.
func TestRemapRefsRewritesBareIDFieldsInBespokeStructs(t *testing.T) {
	data := &SystemBackup{
		CloudStorage: SystemBackupCloudStorage{
			ID:     "old-storage",
			Bucket: "backups",
		},
	}
	err := RemapRefs(data, map[string]string{"old-storage": "new-storage"})
	assert.NoError(t, err)
	assert.Equal(t, "new-storage", data.CloudStorage.ID)
	assert.Equal(t, "backups", data.CloudStorage.Bucket)
}

func TestRemapRefsRewritesSlicesAndNestedPointers(t *testing.T) {
	data := &AppRoutingSettings{
		Port: 8080,
		Domains: []*AppDomain{
			{Domain: "a.example.com", SSLCert: ObjectID{ID: "cert-1"}},
			{Domain: "b.example.com", SSLCert: ObjectID{ID: "cert-2"}},
		},
	}
	err := RemapRefs(data, map[string]string{"cert-1": "new-1", "cert-2": "new-2"})
	assert.NoError(t, err)
	assert.Equal(t, "new-1", data.Domains[0].SSLCert.ID)
	assert.Equal(t, "new-2", data.Domains[1].SSLCert.ID)
}

// An id with no mapping entry is left alone rather than blanked.
func TestRemapRefsLeavesUnmappedReferencesUntouched(t *testing.T) {
	data := &SSLCert{Provider: ObjectID{ID: "keep-me"}}
	assert.NoError(t, RemapRefs(data, map[string]string{"other": "new"}))
	assert.Equal(t, "keep-me", data.Provider.ID)
}

// The self-check is the whole safety property: if the walk fails to reach a
// reference, RemapRefs must say so rather than return a half-rewritten value.
func TestRemapRefsSelfCheckDetectsAMissedReference(t *testing.T) {
	data := &unreachableRefData{hidden: ObjectID{ID: "old"}}
	err := RemapRefs(data, map[string]string{"old": "new"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "old")
}
```

Add the fixture type in the same test file. It reports a reference that the walk
cannot reach, because the field is unexported:

```go
// unreachableRefData reports a reference held in an unexported field, which no
// reflection walk can write. It exists only to prove the self-check fires.
type unreachableRefData struct {
	hidden ObjectID
}

func (d *unreachableRefData) GetType() base.SettingType { return base.SettingTypeScript }
func (d *unreachableRefData) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{RefSettingIDs: []string{d.hidden.ID}}
}
func (d *unreachableRefData) GetResourceLinks(s *Setting) []*ResLink { return nil }
func (d *unreachableRefData) Migrate(s *Setting) (bool, error)       { return false, nil }
```

Import `"github.com/hivepaas/hivepaas/hivepaas_app/base"` in the test file for this.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/entity/ -run TestRemapRefs -v`
Expected: FAIL — `undefined: RemapRefs`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/entity/setting_refmap.go
package entity

import (
	"reflect"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// RemapRefs rewrites every reference a setting holds, from the identifiers it
// was exported with to the identifiers it is being imported as.
//
// It walks by value rather than by type. References are stored in at least four
// different shapes - ObjectID, ObjectValue, ObjectIDSlice, and bespoke structs
// carrying a bare `ID string` such as SystemBackupCloudStorage - and a walk that
// recognised only the declared types would silently skip the last group. What
// every shape has in common is the identifier itself, and GetRefObjectIDs
// already reports exactly which identifiers are references, so that is what the
// walk matches on.
//
// After rewriting, the references are read back and compared. A reference the
// walk could not reach fails here, loudly, instead of importing as a pointer
// into another installation's data.
func RemapRefs(data SettingData, mapping map[string]string) error {
	if data == nil || len(mapping) == 0 {
		return nil
	}

	before := collectRefIDs(data)
	remapStringValues(reflect.ValueOf(data), mapping)

	want := make(map[string]bool, len(before))
	for _, id := range before {
		want[gofn.Coalesce(mapping[id], id)] = true
	}

	for _, id := range collectRefIDs(data) {
		if !want[id] {
			return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog(
				"reference %q was not rewritten by RemapRefs; the setting type holds it "+
					"somewhere the walk cannot reach", id)
		}
	}
	return nil
}

func collectRefIDs(data SettingData) []string {
	refIDs := data.GetRefObjectIDs()
	if refIDs == nil {
		return nil
	}
	all := make([]string, 0,
		len(refIDs.RefSettingIDs)+len(refIDs.RefAppIDs)+len(refIDs.RefProjectIDs)+
			len(refIDs.RefProjectEnvIDs)+len(refIDs.RefUserIDs))
	all = append(all, refIDs.RefSettingIDs...)
	all = append(all, refIDs.RefAppIDs...)
	all = append(all, refIDs.RefProjectIDs...)
	all = append(all, refIDs.RefProjectEnvIDs...)
	all = append(all, refIDs.RefUserIDs...)
	return all
}

// remapStringValues replaces every settable string equal to a mapping key.
//
// The shape of this walk follows reencryptValue in setting_reencrypt.go, which
// solved the same problem for encrypted fields.
func remapStringValues(value reflect.Value, mapping map[string]string) {
	switch value.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
		if !value.IsNil() {
			remapStringValues(value.Elem(), mapping)
		}

	case reflect.Struct:
		for i := range value.NumField() {
			if !value.Type().Field(i).IsExported() {
				continue // an unexported field cannot hold data we serialize
			}
			remapStringValues(value.Field(i), mapping)
		}

	case reflect.Slice, reflect.Array:
		for i := range value.Len() {
			remapStringValues(value.Index(i), mapping)
		}

	case reflect.Map:
		remapMapValues(value, mapping)

	case reflect.String:
		if !value.CanSet() {
			return
		}
		if replacement, ok := mapping[value.String()]; ok {
			value.SetString(replacement)
		}
	}
}

// remapMapValues handles maps, whose values are not addressable and so have to
// be copied out, rewritten and put back.
func remapMapValues(value reflect.Value, mapping map[string]string) {
	if value.IsNil() {
		return
	}
	for _, key := range value.MapKeys() {
		entry := reflect.New(value.Type().Elem()).Elem()
		entry.Set(value.MapIndex(key))
		remapStringValues(entry, mapping)
		value.SetMapIndex(key, entry)
	}
}
```

If `WithMsgLog` does not accept a format string, check its signature with
`grep -n "func.*WithMsgLog" hivepaas_app/hperrors/*.go` and adapt the call.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/entity/ -run TestRemapRefs -v`
Expected: PASS, five tests.

- [ ] **Step 5: Run the full entity suite**

Run: `go test ./hivepaas_app/entity/...`
Expected: PASS. `RemapRefs` touches nothing existing, so a failure here means the new file broke compilation.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
git add hivepaas_app/entity/setting_refmap.go hivepaas_app/entity/setting_refmap_test.go
git commit -m "feat(spec): remap setting references by value with a self-check"
```

---

## Task 4: Per-type export policy registry

**Files:**
- Create: `hivepaas_app/entity/setting_spec.go`
- Test: `hivepaas_app/entity/setting_spec_test.go`

**Interfaces:**
- Consumes: `entity.Setting`, `entity.SettingData`, `base.SettingType`, `base.ObjectScopeType`.
- Produces: `entity.SpecDecision`, `entity.SpecPolicy`, `entity.SpecPolicyFor(base.SettingType) SpecPolicy`, `entity.SpecExportDecision(*Setting) SpecDecision`.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/entity/setting_spec_test.go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestSpecExportDecisionSkipsTheThreeSkippedTypes(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeAPIKey,
		base.SettingTypeBackupSnapshot,
		base.SettingTypeApp,
	} {
		d := SpecExportDecision(&Setting{Type: typ})
		assert.False(t, d.Export, "%v must not be exported", typ)
		assert.NotEmpty(t, d.Reason, "%v must say why", typ)
	}
}

func TestSpecExportDecisionAllowsAnOrdinarySetting(t *testing.T) {
	d := SpecExportDecision(&Setting{Type: base.SettingTypeSSLCert})
	assert.True(t, d.Export)
}

// cluster-* rows created by sync live at global scope and carry a Docker object
// id; rows HivePaaS authored live at project or project-env scope and carry a
// ULID. Only the authored ones are configuration.
func TestSpecExportDecisionKeepsOnlyAuthoredClusterRows(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeClusterNetwork,
		base.SettingTypeClusterVolume,
		base.SettingTypeClusterNode,
	} {
		discovered := SpecExportDecision(&Setting{Type: typ, Scope: base.ObjectScopeGlobal})
		assert.False(t, discovered.Export, "%v at global scope is sync discovery", typ)

		authored := SpecExportDecision(&Setting{
			Type: typ, Scope: base.ObjectScopeProject, ObjectID: "prj_1",
		})
		assert.True(t, authored.Export, "%v at project scope is configuration", typ)
	}
}

// Every setting type in base must have a policy. A type added later without one
// fails here rather than silently leaking into a spec.
func TestEverySettingTypeHasASpecPolicy(t *testing.T) {
	for typ := range settingParserMap {
		assert.NotNil(t, SpecPolicyFor(typ), "setting type %v has no registered SpecPolicy", typ)
	}
}

func TestSpecPolicyForAnUnknownTypeIsNil(t *testing.T) {
	assert.Nil(t, SpecPolicyFor(base.SettingType("not-a-real-type")))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/entity/ -run "TestSpec|TestEverySettingType" -v`
Expected: FAIL — `undefined: SpecExportDecision`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/entity/setting_spec.go
package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

// SpecDecision is a policy's answer about one setting.
type SpecDecision struct {
	Export bool
	// Reason explains a refusal, and is what the export report prints.
	Reason string
}

// SpecPolicy decides how one setting type appears in a configuration spec.
//
// It is registered beside the parser, and for the same reason: a setting type
// that gains a policy here cannot be forgotten by the exporter, and one that
// forgets to register is refused rather than exported blindly.
type SpecPolicy interface {
	// Decide reports whether this particular setting belongs in a spec.
	Decide(setting *Setting) SpecDecision
	// Strip removes values that the target regenerates or that would trigger an
	// action on import. It must not remove values that are merely specific to
	// the installation being exported: those are what make a restore exact.
	Strip(data SettingData)
}

var specPolicyMap = make(map[base.SettingType]SpecPolicy, 46) //nolint:mnd

func registerSpecPolicy(typ base.SettingType, policy SpecPolicy) bool {
	specPolicyMap[typ] = policy
	return true
}

func registerDefaultSpecPolicies(types ...base.SettingType) bool {
	for _, typ := range types {
		specPolicyMap[typ] = defaultSpecPolicy{}
	}
	return true
}

// SpecPolicyFor returns the policy for a type, or nil if none is registered.
func SpecPolicyFor(typ base.SettingType) SpecPolicy {
	return specPolicyMap[typ]
}

// SpecExportDecision is the exporter's entry point. An unregistered type is
// refused, so adding a setting type without thinking about the spec produces a
// report entry rather than a silent leak.
func SpecExportDecision(setting *Setting) SpecDecision {
	policy := SpecPolicyFor(setting.Type)
	if policy == nil {
		return SpecDecision{Reason: "no spec policy is registered for this setting type"}
	}
	return policy.Decide(setting)
}

// defaultSpecPolicy exports the setting whole and strips nothing.
type defaultSpecPolicy struct{}

func (defaultSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }
func (defaultSpecPolicy) Strip(SettingData)            {}

// skipSpecPolicy never exports.
type skipSpecPolicy struct{ reason string }

func (p skipSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Reason: p.reason} }
func (skipSpecPolicy) Strip(SettingData)              {}

// clusterSpecPolicy exports only the rows HivePaaS authored.
//
// Every cluster type exists in two flavours. networks_sync.go, volumes_sync.go
// and docker_node_sync.go all hardcode ObjectScopeGlobal and use the Docker
// object id as the setting id; those rows are rediscovered by the next sync and
// mean nothing on another installation. The rows created by
// networkservice/project_env.go and volumeservice/project.go sit at project or
// project-env scope with a real ULID, and record which project owns what -
// which no sync can reconstruct, because no sync ever writes a non-global scope.
type clusterSpecPolicy struct{}

func (clusterSpecPolicy) Decide(setting *Setting) SpecDecision {
	if setting.Scope == base.ObjectScopeGlobal {
		return SpecDecision{Reason: "discovered by cluster sync; recreated automatically"}
	}
	return SpecDecision{Export: true}
}

func (clusterSpecPolicy) Strip(SettingData) {}

// appRoutingSpecPolicy clears the reset flag, which is a command rather than
// configuration: importing Reset: true performs a reset.
type appRoutingSpecPolicy struct{}

func (appRoutingSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }

func (appRoutingSpecPolicy) Strip(data SettingData) {
	if routing, ok := data.(*AppRoutingSettings); ok {
		routing.Reset = false
	}
}

// loggingSpecPolicy drops the endpoints of a managed backend. The setting's own
// comment says they are "derived from the deployed service when Managed, and
// ignored", so carrying them makes every spec disagree with the system.
type loggingSpecPolicy struct{}

func (loggingSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }

func (loggingSpecPolicy) Strip(data SettingData) {
	logging, ok := data.(*LoggingSettings)
	if !ok || !logging.Backend.Managed {
		return
	}
	logging.Backend.Ingest = nil
	logging.Backend.Query = nil
}

//nolint:gochecknoinits // registration mirrors registerSettingParser
var (
	_ = registerSpecPolicy(base.SettingTypeAPIKey, skipSpecPolicy{
		reason: "the secret is a one-way hash and authenticates nobody once imported"})
	_ = registerSpecPolicy(base.SettingTypeBackupSnapshot, skipSpecPolicy{
		reason: "backup history, rediscovered by scanning the repository"})
	_ = registerSpecPolicy(base.SettingTypeApp, skipSpecPolicy{
		reason: "a declared type with no parser and no rows"})

	_ = registerSpecPolicy(base.SettingTypeClusterNetwork, clusterSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeClusterVolume, clusterSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeClusterNode, clusterSpecPolicy{})

	_ = registerSpecPolicy(base.SettingTypeAppRouting, appRoutingSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeLogging, loggingSpecPolicy{})

	// Everything else is exported whole. Keeping this list explicit rather than
	// defaulting to "export" is what makes a new setting type show up in the
	// report instead of appearing in specs unnoticed.
	_ = registerDefaultSpecPolicies(
		base.SettingTypeAccessToken,
		base.SettingTypeAcmeDnsProvider,
		base.SettingTypeAppClone,
		base.SettingTypeAppDeployment,
		base.SettingTypeAppFeatures,
		base.SettingTypeAppKind,
		base.SettingTypeAppPlacement,
		base.SettingTypeBackupRepo,
		base.SettingTypeBackupRepoCleanup,
		base.SettingTypeBasicAuth,
		base.SettingTypeCloudStorage,
		base.SettingTypeCommandPipe,
		base.SettingTypeCommandTemplate,
		base.SettingTypeConfigFile,
		base.SettingTypeDomainSettings,
		base.SettingTypeEmail,
		base.SettingTypeEnvVar,
		base.SettingTypeGithubApp,
		base.SettingTypeHivePaaSService,
		base.SettingTypeIMService,
		base.SettingTypeImageBuild,
		base.SettingTypeNotification,
		base.SettingTypeOAuth,
		base.SettingTypePeriodicJob,
		base.SettingTypeProject,
		base.SettingTypeRegistryAuth,
		base.SettingTypeRepoWebhook,
		base.SettingTypeSSHKey,
		base.SettingTypeSSLCert,
		base.SettingTypeSSLProvider,
		base.SettingTypeSSLRenewal,
		base.SettingTypeSchedJob,
		base.SettingTypeScript,
		base.SettingTypeSecret,
		base.SettingTypeSystemBackup,
		base.SettingTypeSystemCleanup,
		base.SettingTypeTraefikConfig,
		base.SettingTypeTraefikService,
	)
)
```

Note there is deliberately **no** policy stripping `Secret.SwarmRef`,
`ConfigFile.SwarmRef`, `env-var` `IsSystem` entries, the image digest, or
`cluster-volume.NodeID`. Each is specific to the installation rather than
regenerated, and each is needed for an exact restore. The `env-var` case is the
sharpest: `HIVEPAAS_ROOT_PASSWORD` is generated once and the database volume was
initialised with it, so regenerating it on restore leaves the application
holding a password the database will not accept.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/entity/ -run "TestSpec|TestEverySettingType" -v`
Expected: PASS. If `TestEverySettingTypeHasASpecPolicy` fails, the message names
the missing type — add it to `registerDefaultSpecPolicies` or give it a policy.

- [ ] **Step 5: Add the strip tests**

```go
func TestAppRoutingStripClearsTheResetFlag(t *testing.T) {
	data := &AppRoutingSettings{Port: 8080, Reset: true}
	SpecPolicyFor(base.SettingTypeAppRouting).Strip(data)
	assert.False(t, data.Reset)
	assert.Equal(t, 8080, data.Port)
}

func TestLoggingStripDropsManagedEndpointsOnly(t *testing.T) {
	managed := &LoggingSettings{Backend: LoggingBackend{
		Managed: true,
		Ingest:  &LoggingEndpoint{},
		Query:   &LoggingEndpoint{},
	}}
	SpecPolicyFor(base.SettingTypeLogging).Strip(managed)
	assert.Nil(t, managed.Backend.Ingest)
	assert.Nil(t, managed.Backend.Query)

	external := &LoggingSettings{Backend: LoggingBackend{
		Managed: false,
		Ingest:  &LoggingEndpoint{},
	}}
	SpecPolicyFor(base.SettingTypeLogging).Strip(external)
	assert.NotNil(t, external.Backend.Ingest, "an unmanaged backend's endpoints are configuration")
}

// The values deliberately kept. This test is the guard against somebody
// "tidying up" a system-specific value later.
func TestStripKeepsSystemSpecificValues(t *testing.T) {
	secret := &Secret{Key: "db", SwarmRef: &SwarmSecretRef{SecretID: "swarm-1"}}
	SpecPolicyFor(base.SettingTypeSecret).Strip(secret)
	assert.NotNil(t, secret.SwarmRef, "SwarmRef resolves on the same installation")

	envVars := &EnvVars{Data: []*EnvVar{
		{Key: "HIVEPAAS_ROOT_PASSWORD", Value: "generated", IsSystem: true},
		{Key: "APP_ENV", Value: "production"},
	}}
	SpecPolicyFor(base.SettingTypeEnvVar).Strip(envVars)
	assert.Len(t, envVars.Data, 2,
		"system env vars must survive: the database volume was initialised with them")
}
```

Check `LoggingEndpoint` is the real type name before writing this:
`grep -n "type LoggingEndpoint" hivepaas_app/entity/setting_logging.go`

Run: `go test ./hivepaas_app/entity/ -run "TestAppRouting|TestLogging|TestStrip" -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
git add hivepaas_app/entity/setting_spec.go hivepaas_app/entity/setting_spec_test.go
git commit -m "feat(spec): register a per-type export policy beside the parser registry"
```

---

## Task 5: Label denylist and placement constraint filter

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/swarm_labels.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/swarm_labels_test.go`

**Interfaces:**
- Consumes: nothing beyond the standard library.
- Produces: `filterUserLabels(labels map[string]string) map[string]string`, `filterUserConstraints(constraints []string, managedLabel string) []string`, and the constant `labelAppPlacementConstraints = "hivepaas.app.placementConstraints"`.

This is its own task because both functions are small, pure, and carry a high cost when wrong — a leaked Docker Desktop label publishes an absolute host path, and a leaked placement constraint pins an imported app to a node that does not exist.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/swarm_labels_test.go
package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterUserLabelsRemovesEverythingRegenerated(t *testing.T) {
	kept := filterUserLabels(map[string]string{
		// HivePaaS rewrites these when the service is applied
		"hivepaas.app.info":                 `{"name":"a1"}`,
		"hivepaas.app.id":                   "app_1",
		"hivepaas.app.placementConstraints": "node.role==manager",
		// Docker's own
		"com.docker.stack.namespace": "p1",
		// Docker Desktop injects these, and they carry absolute host paths
		"desktop.docker.io/mounts/0/Source": "/Users/tnt/go/src/github.com/hivepaas/hivepaas/.appdata",
		"desktop.docker.io/mounts/0/Target": "/var/lib/hivepaas",
		// Traefik regenerates its own labels from routing settings
		"traefik.http.routers.app.rule": "Host(`x.com`)",
		// except the ones a user wrote by hand
		"traefik.http.routers.x-custom-router-acme.entrypoints": "web",
		// and anything that is plainly the user's
		"team":        "platform",
		"cost-center": "eng",
	})

	assert.Equal(t, map[string]string{
		"traefik.http.routers.x-custom-router-acme.entrypoints": "web",
		"team":        "platform",
		"cost-center": "eng",
	}, kept)
}

func TestFilterUserLabelsHandlesNil(t *testing.T) {
	assert.Empty(t, filterUserLabels(nil))
}

func TestFilterUserConstraintsRemovesOnlyTheOnesHivePaaSAdded(t *testing.T) {
	kept := filterUserConstraints(
		[]string{
			"node.role == manager",
			"node.labels.hivepaas.role == control-plane",
			"node.labels.zone == eu",
		},
		"node.role==manager,node.labels.hivepaas.role==control-plane",
	)
	assert.Equal(t, []string{"node.labels.zone == eu"}, kept)
}

// With no managed-constraints label, every constraint is the user's.
func TestFilterUserConstraintsKeepsAllWhenNothingWasManaged(t *testing.T) {
	kept := filterUserConstraints([]string{"node.labels.zone == eu"}, "")
	assert.Equal(t, []string{"node.labels.zone == eu"}, kept)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -v`
Expected: FAIL — `undefined: filterUserLabels`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specserviceimpl/swarm_labels.go
package specserviceimpl

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// labelAppPlacementConstraints is where HivePaaS records the placement
// constraints it added itself, so that the user's can be told apart from them.
// It must match placementserviceimpl's constant of the same value.
const labelAppPlacementConstraints = "hivepaas.app.placementConstraints"

// managedLabelPrefixes are rewritten whenever the service is applied, so a spec
// that carried them would disagree with the system the moment it was imported.
//
// This is a maintained denylist rather than a single prefix check because third
// parties write labels onto services too: desktop.docker.io is Docker Desktop's,
// not HivePaaS's, and it embeds absolute paths from the machine that ran the
// export. The next such writer will not be called hivepaas either.
var managedLabelPrefixes = []string{
	"hivepaas.",
	"com.docker.stack.",
	"desktop.docker.io/",
}

// traefikCustomMarker marks the Traefik labels a user wrote by hand. Traefik
// regenerates every other traefik.* label from the routing settings - this is
// the same rule updateSwarmServiceLabels uses when it cleans up its own.
const traefikCustomMarker = ".x-custom-"

// filterUserLabels keeps only the labels a person put there.
func filterUserLabels(labels map[string]string) map[string]string {
	kept := make(map[string]string, len(labels))
	for key, value := range labels {
		if isManagedLabel(key) {
			continue
		}
		kept[key] = value
	}
	return kept
}

func isManagedLabel(key string) bool {
	for _, prefix := range managedLabelPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	if strings.HasPrefix(key, "traefik.") && !strings.Contains(key, traefikCustomMarker) {
		return true
	}
	return false
}

// filterUserConstraints removes the placement constraints HivePaaS derived from
// app-placement settings and from volume node pinning, leaving the user's.
//
// managed is the raw value of the labelAppPlacementConstraints label. The
// normalization has to match placement_apply.go's, which compares constraints
// with their spacing removed.
func filterUserConstraints(constraints []string, managed string) []string {
	if len(constraints) == 0 {
		return nil
	}

	managedSet := make([]string, 0, 4) //nolint:mnd
	for _, item := range strings.Split(managed, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			managedSet = append(managedSet, normalizeConstraint(trimmed))
		}
	}

	kept := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		if gofn.Contain(managedSet, normalizeConstraint(constraint)) {
			continue
		}
		kept = append(kept, constraint)
	}
	if len(kept) == 0 {
		return nil
	}
	return kept
}

func normalizeConstraint(constraint string) string {
	k, op, v := dockerhelper.ParsePlacementConstraint(constraint)
	if op == "" {
		return strings.TrimSpace(constraint)
	}
	return k + op + v
}
```

Confirm the helper's import path and signature first:
`grep -rn "func ParsePlacementConstraint" --include="*.go" .`
and confirm the constant matches:
`grep -n "labelAppPlacementConstraints" hivepaas_app/service/placementservice/placementserviceimpl/placement_apply.go`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -v`
Expected: PASS, four tests.

- [ ] **Step 5: Prove the denylist has teeth**

Temporarily remove `"desktop.docker.io/"` from `managedLabelPrefixes`.
Run the tests: `TestFilterUserLabelsRemovesEverythingRegenerated` must fail, showing the host path would have been exported.
Restore it and re-run.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specserviceimpl/swarm_labels.go hivepaas_app/service/specservice/specserviceimpl/swarm_labels_test.go
git commit -m "feat(spec): filter managed labels and derived placement constraints"
```

---

## Task 6: Swarm service to spec mapping

**Files:**
- Create: `hivepaas_app/service/specservice/specmodel/deployment.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/swarm_map.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/coverage_oracle_test.go`

**Interfaces:**
- Consumes: `filterUserLabels`, `filterUserConstraints` (Task 5); `swarm.Service`.
- Produces: `specmodel.Deployment` with fields `Source`, `Container`, `Resources`, `Storage`, `Networks`, `Service`; and `mapSwarmService(svc *swarm.Service, netNames map[string]string) (*specmodel.Deployment, error)`.

**Defining the spec types.** Mirror the field sets from the DTOs, but as our own
types in `specmodel`, with `yaml` tags. Copy each struct field-for-field from the
file named below, changing `json:` tags to `yaml:` and applying the listed
differences. Do not import the DTO packages.

| specmodel type | copy fields from | differences |
|---|---|---|
| `Container` | `appsettingsdto/container_settings_get.go:BaseContainerSettings` | `ServiceLabels`/`ContainerLabels` pass through `filterUserLabels`; drop `Command` and `WorkingDir` (they live in `Source`) |
| `Resources` | `appsettingsdto/resource_settings_get.go:ResourceSettingsResp` | drop `UpdateVer` |
| `Storage` | `appsettingsdto/storage_settings_get.go:Mount` | `Mounts` becomes `map[string]Mount` keyed by target; drop `Key` and `Target` from the value |
| `Networks` | `appsettingsdto/network_settings_get.go:NetworkSettingsResp` | `NetworkAttachment` drops `ID` and keeps `Name`; drop `UpdateVer` |
| `Service` | `appsettingsdto/service_settings_get.go:ServiceSettingsResp` | `Placement.Constraints` passes through `filterUserConstraints`; drop `UpdateVer` |

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go
package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func testService() *swarm.Service {
	return &swarm.Service{
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Labels: map[string]string{
					"hivepaas.app.info":                 `{"name":"a1"}`,
					"hivepaas.app.placementConstraints": "node.role==manager",
					"team":                              "platform",
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image: "ghcr.io/acme/api:1.4.2@sha256:" +
						"6ecdf4e6779ce1a655b53dcbab6d8920d971d7e86aab9a2186ae43fbc9df4dfb",
					Mounts: []mount.Mount{
						{Type: mount.TypeVolume, Source: "vol_1", Target: "/var/lib/postgresql/data"},
						{Type: mount.TypeBind, Source: "/srv/conf", Target: "/etc/app/config"},
					},
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{Target: "8vo4p3pwm1aksdu2ilryn8mpf", Aliases: []string{"api"}},
				},
				Placement: &swarm.Placement{
					Constraints: []string{"node.role == manager", "node.labels.zone == eu"},
					Platforms:   []swarm.Platform{{Architecture: "amd64", OS: "linux"}},
				},
			},
		},
	}
}

func TestMapSwarmServiceStripsTheImageDigest(t *testing.T) {
	out, err := mapSwarmService(testService(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "ghcr.io/acme/api:1.4.2", out.Container.Image,
		"a digest pins an image that may not exist in the target registry")
}

func TestMapSwarmServiceKeysMountsByTarget(t *testing.T) {
	out, err := mapSwarmService(testService(), nil)
	assert.NoError(t, err)

	assert.Len(t, out.Storage.Mounts, 2)
	assert.Equal(t, "vol_1", out.Storage.Mounts["/var/lib/postgresql/data"].Source)
	assert.Equal(t, "/srv/conf", out.Storage.Mounts["/etc/app/config"].Source)
}

// Nothing upstream validates that two mounts do not share a target, so this is
// the only thing holding the invariant a future snapshot will pin data to.
func TestMapSwarmServiceRefusesDuplicateMountTargets(t *testing.T) {
	svc := testService()
	svc.Spec.TaskTemplate.ContainerSpec.Mounts = []mount.Mount{
		{Type: mount.TypeVolume, Source: "vol_1", Target: "/data"},
		{Type: mount.TypeVolume, Source: "vol_2", Target: "/data"},
	}
	_, err := mapSwarmService(svc, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "/data")
}

func TestMapSwarmServiceResolvesNetworkNames(t *testing.T) {
	out, err := mapSwarmService(testService(), map[string]string{
		"8vo4p3pwm1aksdu2ilryn8mpf": "p1_dev_net",
	})
	assert.NoError(t, err)
	assert.Len(t, out.Networks.Attachments, 1)
	assert.Equal(t, "p1_dev_net", out.Networks.Attachments[0].Name)
	assert.Equal(t, []string{"api"}, out.Networks.Attachments[0].Aliases)
}

// An id with no name is a network the exporter could not resolve. It is kept as
// the id rather than dropped, so import can report it rather than lose the
// attachment silently.
func TestMapSwarmServiceKeepsUnresolvedNetworkIDs(t *testing.T) {
	out, err := mapSwarmService(testService(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "8vo4p3pwm1aksdu2ilryn8mpf", out.Networks.Attachments[0].Name)
}

func TestMapSwarmServiceDropsDerivedPlacement(t *testing.T) {
	out, err := mapSwarmService(testService(), nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"node.labels.zone == eu"}, out.Service.Placement.Constraints,
		"node.role == manager was added by HivePaaS and is regenerated")
}

func TestMapSwarmServiceKeepsOnlyUserLabels(t *testing.T) {
	out, err := mapSwarmService(testService(), nil)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"team": "platform"}, out.Container.ServiceLabels)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestMapSwarmService -v`
Expected: FAIL — `undefined: mapSwarmService`.

- [ ] **Step 3: Write the spec types**

Create `specmodel/deployment.go` with the five block types, following the table
above. The parts with real logic, written out:

```go
// hivepaas_app/service/specservice/specmodel/deployment.go
package specmodel

// Deployment is everything needed to recreate an app's running form: the
// app-deployment setting, plus the parts of the Swarm service HivePaaS treats
// as configuration.
//
// The Swarm-derived blocks are optional as a group. An app that has never been
// deployed has no service to read, and exports its settings alone.
type Deployment struct {
	Source    *Source    `yaml:"source,omitempty"`
	Container *Container `yaml:"container,omitempty"`
	Resources *Resources `yaml:"resources,omitempty"`
	Storage   *Storage   `yaml:"storage,omitempty"`
	Networks  *Networks  `yaml:"networks,omitempty"`
	Service   *Service   `yaml:"service,omitempty"`
}

// Storage keys mounts by their target path rather than by list position.
//
// A positional identity would be adequate while a spec carries only
// configuration - a reordered mount list re-imports the same either way. It
// stops being adequate as soon as anything outside the spec points at a mount,
// which is what a snapshot does when it pins volume data. Data pinned to
// mounts[0], a reorder, and a restore is data attached to the wrong volume,
// with no error.
//
// Two mounts at one target is meaningless, which makes the target a key - but
// nothing upstream validates it, so mapSwarmService checks and refuses.
type Storage struct {
	Mounts map[string]Mount `yaml:"mounts,omitempty"`
}

// NetworkAttachment carries the network's name. The Swarm spec stores the id in
// Networks[].Target, which means nothing on another installation and nothing to
// a reader.
type NetworkAttachment struct {
	Name    string   `yaml:"name"`
	Aliases []string `yaml:"aliases,omitempty"`
}
```

Write the remaining types (`Source`, `Container`, `Resources`, `Mount`,
`Networks`, `Service`, `Placement`, and the nested `Privileges`, `Healthcheck`,
`RestartPolicy`, `LogDriver`, `Capabilities`, `Ulimit`, `ResourceLimits`,
`ResourceReservations`, `Memory`, `GenericResource`, `BindOptions`,
`VolumeOptions`, `TmpfsOptions`, `ClusterOptions`, `DNSConfig`,
`HostsFileEntry`, `EndpointSpec`, `PortConfig`, `ServiceModeSpec`,
`PlacementConstraint`, `PlacementPreference`) by copying field-for-field from
the DTO files named in the table, with `yaml` tags.

- [ ] **Step 3b: Commit the types before writing the mapping**

The spec types are mechanical but numerous, and a reviewer cannot meaningfully
check twenty-five struct definitions in the same diff as the logic that fills
them. Commit them on their own:

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specmodel/deployment.go
git commit -m "feat(spec): add spec types for the swarm-derived deployment blocks"
```

- [ ] **Step 4: Write the mapping**

```go
// hivepaas_app/service/specservice/specserviceimpl/swarm_map.go
package specserviceimpl

import (
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// mapSwarmService turns a live Swarm service into the declarative part of it.
//
// It reads swarm.Service directly rather than calling appsettingsdto.Transform*.
// Those DTOs are the dashboard's wire shape and are free to change when the UI
// does; a spec is a durable artifact read back by versions that do not exist
// yet, and must not move underneath itself for a reason that has nothing to do
// with it. ARCHITECTURE.md also places dto above service, so the dependency
// would run the wrong way.
//
// netNames maps Docker network id to name; ids with no entry are kept as-is so
// that import can report an unresolved attachment rather than lose it.
func mapSwarmService(svc *swarm.Service, netNames map[string]string) (*specmodel.Deployment, error) {
	if svc == nil {
		return nil, nil
	}
	spec := &svc.Spec
	task := &spec.TaskTemplate

	out := &specmodel.Deployment{}

	if cs := task.ContainerSpec; cs != nil {
		container, err := mapContainer(cs, spec.Labels)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.Container = container

		storage, err := mapStorage(cs)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.Storage = storage
	}

	out.Resources = mapResources(task.Resources)
	out.Networks = mapNetworks(task.Networks, spec.EndpointSpec, netNames)
	out.Service = mapService(spec, task)

	return out, nil
}

// stripImageDigest removes the @sha256:… suffix docker stack deploy resolves an
// image to. It pins a digest that may not exist in the target registry and
// defeats the meaning of a moving tag. Services HivePaaS creates rarely carry
// one, since HivePaaS controls options.QueryRegistry - but it can be enabled.
func stripImageDigest(image string) string {
	if at := strings.LastIndex(image, "@"); at > 0 {
		return image[:at]
	}
	return image
}

func mapStorage(cs *swarm.ContainerSpec) (*specmodel.Storage, error) {
	if len(cs.Mounts) == 0 {
		return nil, nil
	}
	mounts := make(map[string]specmodel.Mount, len(cs.Mounts))
	for i := range cs.Mounts {
		m := &cs.Mounts[i]
		if _, exists := mounts[m.Target]; exists {
			return nil, hperrors.Wrap(hperrors.ErrSpecMountTargetDuplicated).
				WithParam("Target", m.Target)
		}
		mounts[m.Target] = mapMount(m)
	}
	return &specmodel.Storage{Mounts: mounts}, nil
}

func mapService(spec *swarm.ServiceSpec, task *swarm.TaskSpec) *specmodel.Service {
	out := &specmodel.Service{ModeSpec: mapServiceMode(spec.Mode)}
	if task.Placement == nil {
		return out
	}
	// Platforms is detected by Docker at runtime and has no place in a
	// declarative document, so it is simply not mapped.
	constraints := filterUserConstraints(
		task.Placement.Constraints, spec.Labels[labelAppPlacementConstraints])
	if len(constraints) > 0 || len(task.Placement.Preferences) > 0 {
		out.Placement = &specmodel.Placement{
			Constraints: constraints,
			Preferences: mapPlacementPreferences(task.Placement.Preferences),
		}
	}
	return out
}
```

Write `mapContainer`, `mapResources`, `mapNetworks`, `mapMount`,
`mapServiceMode` and `mapPlacementPreferences` as straight field copies.
`mapContainer` must run `spec.Labels` and `cs.Labels` through `filterUserLabels`
and `cs.Image` through `stripImageDigest`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v`
Expected: PASS, all seven mapping tests plus Tasks 1, 2 and 5.

- [ ] **Step 6: Write the field-coverage oracle**

Owning the field list gives up the guarantee the DTOs provided for free: Get and
Update carry identical field sets, so anything readable was writable. This test
recovers it. Test code may import the usecase layer — a test is not a layer.

```go
// hivepaas_app/service/specservice/specserviceimpl/coverage_oracle_test.go
package specserviceimpl

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// fieldNames returns the yaml or json names of a struct's exported fields.
func fieldNames(t *testing.T, v any, tag string) []string {
	t.Helper()
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	names := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Tag.Get(tag)
		if idx := len(name); idx > 0 {
			if comma := indexByte(name, ','); comma >= 0 {
				name = name[:comma]
			}
		}
		if name == "" || name == "-" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func indexByte(s string, b byte) int {
	for i := range len(s) {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// Every field the spec exports must have a writable counterpart, or export is
// producing something import can never write back.
func TestSpecResourceFieldsAreAllWritable(t *testing.T) {
	spec := fieldNames(t, specmodel.Resources{}, "yaml")
	writable := fieldNames(t, appsettingsdto.UpdateAppResourceSettingsReq{}, "json")

	for _, name := range spec {
		assert.Contains(t, writable, name,
			"spec exports %q but no Update request accepts it", name)
	}
}

// A writable field the spec omits is a coverage gap. It is reported rather than
// asserted away: this is the signal that replaces "a field added to a DTO
// propagates automatically". Update the spec type, then update this list.
func TestSpecResourceFieldCoverageGaps(t *testing.T) {
	spec := fieldNames(t, specmodel.Resources{}, "yaml")
	writable := fieldNames(t, appsettingsdto.UpdateAppResourceSettingsReq{}, "json")

	// Fields that are deliberately not in the spec.
	expectedGaps := []string{"updateVer"}

	var gaps []string
	for _, name := range writable {
		if !contains(spec, name) && !contains(expectedGaps, name) {
			gaps = append(gaps, name)
		}
	}
	assert.Empty(t, gaps,
		"these configurable fields are missing from the spec: %v", gaps)
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
```

Repeat both tests for `Storage`/`UpdateAppStorageSettingsReq`,
`Networks`/`UpdateAppNetworkSettingsReq`, `Service`/`UpdateAppServiceSettingsReq`
and `Container`/`UpdateAppContainerSettingsReq`. For `Storage` the expected gaps
include `key`; for `Container` they include `command` and `workingDir`, which
live in `Source`.

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestSpec -v`
Expected: PASS. A failure names exactly which field is missing.

- [ ] **Step 7: Commit**

```bash
make fmt && make lint
go test ./hivepaas_app/service/specservice/...
git add hivepaas_app/service/specservice
git commit -m "feat(spec): map swarm services onto spec types with a coverage oracle"
```

---

## Task 7: Document assembly - singletons, collections and reference paths

**Files:**
- Create: `hivepaas_app/service/specservice/specmodel/singleton.go`
- Create: `hivepaas_app/service/specservice/specmodel/document.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/assemble.go`
- Test: `hivepaas_app/service/specservice/specmodel/singleton_test.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/assemble_test.go`

**Interfaces:**
- Consumes: `specmodel.DeriveSettingKeys` (Task 2), `entity.Setting`.
- Produces: `specmodel.IsSingletonType(base.SettingType) bool`, `specmodel.SettingBlock`, `specmodel.ExternalRef`, and `assembleSettings(settings []*entity.Setting, scopePath string, index *refIndex) (map[string]any, error)` plus `newRefIndex()`.

**Two shapes, not one.** A setting type that can only exist once per scope needs
no key at all: the block name is the identity. `app-routing` and `env-var` are
the two types in a real database with no name whatsoever, and both are
singletons, which is what makes that safe. Collection types become a map keyed
by the key Task 2 derived.

**References become paths.** In the database a reference is a ULID. In a spec it
is a scope path, so the scope travels with the reference and two settings with
the same name in different scopes can never be confused. A reference whose
target is outside the exported scope cannot be turned into a path, and becomes
an `external` block carrying enough to find it again on the target.

- [ ] **Step 1: Write the failing test for singleton classification**

```go
// hivepaas_app/service/specservice/specmodel/singleton_test.go
package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// These two are the only types in a real installation whose settings carry no
// name at all. If either were treated as a collection, its key would be the
// empty string.
func TestTypesWithNoNameAreSingletons(t *testing.T) {
	assert.True(t, IsSingletonType(base.SettingTypeEnvVar))
	assert.True(t, IsSingletonType(base.SettingTypeAppRouting))
}

func TestNamedCollectionTypesAreNotSingletons(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeSSLCert,
		base.SettingTypeSecret,
		base.SettingTypeSSHKey,
		base.SettingTypeConfigFile,
		base.SettingTypeSchedJob,
		base.SettingTypeOAuth,
	} {
		assert.False(t, IsSingletonType(typ), "%v holds many per scope", typ)
	}
}

func TestSingletonBlockNamesAreUnique(t *testing.T) {
	seen := map[string]base.SettingType{}
	for typ := range singletonBlockNames {
		name := SingletonBlockName(typ)
		assert.NotEmpty(t, name, "%v has no block name", typ)
		prev, dup := seen[name]
		assert.False(t, dup, "block name %q used by both %v and %v", name, prev, typ)
		seen[name] = typ
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run "TestTypes|TestNamed|TestSingleton" -v`
Expected: FAIL — `undefined: IsSingletonType`.

- [ ] **Step 3: Write the classification**

Derive the list first, rather than trusting this one. A singleton is a type read
through `SettingRepo.GetSingle(scope, type)` or the `setting_*_unique.go` family:

```bash
grep -rn "GetUniqueSetting\|GetSingle(" --include="*.go"   hivepaas_app/usecase/ hivepaas_app/service/ | grep -v _test
```

Cross-check against the database, where a singleton is a type with at most one
row per `(scope, object_id)`:

```sql
SELECT type, max(cnt) FROM (
  SELECT type, scope, coalesce(object_id,'') o, count(*) cnt
  FROM settings WHERE deleted_at IS NULL GROUP BY 1,2,3
) t GROUP BY type ORDER BY 2 DESC;
```

Then write:

```go
// hivepaas_app/service/specservice/specmodel/singleton.go
package specmodel

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

// singletonBlockNames maps every setting type that can hold at most one setting
// per scope to the name its block takes in a spec document.
//
// These need no key: the block name is the identity. That is not a convenience
// - `env-var` and `app-routing` settings carry no name at all in a real
// installation, so a key derived from the name would be the empty string.
//
// A type is a singleton here if it is read through SettingRepo.GetSingle or the
// setting_*_unique.go usecases. Everything else is a collection.
var singletonBlockNames = map[base.SettingType]string{
	base.SettingTypeApp:               "app",
	base.SettingTypeAppClone:          "clone",
	base.SettingTypeAppDeployment:     "deployment",
	base.SettingTypeAppFeatures:       "features",
	base.SettingTypeAppKind:           "kind",
	base.SettingTypeAppPlacement:      "placement",
	base.SettingTypeAppRouting:        "routing",
	base.SettingTypeBackupRepoCleanup: "backupRepoCleanup",
	base.SettingTypeDomainSettings:    "domainSettings",
	base.SettingTypeEnvVar:            "envVars",
	base.SettingTypeHivePaaSService:   "hivepaasService",
	base.SettingTypeImageBuild:        "imageBuild",
	base.SettingTypeLogging:           "logging",
	base.SettingTypeProject:           "project",
	base.SettingTypeSSLRenewal:        "sslRenewal",
	base.SettingTypeSystemBackup:      "systemBackup",
	base.SettingTypeSystemCleanup:     "systemCleanup",
	base.SettingTypeTraefikConfig:     "traefikConfig",
	base.SettingTypeTraefikService:    "traefikService",
}

// collectionBlockNames names the block each collection type's keyed map takes.
var collectionBlockNames = map[base.SettingType]string{
	base.SettingTypeAccessToken:     "accessTokens",
	base.SettingTypeAcmeDnsProvider: "acmeDnsProviders",
	base.SettingTypeBackupRepo:      "backupRepos",
	base.SettingTypeBasicAuth:       "basicAuths",
	base.SettingTypeCloudStorage:    "cloudStorages",
	base.SettingTypeClusterNetwork:  "networks",
	base.SettingTypeClusterNode:     "nodes",
	base.SettingTypeClusterVolume:   "volumes",
	base.SettingTypeCommandPipe:     "commandPipes",
	base.SettingTypeCommandTemplate: "commandTemplates",
	base.SettingTypeConfigFile:      "configFiles",
	base.SettingTypeEmail:           "emails",
	base.SettingTypeGithubApp:       "githubApps",
	base.SettingTypeIMService:       "imServices",
	base.SettingTypeNotification:    "notifications",
	base.SettingTypeOAuth:           "oauths",
	base.SettingTypePeriodicJob:     "periodicJobs",
	base.SettingTypeRegistryAuth:    "registryAuths",
	base.SettingTypeRepoWebhook:     "repoWebhooks",
	base.SettingTypeSSHKey:          "sshKeys",
	base.SettingTypeSSLCert:         "sslCerts",
	base.SettingTypeSSLProvider:     "sslProviders",
	base.SettingTypeSchedJob:        "schedJobs",
	base.SettingTypeScript:          "scripts",
	base.SettingTypeSecret:          "secrets",
}

func IsSingletonType(typ base.SettingType) bool {
	_, ok := singletonBlockNames[typ]
	return ok
}

// SingletonBlockName is the YAML key a singleton type occupies.
func SingletonBlockName(typ base.SettingType) string { return singletonBlockNames[typ] }

// CollectionBlockName is the YAML key a collection type's map occupies.
func CollectionBlockName(typ base.SettingType) string { return collectionBlockNames[typ] }
```

Adjust the two maps to whatever the grep and the query actually show. If a type
appears in neither map, `assembleSettings` must refuse it rather than guess —
add that to the test in Step 5.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run "TestTypes|TestNamed|TestSingleton" -v`
Expected: PASS, three tests.

- [ ] **Step 5: Write the failing test for assembly and reference paths**

```go
// hivepaas_app/service/specservice/specserviceimpl/assemble_test.go
package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestAssembleSettingsPutsSingletonsUnderTheirBlockName(t *testing.T) {
	routing := &entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting}
	assert.NoError(t, routing.SetData(&entity.AppRoutingSettings{Port: 8080}))

	out, err := assembleSettings([]*entity.Setting{routing}, "apps/backend", newRefIndex())
	assert.NoError(t, err)

	block, ok := out["routing"]
	assert.True(t, ok, "a singleton takes its block name, with no key")
	assert.NotNil(t, block)
	_, isMap := out["routings"]
	assert.False(t, isMap)
}

func TestAssembleSettingsPutsCollectionsInAKeyedMap(t *testing.T) {
	cert := &entity.Setting{ID: "s1", Type: base.SettingTypeSSLCert, Name: "localhost", Kind: "self-signed"}
	assert.NoError(t, cert.SetData(&entity.SSLCert{Domain: "localhost"}))

	out, err := assembleSettings([]*entity.Setting{cert}, "global", newRefIndex())
	assert.NoError(t, err)

	certs, ok := out["sslCerts"].(map[string]any)
	assert.True(t, ok)
	_, found := certs["localhost"]
	assert.True(t, found, "keyed by the derived key, not by id")
}

// A reference whose target is in the export becomes a path carrying its scope.
func TestAssembleSettingsRewritesInScopeReferencesToPaths(t *testing.T) {
	index := newRefIndex()
	index.add("cert_1", "global/sslCerts/localhost@self-signed")

	routing := &entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting}
	assert.NoError(t, routing.SetData(&entity.AppRoutingSettings{
		Port:    443,
		Domains: []*entity.AppDomain{{Domain: "x.com", SSLCert: entity.ObjectID{ID: "cert_1"}}},
	}))

	out, err := assembleSettings([]*entity.Setting{routing}, "apps/backend", index)
	assert.NoError(t, err)
	assert.Contains(t, mustYAML(t, out), "global/sslCerts/localhost@self-signed")
	assert.NotContains(t, mustYAML(t, out), "cert_1")
}

// A reference outside the exported scope cannot be a path, and becomes an
// external block with enough to find the target on the importing installation.
func TestAssembleSettingsWritesExternalRefsForOutOfScopeTargets(t *testing.T) {
	index := newRefIndex()
	index.addExternal("cert_9", &entity.Setting{
		ID: "cert_9", Type: base.SettingTypeSSLCert, Name: "wildcard", Kind: "letsencrypt",
	})

	routing := &entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting}
	assert.NoError(t, routing.SetData(&entity.AppRoutingSettings{
		Domains: []*entity.AppDomain{{Domain: "x.com", SSLCert: entity.ObjectID{ID: "cert_9"}}},
	}))

	out, err := assembleSettings([]*entity.Setting{routing}, "apps/backend", index)
	assert.NoError(t, err)

	yamlOut := mustYAML(t, out)
	assert.Contains(t, yamlOut, "external:")
	assert.Contains(t, yamlOut, "wildcard")
	assert.Contains(t, yamlOut, "letsencrypt")
}

// A type in neither block-name map is a programming error, not a silent skip.
func TestAssembleSettingsRefusesAnUnnamedType(t *testing.T) {
	_, err := assembleSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingType("brand-new-type")},
	}, "global", newRefIndex())
	assert.Error(t, err)
}
```

Write `mustYAML(t, v)` in the same file as a one-line `yaml.Marshal` helper that
fails the test on error.

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestAssemble -v`
Expected: FAIL — `undefined: assembleSettings`.

- [ ] **Step 7: Write the assembly**

```go
// hivepaas_app/service/specservice/specserviceimpl/assemble.go
package specserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// refIndex resolves a stored identifier to what a spec writes in its place.
//
// Everything reachable inside the exported scope becomes a path, so the scope
// travels with the reference. Everything else becomes an external block: the
// exporter cannot resolve it, so it records what the importing installation
// needs to look it up - which is the situation the scope rule creates every
// time somebody exports a project whose apps use a global certificate.
type refIndex struct {
	paths    map[string]string
	external map[string]*specmodel.ExternalRef
}

func newRefIndex() *refIndex {
	return &refIndex{
		paths:    map[string]string{},
		external: map[string]*specmodel.ExternalRef{},
	}
}

func (r *refIndex) add(settingID, path string) { r.paths[settingID] = path }

func (r *refIndex) addExternal(settingID string, setting *entity.Setting) {
	r.external[settingID] = &specmodel.ExternalRef{
		Type: string(setting.Type),
		Name: setting.Name,
		Kind: setting.Kind,
		ID:   setting.ID,
	}
}

// assembleSettings turns a scope's settings into the map one document holds.
//
// A singleton takes its block name and nothing else; a collection becomes a map
// keyed by the key DeriveSettingKeys produced. A type in neither map is refused
// rather than guessed at, for the same reason an unregistered SpecPolicy is: a
// setting type added later should stop somebody, not appear under a name nobody
// chose.
func assembleSettings(
	settings []*entity.Setting,
	scopePath string,
	index *refIndex,
) (map[string]any, error) {
	byType := map[string][]*entity.Setting{}
	for _, setting := range settings {
		byType[string(setting.Type)] = append(byType[string(setting.Type)], setting)
	}

	out := map[string]any{}
	for _, group := range byType {
		typ := group[0].Type

		switch {
		case specmodel.IsSingletonType(typ):
			body, err := renderSetting(group[0], index)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			out[specmodel.SingletonBlockName(typ)] = body

		case specmodel.CollectionBlockName(typ) != "":
			keys := specmodel.DeriveSettingKeys(group)
			entries := map[string]any{}
			for _, setting := range group {
				body, err := renderSetting(setting, index)
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				entries[keys[setting.ID]] = body
			}
			out[specmodel.CollectionBlockName(typ)] = entries

		default:
			return nil, hperrors.Wrap(hperrors.ErrSpecSettingTypeUnclassified).
				WithParam("Type", string(typ))
		}
	}
	return out, nil
}
```

Write `renderSetting(setting, index)`: parse the setting, call
`entity.RemapRefs` (Task 3) with a mapping from every referenced identifier to
its path, then marshal the parsed data to a `map[string]any` and splice in an
`external` block for each identifier the index had no path for. Add
`specmodel/document.go`, which this task creates in full:

```go
// hivepaas_app/service/specservice/specmodel/document.go
package specmodel

// DocHeader is repeated at the top of every payload file, so that a file
// extracted from a bundle and handed to somebody on its own still says what it
// is.
type DocHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       Kind   `yaml:"kind"`
	Scope      string `yaml:"scope"`
}

// GlobalDoc is global.yaml: the global and hivepaas scope settings.
//
// It does not carry the hivepaas *project* - that is HivePaaS's own stack, and
// selectProjects drops it.
type GlobalDoc struct {
	DocHeader `yaml:",inline"`

	// Settings is the assembled block map: singleton types under their block
	// name, collection types as keyed maps. Built by assembleSettings.
	Settings map[string]any `yaml:",inline"`
}

// ProjectDoc is projects/<key>/project.yaml.
type ProjectDoc struct {
	DocHeader `yaml:",inline"`

	Project string `yaml:"project"`
	Name    string `yaml:"name"`
	Note    string `yaml:"note,omitempty"`
	// Envs names the env files that belong to this project, so a reader of one
	// project file knows what else there is without listing the archive.
	Envs []string `yaml:"envs,omitempty"`

	Settings map[string]any `yaml:",inline"`
}

// EnvDoc is projects/<key>/envs/<env>.yaml: the env's own settings and every
// app in it.
type EnvDoc struct {
	DocHeader `yaml:",inline"`

	Project string `yaml:"project"`
	Env     string `yaml:"env"`
	Name    string `yaml:"name"`
	Color   string `yaml:"color,omitempty"`
	Index   int    `yaml:"index,omitempty"`

	// Labels are the env-scope service labels, already filtered.
	Labels map[string]string `yaml:"labels,omitempty"`

	// Apps is keyed by app key, which is unique within an env.
	Apps map[string]*AppDoc `yaml:"apps,omitempty"`

	Settings map[string]any `yaml:",inline"`
}

// AppDoc is one app inside an env document.
//
// Deployment is a pointer and is omitted entirely for an app that has never
// been deployed - a real state, not an edge case: two of five user apps in a
// development installation have an empty ServiceID.
type AppDoc struct {
	App    string `yaml:"app"`
	Name   string `yaml:"name"`
	Status string `yaml:"status,omitempty"`
	Note   string `yaml:"note,omitempty"`

	Deployment *Deployment `yaml:"deployment,omitempty"`

	Settings map[string]any `yaml:",inline"`
}

// ExternalRef stands in for a reference whose target was outside the export.
type ExternalRef struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Kind string `yaml:"kind,omitempty"`
	// ID is the source identifier, which lets a re-import into the same
	// installation match exactly instead of by name.
	ID string `yaml:"id,omitempty"`
}

// Bundle is what the exporter hands the bundle writer.
type Bundle struct {
	Manifest *Manifest
	// Files maps bundle-relative path to already-serialized content.
	Files map[string][]byte
}
```

`yaml:",inline"` on a `map[string]any` splices the assembled blocks into the
document rather than nesting them under a `settings:` key. Verify yaml.v3
accepts inline on both an embedded struct and a map in the same type - if it
does not, drop the inline tag on `Settings` and let the blocks sit under
`settings:`, then update Task 10's tests to match.

- [ ] **Step 8: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/... -v`
Expected: PASS, all assembly and singleton tests plus the earlier tasks'.

- [ ] **Step 9: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice
git commit -m "feat(spec): assemble documents with singleton blocks and reference paths"
```

---

## Task 8: Scope walking

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/walk.go`
- Create: `hivepaas_app/service/specservice/service.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/service.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/walk_test.go`

**Interfaces:**
- Consumes: `entity.SpecExportDecision` (Task 4), `assembleSettings` and `newRefIndex` (Task 7), `mapSwarmService` (Task 6), `repository.SettingRepo`, `repository.ProjectRepo`, `repository.ProjectEnvRepo`, `repository.AppRepo`, `clusterservice.Service`, `networkservice.Service`.
- Produces: `specservice.Service` with `Export(ctx, db, *ExportReq) (*ExportResp, error)`; `specmodel.GlobalDoc`, `specmodel.ProjectDoc`, `specmodel.EnvDoc`, `specmodel.AppDoc`, `specmodel.Bundle`.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/walk_test.go
package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func TestSelectProjectsExcludesTheHivePaaSProject(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectProjects([]*entity.Project{
		{ID: "p1", Key: "project_a"},
		{ID: "p2", Key: base.HivepaasProjectKey},
		{ID: "p3", Key: "project_b"},
	}, report)

	assert.Len(t, kept, 2)
	assert.Equal(t, "project_a", kept[0].Key)
	assert.Equal(t, "project_b", kept[1].Key)
}

// Child apps exist only as preview environments, created by apppreviewservice
// when a pull request opens and removed when it closes.
func TestSelectAppsExcludesPreviewApps(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectApps([]*entity.App{
		{ID: "a1", Key: "backend"},
		{ID: "a2", Key: "backend-pr-42", ParentID: "a1"},
		{ID: "a3", Key: "frontend"},
	}, report)

	assert.Len(t, kept, 2)
	assert.Equal(t, "backend", kept[0].Key)
	assert.Equal(t, "frontend", kept[1].Key)
	assert.Equal(t, 1, report.CountBySeverity(specmodel.SeveritySkipped))
	assert.Equal(t, specmodel.CodePreviewAppSkipped, report.Issues[0].Code)
}

func TestSelectSettingsAppliesThePolicyAndReports(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingTypeSSLCert, Name: "a"},
		{ID: "s2", Type: base.SettingTypeAPIKey, Name: "b"},
		{ID: "s3", Type: base.SettingTypeClusterNetwork, Scope: base.ObjectScopeGlobal, Name: "bridge"},
		{ID: "s4", Type: base.SettingTypeClusterNetwork, Scope: base.ObjectScopeProject, Name: "default"},
	}, "global", report)

	assert.Len(t, kept, 2)
	assert.Equal(t, "s1", kept[0].ID)
	assert.Equal(t, "s4", kept[1].ID)
	assert.Equal(t, 2, report.CountBySeverity(specmodel.SeveritySkipped))
}

// A setting type with no registered policy must be refused, not exported.
func TestSelectSettingsRefusesAnUnclassifiedType(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingType("brand-new-type"), Name: "x"},
	}, "global", report)

	assert.Empty(t, kept)
	assert.Equal(t, specmodel.CodeTypeUnclassified, report.Issues[0].Code)
}

// Ordering must not depend on what the database returned.
func TestSelectAppsIsOrderedByKey(t *testing.T) {
	report := &specmodel.Report{}
	kept := selectApps([]*entity.App{
		{ID: "a3", Key: "zeta"},
		{ID: "a1", Key: "alpha"},
		{ID: "a2", Key: "mu"},
	}, report)
	assert.Equal(t, []string{"alpha", "mu", "zeta"},
		[]string{kept[0].Key, kept[1].Key, kept[2].Key})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run "TestSelect" -v`
Expected: FAIL — `undefined: selectProjects`.

- [ ] **Step 3: Write the selection functions**

```go
// hivepaas_app/service/specservice/specserviceimpl/walk.go
package specserviceimpl

import (
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// selectProjects drops the hivepaas project.
//
// It holds HivePaaS's own stack - app, worker, db, redis, traefik, adminer,
// agent, updater, redis_ui - modelled as an ordinary project. Importing it
// elsewhere would redeploy HivePaaS's own database and proxy over the ones
// install.sh had just created. base.UnallowedProjectKeys already forbids anyone
// from creating another project with this key.
func selectProjects(projects []*entity.Project, report *specmodel.Report) []*entity.Project {
	kept := make([]*entity.Project, 0, len(projects))
	for _, project := range projects {
		if project.Key == base.HivepaasProjectKey {
			continue // HivePaaS's own stack; not silently, but not a user issue either
		}
		kept = append(kept, project)
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Key < kept[j].Key })
	return kept
}

// selectApps drops preview apps.
//
// The only thing that sets ParentID is apppreviewservice - "Preview app must be
// a child app of the current" - so a child app is always a preview environment
// for a pull request, created when it opens and destroyed when it closes.
// Restoring one for a request that closed months ago is noise, and the preview
// machinery would remove it anyway.
func selectApps(apps []*entity.App, report *specmodel.Report) []*entity.App {
	kept := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		if app.ParentID != "" {
			report.Add(specmodel.Issue{
				Severity: specmodel.SeveritySkipped,
				Code:     specmodel.CodePreviewAppSkipped,
				Path:     "apps/" + app.Key,
				Detail:   map[string]any{"parentId": app.ParentID},
				Action:   "not exported; preview apps are recreated by their own lifecycle",
			})
			continue
		}
		kept = append(kept, app)
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Key < kept[j].Key })
	return kept
}

// selectSettings applies each type's policy and records every refusal.
func selectSettings(
	settings []*entity.Setting,
	pathPrefix string,
	report *specmodel.Report,
) []*entity.Setting {
	kept := make([]*entity.Setting, 0, len(settings))
	for _, setting := range settings {
		decision := entity.SpecExportDecision(setting)
		if decision.Export {
			kept = append(kept, setting)
			continue
		}

		code := specmodel.CodeTypeSkipped
		if entity.SpecPolicyFor(setting.Type) == nil {
			code = specmodel.CodeTypeUnclassified
		}
		report.Add(specmodel.Issue{
			Severity: specmodel.SeveritySkipped,
			Code:     code,
			Path:     pathPrefix + "/" + string(setting.Type) + "/" + setting.Name,
			Detail:   map[string]any{"type": string(setting.Type), "id": setting.ID},
			Action:   decision.Reason,
		})
	}
	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].Type != kept[j].Type {
			return kept[i].Type < kept[j].Type
		}
		if !kept[i].CreatedAt.Equal(kept[j].CreatedAt) {
			return kept[i].CreatedAt.Before(kept[j].CreatedAt)
		}
		return kept[i].ID < kept[j].ID
	})
	return kept
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run "TestSelect" -v`
Expected: PASS, five tests.

- [ ] **Step 5: Write the service interface, constructor and Export skeleton**

```go
// hivepaas_app/service/specservice/service.go
package specservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type Service interface {
	// Export builds a configuration bundle for a scope. It reads only.
	Export(ctx context.Context, db database.IDB, req *ExportReq) (*ExportResp, error)
}

type ExportReq struct {
	Scope       *entity.ObjectScope
	SecretsMode specmodel.SecretsMode
	// Passphrase is required when SecretsMode is encrypted.
	Passphrase string
}

type ExportResp struct {
	// Filename is what the download is called, without a directory.
	Filename string
	// Content is the bundle. The caller closes it.
	Content io.ReadCloser
	Size    int64
	Report  *specmodel.Report
}
```

```go
// hivepaas_app/service/specservice/specserviceimpl/service.go
package specserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
)

func New(
	settingRepo repository.SettingRepo,
	projectRepo repository.ProjectRepo,
	projectEnvRepo repository.ProjectEnvRepo,
	appRepo repository.AppRepo,

	scopeService scopeservice.Service,
	clusterService clusterservice.Service,
	networkService networkservice.Service,
) specservice.Service {
	return &service{
		settingRepo:    settingRepo,
		projectRepo:    projectRepo,
		projectEnvRepo: projectEnvRepo,
		appRepo:        appRepo,

		scopeService:   scopeService,
		clusterService: clusterService,
		networkService: networkService,
	}
}

type service struct {
	settingRepo    repository.SettingRepo
	projectRepo    repository.ProjectRepo
	projectEnvRepo repository.ProjectEnvRepo
	appRepo        repository.AppRepo

	scopeService   scopeservice.Service
	clusterService clusterservice.Service
	networkService networkservice.Service
}
```

Check the repository interface names first:
`grep -n "ProjectEnvRepo\|ProjectRepo\|AppRepo" hivepaas_app/repository/*.go | head`

Then write `Export` in `walk.go`: load the scope's settings and children through
the repositories, run them through `selectSettings`/`selectProjects`/`selectApps`,
call `DeriveSettingKeys` per (scope, type) group, and for each app with a
non-empty `ServiceID` call `clusterService.ServiceInspect` and `mapSwarmService`.

An app whose `ServiceInspect` fails records a `SeverityFixable` issue with code
`specmodel.CodeServiceUnavailable` and exports its settings alone. Two of the
five user apps in a real development database have an empty `ServiceID`, so this
path is normal, not exceptional.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
go test ./hivepaas_app/service/specservice/...
git add hivepaas_app/service/specservice
git commit -m "feat(spec): walk scopes, excluding the hivepaas project and preview apps"
```

---

## Task 9: Secret modes

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/secrets.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/secrets_test.go`

**Interfaces:**
- Consumes: `entity.Setting`, `specmodel.SecretsMode`.
- Produces: `revealSettingSecrets(setting *entity.Setting, mode specmodel.SecretsMode) error`.

**Reuse, do not reimplement.** `usecase/settings/setting_secrets.go` already
defines `secretDecrypter` and the reveal path, and its comment says why: reaching
decryption through one interface "is what lets one place cover all of them - and
what makes a setting type added later covered without anybody remembering to come
back here". `specservice` cannot import the usecase, so it declares the same
one-method interface locally — an interface, not logic, and Go interfaces are
satisfied structurally. The capability check and audit record stay in the
usecase, in Task 11.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/secrets_test.go
package specserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

func storedSecret(t *testing.T, plain string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "db-password"}
	assert.NoError(t, setting.SetData(&entity.Secret{
		Key: "DB_PASSWORD", Value: entity.NewEncryptedField(plain),
	}))
	return &entity.Setting{ID: setting.ID, Type: setting.Type, Name: setting.Name, Data: setting.Data}
}

func TestRevealSettingSecretsLeavesCiphertextInNoneMode(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModeNone))

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	assert.True(t, data.Value.IsEncrypted(), "none mode must not decrypt anything")
	assert.NotContains(t, setting.Data, "hunter2")
}

func TestRevealSettingSecretsDecryptsInPlaintextMode(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModePlaintext))

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	plain, err := data.Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain)
}

func TestRevealSettingSecretsDecryptsInEncryptedMode(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModeEncrypted))

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	plain, err := data.Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", plain,
		"the bundle is sealed by age, so the values inside it are plain")
}

// A type holding no secrets must not error just because a mode asks for them.
func TestRevealSettingSecretsIgnoresTypesWithoutSecrets(t *testing.T) {
	useDataKey(t)
	setting := &entity.Setting{ID: "s2", Type: base.SettingTypeScript}
	assert.NoError(t, setting.SetData(&entity.Script{Data: "echo hi"}))

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModePlaintext))
}

func TestSecretsModeNoneOmitsSecretsFromSerializedOutput(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")
	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModeNone))
	assert.False(t, strings.Contains(setting.Data, "hunter2"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestReveal -v`
Expected: FAIL — `undefined: revealSettingSecrets`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specserviceimpl/secrets.go
package specserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// secretDecrypter is implemented by every setting type that stores secrets.
//
// It is declared here rather than imported because usecase/settings sits above
// this package in the layering. It is an interface, not logic: Go satisfies it
// structurally, so every type that already implements Decrypt for the reveal
// path is covered here too, including ones added later.
type secretDecrypter interface {
	Decrypt() error
}

// revealSettingSecrets decrypts a setting's secrets in place when the mode calls
// for it.
//
// In none mode nothing is touched, so the values stay as the ciphertext they
// were loaded with and serialize back out sealed by a key the bundle does not
// contain. In encrypted mode they are decrypted, because the bundle as a whole
// is then wrapped by age under the operator's passphrase.
//
// The capability check and the audit record are the usecase's, not this
// function's - see specuc.ExportSpec.
func revealSettingSecrets(setting *entity.Setting, mode specmodel.SecretsMode) error {
	if !mode.RevealsSecrets() || setting == nil {
		return nil
	}

	data, err := setting.Parse()
	if err != nil {
		return hperrors.Wrap(err)
	}
	decrypter, ok := data.(secretDecrypter)
	if !ok {
		return nil // the type holds no secrets, so there is nothing to reveal
	}
	if err = decrypter.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run "TestReveal|TestSecretsMode" -v`
Expected: PASS, five tests.

- [ ] **Step 5: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specserviceimpl/secrets.go hivepaas_app/service/specservice/specserviceimpl/secrets_test.go
git commit -m "feat(spec): reveal secrets per mode, reusing the Decrypt seam"
```

---

## Task 10: Deterministic YAML serialization

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/serialize.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/serialize_test.go`

**Interfaces:**
- Consumes: `specmodel.GlobalDoc`, `specmodel.ProjectDoc`, `specmodel.EnvDoc`.
- Produces: `marshalDoc(doc any) ([]byte, error)`.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/serialize_test.go
package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// Two runs over unchanged data must produce identical bytes, or the format is
// useless for review and for git.
func TestMarshalDocIsByteIdenticalAcrossRuns(t *testing.T) {
	doc := &specmodel.EnvDoc{
		APIVersion: specmodel.APIVersion,
		Kind:       specmodel.KindSpec,
		Scope:      "project-env",
		Project:    "project_a",
		Env:        "dev",
	}

	first, err := marshalDoc(doc)
	assert.NoError(t, err)
	second, err := marshalDoc(doc)
	assert.NoError(t, err)
	assert.Equal(t, string(first), string(second))
}

// Go map iteration is randomized, so a map written straight out would differ
// between runs. Every map-valued field must come out sorted.
func TestMarshalDocSortsMapKeys(t *testing.T) {
	doc := &specmodel.EnvDoc{
		APIVersion: specmodel.APIVersion,
		Kind:       specmodel.KindSpec,
		Labels: map[string]string{
			"zebra": "1", "alpha": "2", "mu": "3", "beta": "4", "omega": "5",
		},
	}

	// Twenty runs: with randomized iteration, an unsorted encoder fails here
	// with overwhelming probability.
	want, err := marshalDoc(doc)
	assert.NoError(t, err)
	for range 20 {
		got, err := marshalDoc(doc)
		assert.NoError(t, err)
		assert.Equal(t, string(want), string(got))
	}
}

func TestMarshalDocUsesTwoSpaceIndent(t *testing.T) {
	doc := &specmodel.EnvDoc{
		APIVersion: specmodel.APIVersion,
		Kind:       specmodel.KindSpec,
		Labels:     map[string]string{"a": "1"},
	}
	out, err := marshalDoc(doc)
	assert.NoError(t, err)
	assert.Contains(t, string(out), "\n  a: \"1\"")
}
```

Add a `Labels map[string]string` field to `specmodel.EnvDoc` for this test; it is
where env-scope service labels belong anyway.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestMarshalDoc -v`
Expected: FAIL — `undefined: marshalDoc`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specserviceimpl/serialize.go
package specserviceimpl

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// yamlIndent is two spaces. yaml.v3 defaults to four, which reads badly at the
// nesting depth an app reaches.
const yamlIndent = 2

// marshalDoc writes one payload file.
//
// yaml.v3 sorts map keys already, which is what makes two runs over unchanged
// data byte-identical - Go's randomized map iteration would otherwise reorder
// every label block between exports and make every diff noise. Struct fields
// come out in declaration order, so ordering within a document is decided by
// how the types in specmodel are written.
func marshalDoc(doc any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndent)

	if err := encoder.Encode(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := encoder.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestMarshalDoc -v`
Expected: PASS, three tests.

If `TestMarshalDocSortsMapKeys` fails, yaml.v3 is not sorting and the encoder
needs an explicit sorted intermediate — build a `yaml.Node` with keys sorted by
`sort.Strings` before encoding. Verify which is true before changing anything.

- [ ] **Step 5: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specserviceimpl/serialize.go hivepaas_app/service/specservice/specserviceimpl/serialize_test.go
git commit -m "feat(spec): serialize spec documents deterministically"
```

---

## Task 11: Bundle writing and age encryption

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/bundle.go`
- Test: `hivepaas_app/service/specservice/specserviceimpl/bundle_test.go`

**Interfaces:**
- Consumes: `marshalDoc` (Task 10), `specmodel.Manifest`, `filearchiver.CompressTarGz`, `filippo.io/age`.
- Produces: `writeBundle(dir string, files map[string][]byte, manifest *specmodel.Manifest) (string, error)`, `encryptBundle(src, dst, passphrase string) error`.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/service/specservice/specserviceimpl/bundle_test.go
package specserviceimpl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func TestWriteBundleLaysOutFilesPerEnv(t *testing.T) {
	dir := t.TempDir()
	manifest := &specmodel.Manifest{
		APIVersion: specmodel.APIVersion,
		Kind:       specmodel.KindSpec,
		ExportedAt: time.Date(2026, 9, 16, 10, 15, 0, 0, time.UTC),
		Scope:      "global",
		SecretsMode: specmodel.SecretsModeNone,
		Files: []string{
			"global.yaml",
			"projects/project_a/project.yaml",
			"projects/project_a/envs/dev.yaml",
		},
	}
	files := map[string][]byte{
		"global.yaml":                     []byte("kind: Spec\n"),
		"projects/project_a/project.yaml": []byte("kind: Spec\n"),
		"projects/project_a/envs/dev.yaml": []byte("kind: Spec\n"),
	}

	path, err := writeBundle(dir, files, manifest)
	assert.NoError(t, err)
	assert.FileExists(t, path)
	assert.Equal(t, ".gz", filepath.Ext(path))

	// spec.yaml is written even though it is not in files
	staged := filepath.Join(dir, "stage", "spec.yaml")
	content, err := os.ReadFile(staged)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "apiVersion: hivepaas.com/v1")
}

func TestEncryptBundleProducesSomethingAgeCanOpen(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tar.gz")
	assert.NoError(t, os.WriteFile(src, []byte("not really a tarball, but bytes"), 0o600))

	dst := filepath.Join(dir, "bundle.tar.gz.age")
	assert.NoError(t, encryptBundle(src, dst, "correct horse battery staple"))

	sealed, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.NotContains(t, string(sealed), "not really a tarball")

	plain, err := decryptBundleForTest(t, dst, "correct horse battery staple")
	assert.NoError(t, err)
	assert.Equal(t, "not really a tarball, but bytes", string(plain))
}

func TestEncryptBundleRefusesAnEmptyPassphrase(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tar.gz")
	assert.NoError(t, os.WriteFile(src, []byte("x"), 0o600))

	err := encryptBundle(src, filepath.Join(dir, "out.age"), "")
	assert.Error(t, err)
}
```

Write `decryptBundleForTest` in the same file using `age.NewScryptIdentity` and
`age.Decrypt`, mirroring how `encryptBundle` seals.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run "TestWriteBundle|TestEncryptBundle" -v`
Expected: FAIL — `undefined: writeBundle`.

- [ ] **Step 3: Write the implementation**

```go
// hivepaas_app/service/specservice/specserviceimpl/bundle.go
package specserviceimpl

import (
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filearchiver"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	manifestFilename = "spec.yaml"
	stageDirName     = "stage"
	bundleFilename   = "bundle.tar.gz"

	stageDirPerm  = 0o700
	stageFilePerm = 0o600
)

// writeBundle stages the payload files and the manifest under dir and archives
// them, returning the archive path.
func writeBundle(
	dir string,
	files map[string][]byte,
	manifest *specmodel.Manifest,
) (string, error) {
	stage := filepath.Join(dir, stageDirName)

	manifestBytes, err := marshalDoc(manifest)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if err = writeStagedFile(stage, manifestFilename, manifestBytes); err != nil {
		return "", hperrors.Wrap(err)
	}
	for name, content := range files {
		if err = writeStagedFile(stage, name, content); err != nil {
			return "", hperrors.Wrap(err)
		}
	}

	archive := filepath.Join(dir, bundleFilename)
	// Verified signature:
	//   CompressTarGz(srcFilename, destFilename string, level CompressionLevel)
	//       (cmdErr string, err error)
	// The first return is the failed command's output, which belongs in the log
	// rather than in the error a caller sees.
	cmdErr, err := filearchiver.CompressTarGz(stage, archive, filearchiver.CompressionLevelDefault)
	if err != nil {
		return "", hperrors.Wrap(err).WithMsgLog("tar failed: %s", cmdErr)
	}
	return archive, nil
}

func writeStagedFile(stage, name string, content []byte) error {
	path := filepath.Join(stage, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), stageDirPerm); err != nil {
		return hperrors.Wrap(err)
	}
	if err := os.WriteFile(path, content, stageFilePerm); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// encryptBundle wraps the whole archive with age under a passphrase.
//
// This is the same path sysbackupservice takes for an encrypted backup, and it
// is deliberately not the app secret. The app secret protects every stored
// secret in the installation, so shipping it to make an import possible would
// send the most sensitive credential there is along with the file - and it
// would not work anyway, since stored ciphertext is sealed by a data key
// generated per installation that the target does not have.
//
// age carries its own salt in its header, so the bundle needs no key material
// of its own, and the result opens with the standard age CLI without HivePaaS
// running at all.
func encryptBundle(src, dst, passphrase string) error {
	if passphrase == "" {
		return hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
	}

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return hperrors.Wrap(err)
	}

	in, err := os.Open(src)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stageFilePerm)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer out.Close()

	writer, err := age.Encrypt(out, recipient)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, err = io.Copy(writer, in); err != nil {
		return hperrors.Wrap(err)
	}
	if err = writer.Close(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
```

Signatures already verified against the codebase:
`CompressTarGz(src, dest string, level CompressionLevel) (cmdErr string, err error)`
with `CompressionLevelDefault = ""`, and `filippo.io/age` is a direct import
already used by `sysbackupservice` and `schedjobexecservice`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run "TestWriteBundle|TestEncryptBundle" -v`
Expected: PASS, three tests.

- [ ] **Step 5: Run the whole service suite**

Run: `go test ./hivepaas_app/service/specservice/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
make fmt && make lint
git add hivepaas_app/service/specservice/specserviceimpl/bundle.go hivepaas_app/service/specservice/specserviceimpl/bundle_test.go
git commit -m "feat(spec): write and optionally age-encrypt the bundle"
```

---

## Task 12: Usecase, handler, routes and wiring

**Files:**
- Create: `hivepaas_app/usecase/specuc/uc.go`
- Create: `hivepaas_app/usecase/specuc/export.go`
- Create: `hivepaas_app/usecase/specuc/specdto/export.go`
- Create: `hivepaas_app/interface/api/handler/spechandler/handler.go`
- Create: `hivepaas_app/interface/api/handler/spechandler/export.go`
- Modify: `hivepaas_app/registry/provides.go`
- Modify: `hivepaas_app/interface/api/server/router_projects.go` and the global/system router
- Test: `hivepaas_app/usecase/specuc/export_test.go`

**Interfaces:**
- Consumes: `specservice.Service.Export`, `permission.Manager.AuthorizeSecretReveal`, `auditservice`.
- Produces: `specuc.UC.ExportSpec(ctx, auth, req) (*specdto.ExportSpecResp, error)`; routes `GET /spec/export`, `GET /projects/:projectID/spec/export`, `GET /projects/:projectID/:env/spec/export`, `GET /projects/:projectID/:env/apps/:appID/spec/export`.

**The capability gate lives here.** `encrypted` and `plaintext` both decrypt every
secret in scope, so both must pass `AuthorizeSecretReveal` and leave an audit
record. `none` decrypts nothing and needs neither. Follow
`appserviceimpl/secret_reveal.go` for the `RevealSubject` shape.

- [ ] **Step 1: Write the failing test**

```go
// hivepaas_app/usecase/specuc/export_test.go
package specuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

func TestExportSpecRequiresRevealForPlaintextMode(t *testing.T) {
	uc, perm := newTestUC(t)
	perm.denyReveal = true

	_, err := uc.ExportSpec(context.Background(), testAuth(), &specdto.ExportSpecReq{
		SecretsMode: specmodel.SecretsModePlaintext,
	})
	assert.Error(t, err)
	assert.Equal(t, 1, perm.revealCalls, "the capability gate must be consulted")
}

func TestExportSpecRequiresRevealForEncryptedMode(t *testing.T) {
	uc, perm := newTestUC(t)
	perm.denyReveal = true

	_, err := uc.ExportSpec(context.Background(), testAuth(), &specdto.ExportSpecReq{
		SecretsMode: specmodel.SecretsModeEncrypted,
		Passphrase:  "pw",
	})
	assert.Error(t, err)
	assert.Equal(t, 1, perm.revealCalls)
}

func TestExportSpecNeedsNoRevealForNoneMode(t *testing.T) {
	uc, perm := newTestUC(t)
	perm.denyReveal = true

	_, err := uc.ExportSpec(context.Background(), testAuth(), &specdto.ExportSpecReq{
		SecretsMode: specmodel.SecretsModeNone,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, perm.revealCalls, "none mode decrypts nothing")
}

func TestExportSpecRejectsEncryptedWithoutAPassphrase(t *testing.T) {
	uc, _ := newTestUC(t)
	_, err := uc.ExportSpec(context.Background(), testAuth(), &specdto.ExportSpecReq{
		SecretsMode: specmodel.SecretsModeEncrypted,
	})
	assert.Error(t, err)
}

func TestExportSpecRejectsAnUnknownMode(t *testing.T) {
	uc, _ := newTestUC(t)
	_, err := uc.ExportSpec(context.Background(), testAuth(), &specdto.ExportSpecReq{
		SecretsMode: specmodel.SecretsMode("clear"),
	})
	assert.Error(t, err)
}

func testAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "u1"}}}
}
```

Write `newTestUC` returning a `*UC` wired with a fake `specservice.Service` that
returns an empty bundle, and a fake permission manager recording `revealCalls`
and honouring `denyReveal`. Follow the fakes in
`hivepaas_app/permission/permissionimpl/secret_reveal_test.go`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/usecase/specuc/ -v`
Expected: FAIL — the package does not exist.

- [ ] **Step 3: Write the DTO and the usecase**

```go
// hivepaas_app/usecase/specuc/specdto/export.go
package specdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ExportSpecReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	SecretsMode specmodel.SecretsMode `json:"-" mapstructure:"secretsMode"`
	Passphrase  string                `json:"-" mapstructure:"passphrase"`
}

func NewExportSpecReq() *ExportSpecReq {
	return &ExportSpecReq{SecretsMode: specmodel.SecretsModeNone}
}

// ModifyRequest defaults the mode, so that a caller who omits it gets the
// harmless one rather than an error.
func (req *ExportSpecReq) ModifyRequest() {
	if req.SecretsMode == "" {
		req.SecretsMode = specmodel.SecretsModeNone
	}
}

func (req *ExportSpecReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4) //nolint:mnd
	if req.ProjectID != "" {
		validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	}
	if req.AppID != "" {
		validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ExportSpecResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *settings.BaseDownloadDataResp `json:"-"`
	// Report travels in a header rather than the body, since the body is the
	// bundle. The handler writes it as X-HivePaaS-Spec-Report.
	Report *specmodel.Report `json:"-"`
}
```

```go
// hivepaas_app/usecase/specuc/export.go
package specuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// ExportSpec builds a configuration bundle for a scope.
//
// Both secret-bearing modes decrypt every secret the scope owns, so both pass
// the same capability gate and leave the same audit record as any other reveal.
// The difference between them is only whether the plaintext lands on disk or
// inside an age envelope - not whether it was read.
func (uc *UC) ExportSpec(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ExportSpecReq,
) (*specdto.ExportSpecResp, error) {
	if !req.SecretsMode.IsValid() {
		return nil, hperrors.Wrap(hperrors.ErrSpecSecretsModeInvalid).
			WithParam("Mode", string(req.SecretsMode))
	}
	if req.SecretsMode == specmodel.SecretsModeEncrypted && req.Passphrase == "" {
		return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
	}

	scope, err := uc.buildScope(ctx, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if req.SecretsMode.RevealsSecrets() {
		err = uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
			Scope:    scope.ScopeType,
			ObjectID: scope.ScopeObjectID(),
			Source:   base.AuditLogSourceAPIRead,
			ResType:  base.ResourceTypeSetting,
			ResName:  fmt.Sprintf("spec export (%s)", req.SecretsMode),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	resp, err := uc.specService.Export(ctx, uc.db, &specservice.ExportReq{
		Scope:       scope,
		SecretsMode: req.SecretsMode,
		Passphrase:  req.Passphrase,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &specdto.ExportSpecResp{
		Data: &settings.BaseDownloadDataResp{
			ContentType:   "application/gzip",
			ContentLength: resp.Size,
			Content:       resp.Content,
			ExtraHeaders: map[string]string{
				"Content-Disposition": fmt.Sprintf("attachment; filename=%q", resp.Filename),
			},
		},
		Report: resp.Report,
	}, nil
}
```

Write `uc.go` following `hivepaas_app/usecase/appsettingsuc/uc.go`, and
`buildScope` mapping the request's ids onto `entity.NewObjectScopeGlobal`,
`NewObjectScopeProject`, `NewObjectScopeProjectEnv` or `NewObjectScopeApp`.
Check `base.AuditLogSourceAPIRead` exists:
`grep -n "AuditLogSourceAPI" hivepaas_app/base/audit.go`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/usecase/specuc/ -v`
Expected: PASS, five tests.

- [ ] **Step 5: Write the handler and register the routes**

Follow `basesettinghandler/download.go` for streaming a
`settings.BaseDownloadDataResp`. Add the report as a header:

```go
if resp.Report != nil && len(resp.Report.Issues) > 0 {
	reportJSON, err := json.Marshal(resp.Report)
	if err == nil {
		c.Header("X-HivePaaS-Spec-Report", string(reportJSON))
	}
}
```

Register the routes beside the scope they belong to, following
`router_apps.go`'s grouping style:

```go
{ // Spec export
	rootGroup.GET("/spec/export", specHandler.ExportGlobalSpec)
	projectGroup.GET("/spec/export", specHandler.ExportProjectSpec)
	projectEnvGroup.GET("/spec/export", specHandler.ExportProjectEnvSpec)
	appGroup.GET("/:appID/spec/export", specHandler.ExportAppSpec)
}
```

- [ ] **Step 6: Wire the DI registry**

In `hivepaas_app/registry/provides.go`, add the three imports and three
constructors alongside their neighbours:

```go
specserviceimpl.New,
specuc.New,
spechandler.New,
```

Add `specHandler` to the handler registry struct the same way `appHandler` is
declared, and add the usecase field to whatever aggregate the handler takes.

- [ ] **Step 7: Verify the whole thing builds and runs**

```bash
make fmt && make lint && make test
go build ./...
```

Then run it end to end against the local installation. The API base path is
`/_` (`config.HTTPServer.BasePath`, default `/_`), and dev mode can mint a token
without a browser:

```bash
make local-app-run

# in another shell - pick any existing user id
USER_ID=$(PGPASSWORD=abc123 psql -h localhost -p 35432 -U hivepaas -d hivepaas \
  -tAc "SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1")

TOKEN=$(curl -sS -X POST 'http://localhost:10000/_/dev-helper/dev-mode-login' \
  -H 'Content-Type: application/json' -d "{\"userId\":\"$USER_ID\"}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])')

curl -sS -D/tmp/spec-headers.txt -H "Authorization: Bearer $TOKEN" \
  -o /tmp/spec.tar.gz 'http://localhost:10000/_/spec/export?secretsMode=none'
tar -tzf /tmp/spec.tar.gz
```

Expected: `spec.yaml`, `global.yaml`, and `projects/<key>/project.yaml` plus
`projects/<key>/envs/<env>.yaml` for each project except `hivepaas`. Check
`/tmp/spec-headers.txt` for `X-HivePaaS-Spec-Report` and read what it skipped.

If the Authorization header is not how this API takes a token, check what the
auth middleware reads:
`grep -rn "Authorization\|Bearer\|cookie" hivepaas_app/interface/api/ | grep -i "middleware\|auth" | head`

Then confirm determinism against the real installation:

```bash
curl -sS -H "Authorization: Bearer $TOKEN" \
  -o /tmp/spec2.tar.gz 'http://localhost:10000/_/spec/export?secretsMode=none'
mkdir -p /tmp/s1 /tmp/s2
tar -xzf /tmp/spec.tar.gz -C /tmp/s1 && tar -xzf /tmp/spec2.tar.gz -C /tmp/s2
diff -r /tmp/s1 /tmp/s2
```

Expected: only `spec.yaml` differs, and only in `exportedAt`.

Finally confirm the two exclusions hold against real data:

```bash
# the hivepaas project must not appear
ls /tmp/s1/stage/projects/ | grep -x hivepaas && echo "BUG: hivepaas project exported"
# no plaintext secret leaked in none mode
grep -rn "hpenc:" /tmp/s1 | head   # ciphertext is fine
```

- [ ] **Step 8: Commit**

```bash
git add hivepaas_app/usecase/specuc hivepaas_app/interface/api/handler/spechandler \
        hivepaas_app/registry/provides.go hivepaas_app/interface/api/server
git commit -m "feat(spec): expose spec export through the API"
```

---

## Task 13: Documentation

**Files:**
- Modify: `docs/DEVELOPMENT.md`
- Modify: `docs/ARCHITECTURE.md`

- [ ] **Step 1: Document the export endpoints in DEVELOPMENT.md**

Add a short section under the existing numbered sections describing the four
endpoints, the three secret modes, and how to read a bundle:

```bash
curl -sS -o spec.tar.gz 'http://localhost:10000/api/v1/spec/export?secretsMode=none'
tar -xzf spec.tar.gz
# encrypted bundles open with the standard age CLI, without HivePaaS:
age -d -o spec.tar.gz spec.tar.gz.age
```

- [ ] **Step 2: Add the spec layer to ARCHITECTURE.md**

Add a row to the layer table or a short paragraph in §4 noting that
`specservice/specmodel` is a durable contract that deliberately shares no types
with the dashboard DTOs, and must not import `usecase/`.

- [ ] **Step 3: Commit**

```bash
git add docs/DEVELOPMENT.md docs/ARCHITECTURE.md
git commit -m "docs: describe configuration spec export"
```

---

## Not in this plan

The spec's §5 import contract is deliberately unimplemented. Before import is
built, one refactor is required that this plan does not perform: the logic that
applies the five Swarm blocks back onto a service lives in `appsettingsuc`
(`applyAppResourceSettings` and its siblings), and `ARCHITECTURE.md` forbids one
usecase importing another. That logic has to move down into a service both
usecases call. Export does not need it and does not wait for it.
