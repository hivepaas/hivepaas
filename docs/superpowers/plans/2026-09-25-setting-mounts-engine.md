# Setting Mounts - Backend Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An app's `app-setting-mount` entries become files in its containers that follow their source setting from then on, and the Docker API labels follow the `hivepaas.<object>.<field>` convention.

**Architecture:**
- **The setting.** A new app-scope collection type, `app-setting-mount`. The entry's key is the setting's name. The source is an `ObjectID`, so `res_link`, import remapping and `ERR_SETTING_IN_USE` come for free.
- **The package** `service/settingmountservice` holds what needs no database:
  - the parts registry;
  - rotation keys: an HMAC over a part's inputs and its version;
  - object names and labels;
  - path rules.
- **The engine** in `settingmountserviceimpl`:
  - `Resolve` reads entries and sources and renders files;
  - `ApplyToService` makes sure every file's Docker object exists and swaps the spec's mount references, inside the caller's one update;
  - `Sweep` removes what nothing references any more;
  - `Refresh` is one update followed by `Sweep`.
- **The refresh task.** `task:setting-mount-refresh` is recorded in the same transaction as the write that makes it necessary, and scheduled after commit.

**Tech Stack:** Go, bun, moby client (`services/docker`), testify.

**Spec:** `docs/superpowers/specs/2026-09-25-setting-mounts-design.md` §1-§6, §11, §12, §13 plan 1.

## Global Constraints

- **Setting type:** `app-setting-mount` (`base.SettingTypeAppSettingMount`). Block `settings.settingMounts`. App scope only, never inherited.
- **Entry key:** `^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$`, never `tls`.
- **Paths:**
  - absolute and clean;
  - not `/`;
  - not `/run/secrets/tls` and nothing under it.
- **Parts (§2):**

  | source | part | required | sensitive | notes |
  |---|---|---|---|---|
  | `ssl-cert` | `certificate` | yes | | |
  | | `privateKey` | yes | yes | |
  | | `caCertificate` | | | |
  | `ssh-key` | `privateKey` | yes | yes | |
  | | `publicKey` | | | |
  | `basic-auth` | `username` | yes | | |
  | | `password` | yes | yes | |
  | | `htpasswd` | yes | yes | `username:<bcrypt>\n` |

- **Secret or config:** a sensitive part becomes a Docker secret, any other a Docker config.
- **Names (§3):**
  - `<GlobalKey>_mount_<entry>_<part>_<hash8>` for an entry;
  - `<GlobalKey>_tls_<part>_<hash8>` for TLS (plan 3);
  - lowercased, at most 64 characters;
  - a name that would be longer cuts the `GlobalKey` and appends 8 hex characters of its SHA-256.
- **Labels:**
  - `hivepaas.app.id=<app id>`;
  - `hivepaas.settingMount.entry=<entry key | tls>`;
  - `hivepaas.settingMount.part=<part>`.
- **Docker API labels become:** `hivepaas.dockerApi.app`, `hivepaas.dockerApi.socket`, `hivepaas.dockerApi.network`.
- **File defaults:** uid `"0"`, gid `"0"`, mode `0444`.
- **Rotation key:** hex HMAC-SHA256, keyed with `config.Current().Secret`, over:
  - the source type;
  - the part's name;
  - the part's version;
  - each input, length-prefixed.

  It is never computed from rendered bytes, because bcrypt salts afresh.
- **Gates:** `go build ./...`, `golangci-lint run ./...` (whole repo; 120-char lines, US spelling), `go test ./...`. No DTO changes, so no `make gen-swag`.
- **Git:**
  - branch `feat/setting-mounts` off `main`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, delete the branch, do not push;
  - never stage files this plan does not name. The user may have uncommitted work.
- **Not in this plan** (plans 2-4):
  - entry endpoints and §7's gate;
  - export and import of entries (export skips the type here, and an app document with `settingMounts` stays refused as unsupported);
  - clone and preview handling beyond what `ApplyToService` does on a deployment;
  - TLS passthrough;
  - the dashboard.

## Review Focus

1. **A source outside the app's scope, or disabled, is never mounted.** `Resolve` reads sources through the app's scope with `requireActive`, and checks status and type again itself. *Test: Task 4, "a source the scope does not return mounts nothing".*
2. **A path claimed twice is dropped the same way every time.** Entries are taken in key order and the first claim wins. A path an ordinary secret or config already targets goes to the ordinary one. *Tests: Task 4 "two entries claiming a path", Task 5 "an ordinary secret keeps its target".*
3. **A secret that is still in use does not fail a deployment.** `Sweep` retries, then gives up quietly, and the next sweep removes it. Callers ignore its error. *Test: Task 5 "sweep gives up on what stays in use and removes the rest".*
4. **A name Docker already has is reused, not recreated.** The same data twice creates nothing the second time, and a create conflict falls back to the existing object. *Test: Task 5 "the same data twice creates nothing new".*
5. **A source that cannot be decrypted leaves the service as it was.** `ApplyToService` returns the error before touching the spec. *Test: Task 5 "a source that fails to render changes nothing".*

---

### Task 1: Rename the Docker API labels

**Files:**
- Modify: `hivepaas_app/pkg/dockerproxy/policy.go:37`
- Modify: `hivepaas_app/service/dockerapiservice/types.go:12,19`
- Modify: `docs/superpowers/specs/2026-09-24-docker-api-access-design.md:196,254`
- Test: `hivepaas_app/service/dockerapiservice/types_test.go` (create)

**Interfaces:**
- Produces: `dockerproxy.OwnerLabel = "hivepaas.dockerApi.app"`, `dockerapiservice.NetworkLabel = "hivepaas.dockerApi.network"`, `dockerapiservice.SocketVolumeLabel = "hivepaas.dockerApi.socket"`.

- [ ] **Step 1: Branch**

```bash
git checkout main && git checkout -b feat/setting-mounts
```

- [ ] **Step 2: Write the failing test** in `hivepaas_app/service/dockerapiservice/types_test.go`:

```go
package dockerapiservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

// HivePaaS labels read hivepaas.<object>.<field>, camelCase, as hivepaas.app.id
// does.
func TestDockerAPILabelsFollowTheConvention(t *testing.T) {
	assert.Equal(t, "hivepaas.dockerApi.app", dockerproxy.OwnerLabel)
	assert.Equal(t, "hivepaas.dockerApi.network", NetworkLabel)
	assert.Equal(t, "hivepaas.dockerApi.socket", SocketVolumeLabel)
}
```

- [ ] **Step 3: Run it.** `go test ./hivepaas_app/service/dockerapiservice/ -run TestDockerAPILabels` should FAIL on all three.

- [ ] **Step 4: Change the constants:**
  - `OwnerLabel = "hivepaas.dockerApi.app"`;
  - `NetworkLabel = "hivepaas.dockerApi.network"`;
  - `SocketVolumeLabel = "hivepaas.dockerApi.socket"`.

  In the access design spec, replace `hivepaas.docker-api.app` and `hivepaas.docker-api.network` with the new names. Older plan documents are history and stay as they are.

- [ ] **Step 5: Run** `go test ./hivepaas_app/pkg/dockerproxy/... ./hivepaas_app/service/dockerapiservice/... ./hivepaas_app/usecaseagent/...`. It should PASS, since every other use goes through the constants. Then run `grep -rn "docker-api\.\(app\|socket\|network\)" hivepaas_app`, which should find nothing.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/pkg/dockerproxy/policy.go hivepaas_app/service/dockerapiservice/types.go \
  hivepaas_app/service/dockerapiservice/types_test.go docs/superpowers/specs/2026-09-24-docker-api-access-design.md
git commit -m "refactor(dockerapi): labels follow hivepaas.<object>.<field>

Nothing with the old names has been released.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The `app-setting-mount` setting type

**Files:**
- Modify: `hivepaas_app/base/setting.go` (constant, alphabetical, after `SettingTypeAppRouting`)
- Create: `hivepaas_app/entity/setting_app_setting_mount.go`
- Create: `hivepaas_app/entity/setting_app_setting_mount_migration.go`
- Test: `hivepaas_app/entity/setting_app_setting_mount_test.go`
- Modify: `hivepaas_app/entity/setting_spec.go`. Add to the `registerSpecPolicy` group, next to `SettingTypeApp`:

  ```go
  _ = registerSpecPolicy(base.SettingTypeAppSettingMount, skipSpecPolicy{
  	reason: "setting mounts are exported once import checks who may mount a sensitive part",
  })
  ```
- Modify: `hivepaas_app/service/specservice/specmodel/singleton.go`. In `collectionBlockNames`, add `base.SettingTypeAppSettingMount: "settingMounts",`.
- Modify: `hivepaas_app/service/specservice/specserviceimpl/import_policy.go`:
  - a constant `reasonMountsSettings = "only an app mounts settings, and import does not check who may mount a sensitive part yet"`;
  - the entry `base.SettingTypeAppSettingMount: {skip: reasonMountsSettings},`.

**Interfaces:**
- Produces:
  - `base.SettingTypeAppSettingMount SettingType = "app-setting-mount"`;
  - `entity.CurrentAppSettingMountVersion = 1`;
  - `entity.AppSettingMount{Source ObjectID; Files []*AppSettingMountFile}`;
  - `entity.AppSettingMountFile{Part, Path, UID, GID string; Mode fileutil.FileMode}`;
  - `(*Setting).AsAppSettingMount() (*AppSettingMount, error)` and `MustAsAppSettingMount()`.

- [ ] **Step 1: Write the failing test** `hivepaas_app/entity/setting_app_setting_mount_test.go`:

```go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func TestAppSettingMountReadsWhatWasStored(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeAppSettingMount, Data: `{"source":{"id":"cert_1"},` +
		`"files":[{"part":"certificate","path":"/etc/app/tls/cert.pem"},` +
		`{"part":"privateKey","path":"/etc/app/tls/key.pem","uid":"1000","mode":"0400"}]}`}

	got, err := setting.AsAppSettingMount()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, &AppSettingMount{
		Source: ObjectID{ID: "cert_1"},
		Files: []*AppSettingMountFile{
			{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
			{Part: "privateKey", Path: "/etc/app/tls/key.pem", UID: "1000", Mode: fileutil.FileMode(0o400)},
		},
	}, got)
}

// The source is a reference like any other: written to res_link, remapped by
// import, and in use while linked.
func TestAppSettingMountReferencesItsSource(t *testing.T) {
	assert.Equal(t, []string{"cert_1"},
		(&AppSettingMount{Source: ObjectID{ID: "cert_1"}}).GetRefObjectIDs().RefSettingIDs)
	assert.Empty(t, (&AppSettingMount{}).GetRefObjectIDs().RefSettingIDs)
}
```

- [ ] **Step 2: Run it.** `go test ./hivepaas_app/entity/ -run TestAppSettingMount` should FAIL to compile.

- [ ] **Step 3: Implement.** `hivepaas_app/entity/setting_app_setting_mount.go`:

```go
package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

const (
	CurrentAppSettingMountVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppSettingMount, &appSettingMountParser{})

type appSettingMountParser struct {
}

func (s *appSettingMountParser) New() SettingData {
	return &AppSettingMount{}
}

// AppSettingMount mounts parts of another setting - a certificate and its key, a
// basic auth pair as htpasswd - as files in the app's containers, which follow
// that setting from then on. The entry's key is the setting's name. See
// docs/superpowers/specs/2026-09-25-setting-mounts-design.md.
type AppSettingMount struct {
	// Source is the setting the files come from, every one of them.
	Source ObjectID               `json:"source"`
	Files  []*AppSettingMountFile `json:"files"`
}

// AppSettingMountFile puts one part of the source at a path. UID, GID and Mode
// are those of secrets and config files when empty.
type AppSettingMountFile struct {
	Part string            `json:"part"`
	Path string            `json:"path"`
	UID  string            `json:"uid,omitempty"`
	GID  string            `json:"gid,omitempty"`
	Mode fileutil.FileMode `json:"mode,omitempty"`
}

func (s *AppSettingMount) GetType() base.SettingType {
	return base.SettingTypeAppSettingMount
}

func (s *AppSettingMount) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Source.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Source.ID)
	}
	return refIDs
}

func (s *AppSettingMount) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppSettingMount() (*AppSettingMount, error) {
	return parseSettingAs[*AppSettingMount](s)
}

func (s *Setting) MustAsAppSettingMount() *AppSettingMount {
	return gofn.Must(s.AsAppSettingMount())
}
```

`hivepaas_app/entity/setting_app_setting_mount_migration.go`:

```go
package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppSettingMount) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppSettingMountVersion {
		return false, nil
	}
	if setting.Version > CurrentAppSettingMountVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so an older row is one written before versions
	// were set: the data is already in its shape.
	setting.Version = CurrentAppSettingMountVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
```

Then add the constant and the three registrations listed under **Files**.

- [ ] **Step 4: Run** `go test ./hivepaas_app/entity/... ./hivepaas_app/service/specservice/...`. It should PASS, including:
  - `TestEverySettingTypeHasASpecPolicy`;
  - `TestEverySettingTypeIsClassifiedExactlyOnce`;
  - `TestBlockNamesAreUnique`;
  - `TestEveryBlockTypeHasAnImportPolicy`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/base/setting.go hivepaas_app/entity/setting_app_setting_mount*.go \
  hivepaas_app/entity/setting_spec.go hivepaas_app/service/specservice/specmodel/singleton.go \
  hivepaas_app/service/specservice/specserviceimpl/import_policy.go
git commit -m "feat(settings): the app-setting-mount setting type

Export and import skip it until they check who may mount a sensitive part.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: Parts, rotation keys, names, labels and paths

**Files:**
- Create: `hivepaas_app/service/settingmountservice/parts.go`
- Create: `hivepaas_app/service/settingmountservice/names.go`
- Create: `hivepaas_app/service/settingmountservice/paths.go`
- Test: `hivepaas_app/service/settingmountservice/parts_test.go`, `names_test.go`, `paths_test.go`

**Interfaces:**
- Consumes: `entity.SSLCert`, `entity.SSHKey`, `entity.BasicAuth`, `EncryptedField.GetPlain()`, `htpasswd.HashPassword`.
- Produces:
  - `type Part struct{Name string; Required, Sensitive bool; Version int; Inputs []string; Render func([]string) ([]byte, error)}`;
  - `SourceTypes() []base.SettingType`, `IsSourceType(base.SettingType) bool`;
  - `PartsOf(base.SettingType) []*Part`, `PartOf(base.SettingType, string) *Part`;
  - `Values(*entity.Setting) (map[string]string, error)`;
  - `Usable(base.SettingType, map[string]string) bool`;
  - `(*Part).Empty(map[string]string) bool`, `(*Part).RenderFrom(map[string]string) ([]byte, error)`;
  - `RotationKey(key []byte, typ base.SettingType, part *Part, values map[string]string) string`;
  - constants `LabelAppID`, `LabelEntry`, `LabelPart`, `TLSEntry`, `TLSDir`;
  - `ValidEntryKey(string) bool`;
  - `ObjectName(globalKey, entry, part, rotation string) string`;
  - `Labels(appID, entry, part string) map[string]string`;
  - `ValidPath(string) bool`;
  - `SecretTarget(name string) string`, `ConfigTarget(name string) string`.

- [ ] **Step 1: Write the failing tests.** `parts_test.go`:

```go
package settingmountservice

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

var testKey = []byte("rotation-key")

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

func sourceSetting(t *testing.T, typ base.SettingType, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "src_1", Type: typ, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func TestEachPartRendersFromItsSource(t *testing.T) {
	useDataKey(t)
	cert := sourceSetting(t, base.SettingTypeSSLCert, &entity.SSLCert{
		Certificate: "CERT", PrivateKey: entity.NewEncryptedField("KEY"), CACertificate: "CA"})
	ssh := sourceSetting(t, base.SettingTypeSSHKey, &entity.SSHKey{
		PublicKey: "PUB", PrivateKey: entity.NewEncryptedField("PRIV")})
	auth := sourceSetting(t, base.SettingTypeBasicAuth, &entity.BasicAuth{
		Username: "admin", Password: entity.NewEncryptedField("s3cret")})

	for _, tc := range []struct {
		setting *entity.Setting
		part    string
		want    string
	}{
		{cert, "certificate", "CERT"}, {cert, "privateKey", "KEY"}, {cert, "caCertificate", "CA"},
		{ssh, "privateKey", "PRIV"}, {ssh, "publicKey", "PUB"},
		{auth, "username", "admin"}, {auth, "password", "s3cret"},
	} {
		values, err := Values(tc.setting)
		assert.NoError(t, err)
		out, err := PartOf(tc.setting.Type, tc.part).RenderFrom(values)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, string(out), "%s %s", tc.setting.Type, tc.part)
	}
}

func TestHtpasswdIsAUsernameAndABcryptOfThePassword(t *testing.T) {
	useDataKey(t)
	auth := sourceSetting(t, base.SettingTypeBasicAuth, &entity.BasicAuth{
		Username: "admin", Password: entity.NewEncryptedField("s3cret")})
	values, err := Values(auth)
	assert.NoError(t, err)

	out, err := PartOf(base.SettingTypeBasicAuth, "htpasswd").RenderFrom(values)
	assert.NoError(t, err)

	user, hash, found := strings.Cut(strings.TrimSuffix(string(out), "\n"), ":")
	assert.True(t, found)
	assert.Equal(t, "admin", user)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("s3cret")))
}

func TestSensitivePartsAreTheSecretOnes(t *testing.T) {
	sensitive := map[string]bool{}
	for _, typ := range SourceTypes() {
		for _, part := range PartsOf(typ) {
			if part.Sensitive {
				sensitive[string(typ)+"/"+part.Name] = true
			}
		}
	}
	assert.Equal(t, map[string]bool{
		"ssl-cert/privateKey": true, "ssh-key/privateKey": true,
		"basic-auth/password": true, "basic-auth/htpasswd": true,
	}, sensitive)
}

// bcrypt salts every hash afresh, so the key reads the inputs, never the output:
// comparing output would restart the app on every refresh.
func TestTheRotationKeyFollowsInputsAndVersionOnly(t *testing.T) {
	part := PartOf(base.SettingTypeBasicAuth, "htpasswd")
	values := map[string]string{"username": "admin", "password": "s3cret"}
	key := RotationKey(testKey, base.SettingTypeBasicAuth, part, values)

	assert.Equal(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, part, values), "stable")
	assert.NotEqual(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, part,
		map[string]string{"username": "admin", "password": "other"}), "an input changed")
	bumped := *part
	bumped.Version++
	assert.NotEqual(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, &bumped, values), "the version changed")
	assert.NotEqual(t, key, RotationKey([]byte("another-key"), base.SettingTypeBasicAuth, part, values),
		"keyed: a name says nothing of a password to whoever lists secrets")
	// Inputs are length-prefixed: moving a character from one to the next is a change.
	assert.NotEqual(t,
		RotationKey(testKey, base.SettingTypeBasicAuth, part, map[string]string{"username": "ab", "password": "c"}),
		RotationKey(testKey, base.SettingTypeBasicAuth, part, map[string]string{"username": "a", "password": "bc"}))
}

func TestASourceIsUsableWhenEveryRequiredPartHasItsInputs(t *testing.T) {
	assert.True(t, Usable(base.SettingTypeSSLCert, map[string]string{"certificate": "C", "privateKey": "K"}))
	assert.False(t, Usable(base.SettingTypeSSLCert, map[string]string{"certificate": "", "privateKey": "K"}),
		"a certificate not obtained yet")
	assert.False(t, Usable(base.SettingType("secret"), map[string]string{}), "not a source type")
}
```

`names_test.go`:

```go
package settingmountservice

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rotation = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestObjectNamesFollowTheSecretsConvention(t *testing.T) {
	assert.Equal(t, "shop_prod_api_mount_tls-cert_privatekey_01234567",
		ObjectName("Shop_Prod_API", "tls-cert", "privateKey", rotation))
	assert.Equal(t, "shop_prod_api_tls_certificate_01234567",
		ObjectName("shop_prod_api", TLSEntry, "certificate", rotation))
}

// Docker caps names at 64; the GlobalKey is cut and a hash of it keeps two long
// keys that share a beginning apart.
func TestALongNameIsShortenedAndStaysUnique(t *testing.T) {
	long := strings.Repeat("project-with-a-long-name_", 3) + "env_app"
	a := ObjectName(long+"-a", "abcdefghij0123456789", "caCertificate", rotation)
	b := ObjectName(long+"-b", "abcdefghij0123456789", "caCertificate", rotation)
	assert.LessOrEqual(t, len(a), 64)
	assert.LessOrEqual(t, len(b), 64)
	assert.NotEqual(t, a, b)
	assert.True(t, strings.HasSuffix(a, "_mount_abcdefghij0123456789_cacertificate_01234567"), a)
}

func TestEntryKeys(t *testing.T) {
	for key, want := range map[string]bool{
		"tls-cert": true, "a": true, "a1-b2": true, strings.Repeat("a", 20): true,
		"tls": false, "": false, "-a": false, "a-": false, "A": false, "a_b": false, strings.Repeat("a", 21): false,
	} {
		assert.Equal(t, want, ValidEntryKey(key), key)
	}
}

func TestLabels(t *testing.T) {
	assert.Equal(t, map[string]string{
		"hivepaas.app.id": "app_1", "hivepaas.settingMount.entry": "tls-cert", "hivepaas.settingMount.part": "privateKey",
	}, Labels("app_1", "tls-cert", "privateKey"))
}
```

`paths_test.go`:

```go
package settingmountservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	for path, want := range map[string]bool{
		"/etc/app/tls/cert.pem": true, "/run/secrets/app.pem": true, "/run/secrets/tlsx": true,
		"": false, "/": false, "etc/app": false, "/etc/../app": false, "/etc/app/": false,
		"/run/secrets/tls": false, "/run/secrets/tls/cert.pem": false,
	} {
		assert.Equal(t, want, ValidPath(path), path)
	}
}

func TestTargetsOfOrdinaryReferences(t *testing.T) {
	assert.Equal(t, "/run/secrets/db_password", SecretTarget("db_password"))
	assert.Equal(t, "/etc/app/key", SecretTarget("/etc/app/key"))
	assert.Equal(t, "/app.conf", ConfigTarget("app.conf"))
	assert.Equal(t, "/etc/app.conf", ConfigTarget("/etc/app.conf"))
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/settingmountservice/`. It should FAIL to compile.

- [ ] **Step 3: Implement.** `parts.go`:

```go
// Package settingmountservice mounts parts of settings - a certificate and its
// key, a basic auth pair as htpasswd - as files in an app's containers, which
// follow those settings from then on. What needs no database lives here: the
// parts a source type offers, how a file is named and labeled, and where it may
// go. See docs/superpowers/specs/2026-09-25-setting-mounts-design.md.
package settingmountservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"slices"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/htpasswd"
)

// Part is one file a source type offers. A field is the simplest part; a virtual
// one such as htpasswd is computed. The engine does not tell them apart.
type Part struct {
	Name string
	// Required parts must all have their inputs for the source to be mounted at
	// all; an optional part without them is left out alone.
	Required bool
	// Sensitive parts become Docker secrets, and mounting one is revealing it.
	Sensitive bool
	// Version is raised whenever Render's output changes shape for the same
	// inputs, which is what replaces files already mounted.
	Version int
	// Inputs are the source's values the part reads, as Values names them.
	Inputs []string
	// Render turns the inputs' values, in Inputs' order, into the file. Nil
	// writes the one input as it is.
	Render func(values []string) ([]byte, error)
}

type sourceType struct {
	parts  []*Part
	values func(*entity.Setting) (map[string]string, error)
}

var registry = map[base.SettingType]*sourceType{
	base.SettingTypeSSLCert: {
		parts: []*Part{
			{Name: "certificate", Required: true, Version: 1, Inputs: []string{"certificate"}},
			{Name: "privateKey", Required: true, Sensitive: true, Version: 1, Inputs: []string{"privateKey"}},
			{Name: "caCertificate", Version: 1, Inputs: []string{"caCertificate"}},
		},
		values: sslCertValues,
	},
	base.SettingTypeSSHKey: {
		parts: []*Part{
			{Name: "privateKey", Required: true, Sensitive: true, Version: 1, Inputs: []string{"privateKey"}},
			{Name: "publicKey", Version: 1, Inputs: []string{"publicKey"}},
		},
		values: sshKeyValues,
	},
	base.SettingTypeBasicAuth: {
		parts: []*Part{
			{Name: "username", Required: true, Version: 1, Inputs: []string{"username"}},
			{Name: "password", Required: true, Sensitive: true, Version: 1, Inputs: []string{"password"}},
			// bcrypt, which Traefik, Apache, Caddy and HivePaaS's registry read,
			// and nginx where the system's crypt is libxcrypt.
			{Name: "htpasswd", Required: true, Sensitive: true, Version: 1,
				Inputs: []string{"username", "password"}, Render: renderHtpasswd},
		},
		values: basicAuthValues,
	},
}

// SourceTypes are the setting types an entry may mount from.
func SourceTypes() []base.SettingType {
	types := make([]base.SettingType, 0, len(registry))
	for typ := range registry {
		types = append(types, typ)
	}
	slices.Sort(types)
	return types
}

func IsSourceType(typ base.SettingType) bool {
	return registry[typ] != nil
}

// PartsOf are the parts a source type offers, nil for a type that is not one.
func PartsOf(typ base.SettingType) []*Part {
	if src := registry[typ]; src != nil {
		return src.parts
	}
	return nil
}

// PartOf is a source type's part by name, nil when it offers none of that name.
func PartOf(typ base.SettingType, name string) *Part {
	for _, part := range PartsOf(typ) {
		if part.Name == name {
			return part
		}
	}
	return nil
}

// Values are what a source's parts read, decrypted, by input name.
func Values(setting *entity.Setting) (map[string]string, error) {
	src := registry[setting.Type]
	if src == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("%s is not a mount source", setting.Type)
	}
	values, err := src.values(setting)
	return values, hperrors.Wrap(err)
}

// Usable reports whether every required part of a source has its inputs: a
// setting that cannot be used is not in the container.
func Usable(typ base.SettingType, values map[string]string) bool {
	if !IsSourceType(typ) {
		return false
	}
	for _, part := range PartsOf(typ) {
		if part.Required && part.Empty(values) {
			return false
		}
	}
	return true
}

// Empty reports whether one of the part's inputs is.
func (p *Part) Empty(values map[string]string) bool {
	for _, input := range p.Inputs {
		if values[input] == "" {
			return true
		}
	}
	return false
}

// RenderFrom is the file's bytes.
func (p *Part) RenderFrom(values map[string]string) ([]byte, error) {
	inputs := p.inputs(values)
	if p.Render == nil {
		return []byte(inputs[0]), nil
	}
	out, err := p.Render(inputs)
	return out, hperrors.Wrap(err)
}

func (p *Part) inputs(values map[string]string) []string {
	inputs := make([]string, len(p.Inputs))
	for i, name := range p.Inputs {
		inputs[i] = values[name]
	}
	return inputs
}

// RotationKey decides whether a part's file is new: it changes with the part's
// inputs and version, and with nothing else. It is keyed so that the name it
// ends up in says nothing of a password to whoever can list secrets.
func RotationKey(key []byte, typ base.SettingType, part *Part, values map[string]string) string {
	mac := hmac.New(sha256.New, key)
	for _, field := range append([]string{string(typ), part.Name, strconv.Itoa(part.Version)},
		part.inputs(values)...) {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		mac.Write(size[:])
		mac.Write([]byte(field))
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func renderHtpasswd(values []string) ([]byte, error) {
	hash, err := htpasswd.HashPassword(values[1])
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return []byte(values[0] + ":" + hash + "\n"), nil
}

func sslCertValues(setting *entity.Setting) (map[string]string, error) {
	cert, err := setting.AsSSLCert()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	key, err := cert.PrivateKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{"certificate": cert.Certificate, "privateKey": key,
		"caCertificate": cert.CACertificate}, nil
}

func sshKeyValues(setting *entity.Setting) (map[string]string, error) {
	sshKey, err := setting.AsSSHKey()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	key, err := sshKey.PrivateKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{"privateKey": key, "publicKey": sshKey.PublicKey}, nil
}

func basicAuthValues(setting *entity.Setting) (map[string]string, error) {
	auth, err := setting.AsBasicAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	password, err := auth.Password.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{"username": auth.Username, "password": password}, nil
}
```

Before writing this, check that `AsSSLCert`, `AsSSHKey` and `AsBasicAuth` exist in `entity/setting_ssl_cert.go`, `setting_ssh_key.go` and `setting_basic_auth.go` (`grep -n "func (s \*Setting) As" hivepaas_app/entity/setting_s*.go hivepaas_app/entity/setting_basic_auth.go`). If `golang.org/x/crypto/bcrypt` is not in `vendor/modules.txt` for the test, use `pkg/htpasswd`'s own verify function if it has one (`grep -n "^func" hivepaas_app/pkg/htpasswd/*.go`).

`names.go`:

```go
package settingmountservice

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const (
	// LabelAppID is the app a mounted object belongs to, as elsewhere
	// (appservice.LabelLogAppID).
	LabelAppID = "hivepaas.app.id"
	// LabelEntry marks an object as a mounted setting's, with the entry's key,
	// or TLSEntry.
	LabelEntry = "hivepaas.settingMount.entry"
	// LabelPart is the part the object holds.
	LabelPart = "hivepaas.settingMount.part"

	// TLSEntry is the entry TLS passthrough mounts under, which no entry may be
	// called.
	TLSEntry = "tls"

	maxNameLen = 64
	hashLen    = 8
)

var entryKeyPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$`)

// ValidEntryKey reports whether an entry may be called key.
func ValidEntryKey(key string) bool {
	return key != TLSEntry && entryKeyPattern.MatchString(key)
}

// ObjectName is the Docker secret or config a file is held in, named as secrets
// and config files are: GlobalKey + "_" + a name, lowercased. Docker caps names
// at 64 characters and a GlobalKey can be longer; such a name keeps as much of
// the GlobalKey as fits, and a hash of the whole of it.
func ObjectName(globalKey, entry, part, rotation string) string {
	middle := "_mount_" + entry
	if entry == TLSEntry {
		middle = "_tls"
	}
	suffix := strings.ToLower(middle + "_" + part + "_" + rotation[:min(hashLen, len(rotation))])
	prefix := strings.ToLower(globalKey)
	if len(prefix)+len(suffix) > maxNameLen {
		sum := sha256.Sum256([]byte(prefix))
		keep := max(0, maxNameLen-len(suffix)-hashLen)
		prefix = prefix[:min(keep, len(prefix))] + hex.EncodeToString(sum[:])[:hashLen]
	}
	return prefix + suffix
}

// Labels are those of a mounted object.
func Labels(appID, entry, part string) map[string]string {
	return map[string]string{LabelAppID: appID, LabelEntry: entry, LabelPart: part}
}
```

`paths.go`:

```go
package settingmountservice

import (
	"path"
	"strings"
)

// TLSDir is where TLS passthrough mounts, which no entry may reach into.
const TLSDir = "/run/secrets/tls"

const (
	secretsDir = "/run/secrets"
	configsDir = "/"
)

// ValidPath reports whether a file may be mounted at p: absolute, clean, not the
// root, and outside TLSDir.
func ValidPath(p string) bool {
	if p == "" || p == "/" || !path.IsAbs(p) || path.Clean(p) != p {
		return false
	}
	return p != TLSDir && !strings.HasPrefix(p, TLSDir+"/")
}

// SecretTarget is where a secret reference's file lands: a relative name is
// under /run/secrets.
func SecretTarget(name string) string {
	if path.IsAbs(name) {
		return path.Clean(name)
	}
	return path.Join(secretsDir, name)
}

// ConfigTarget is where a config reference's file lands: a relative name is
// under the root.
func ConfigTarget(name string) string {
	if path.IsAbs(name) {
		return path.Clean(name)
	}
	return path.Join(configsDir, name)
}
```

- [ ] **Step 4: Run** `go test ./hivepaas_app/service/settingmountservice/`. It should PASS.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/settingmountservice/
git commit -m "feat(settingmounts): the parts registry, rotation keys, names and paths

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: `Resolve`, the files an app should have

**Files:**
- Create: `hivepaas_app/service/settingmountservice/service.go`
- Create: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/service.go`
- Create: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/resolve.go`
- Test: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/resolve_test.go`

**Interfaces:**
- Consumes: everything Task 3 produces; `settingRepo.List`, `settingRepo.ListByIDs(ctx, db, scope, ids, requireActive, opts...)`.
- Produces:
  - the `settingmountservice.Service` interface below;
  - `settingmountservice.File`;
  - `settingmountserviceimpl.New(...)`;
  - the seams `loadEntries`, `loadSources`, `loadReaders` and `rotationKey`, which tests replace.

- [ ] **Step 1: Write the interface.** `service.go`:

```go
package settingmountservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

type Service interface {
	// Resolve is the files app should have now: those of its active entries
	// whose source is active, visible from its scope and usable.
	Resolve(ctx context.Context, db database.IDB, app *entity.App) ([]*File, error)
	// ApplyToService brings spec's mounted files to what Resolve says, creating
	// the objects Docker does not have yet. It is called inside the caller's
	// service update, and again on each retry of it.
	ApplyToService(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) error
	// Sweep removes the app's mounted objects its service no longer references.
	// It is called once the update is made; what stays in use is left for the
	// next sweep.
	Sweep(ctx context.Context, app *entity.App) error
	// Refresh is one service update that applies, then a sweep.
	Refresh(ctx context.Context, db database.IDB, app *entity.App) error
	// RemoveApp removes every mounted object of an app whose service is gone.
	RemoveApp(ctx context.Context, appID string) error

	// RecordRefresh records, in db's transaction, a refresh of the apps that
	// read one of settings: an entry's own app, or the apps whose entries mount
	// a source. It records nothing, and returns nil, when no app does.
	RecordRefresh(ctx context.Context, db database.IDB, settings ...*entity.Setting) (*entity.Task, error)
	// Schedule hands recorded tasks to the queue once their transaction has
	// committed. A failure is left to the queue's own scan.
	Schedule(ctx context.Context, tasks ...*entity.Task)
}

// File is one part of a source, at a path of an app's containers.
type File struct {
	// Entry is the entry's key, or TLSEntry.
	Entry     string
	Part      string
	Path      string
	UID       string
	GID       string
	Mode      fileutil.FileMode
	Sensitive bool
	Data      []byte
	// Rotation is the part's RotationKey: a new one is a new object.
	Rotation string
}
```

- [ ] **Step 2: Write the failing tests.** `resolve_test.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

var testApp = &entity.App{ID: "app_1", GlobalKey: "shop_prod_api", ServiceID: "svc_1"}

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

func entry(t *testing.T, key string, status base.SettingStatus, data *entity.AppSettingMount) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "entry_" + key, Type: base.SettingTypeAppSettingMount,
		Scope: base.ObjectScopeApp, ObjectID: testApp.ID, Name: key, Status: status}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func certSource(t *testing.T, id, cert, key string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: base.SettingTypeSSLCert, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(&entity.SSLCert{Certificate: cert,
		PrivateKey: entity.NewEncryptedField(key)}))
	return setting
}

func certFiles(source string) *entity.AppSettingMount {
	return &entity.AppSettingMount{Source: entity.ObjectID{ID: source}, Files: []*entity.AppSettingMountFile{
		{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
		{Part: "privateKey", Path: "/etc/app/tls/key.pem", UID: "1000", Mode: 0o400},
		{Part: "caCertificate", Path: "/etc/app/tls/ca.pem"},
	}}
}

// fixture is a service whose seams answer from memory: entries of the app, and
// the sources its scope sees.
func fixture(t *testing.T, entries []*entity.Setting, sources ...*entity.Setting) *service {
	t.Helper()
	useDataKey(t)
	svc := &service{rotationKey: func() []byte { return []byte("k") }, removalRetryDelay: time.Millisecond}
	svc.loadEntries = func(context.Context, database.IDB, string) ([]*entity.Setting, error) {
		return entries, nil
	}
	svc.loadSources = func(_ context.Context, _ database.IDB, _ *entity.App, ids []string) ([]*entity.Setting, error) {
		var out []*entity.Setting
		for _, source := range sources {
			for _, id := range ids {
				if source.ID == id {
					out = append(out, source)
				}
			}
		}
		return out, nil
	}
	return svc
}

func paths(files []*settingmountservice.File) []string {
	var out []string
	for _, file := range files {
		out = append(out, file.Path)
	}
	return out
}

func TestResolveRendersAnEntrysFiles(t *testing.T) {
	svc := fixture(t, []*entity.Setting{entry(t, "tls-cert", base.SettingStatusActive, certFiles("cert_1"))},
		certSource(t, "cert_1", "CERT", "KEY"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	if !assert.Len(t, files, 2, "no CA certificate, so no ca.pem") {
		return
	}
	assert.Equal(t, &settingmountservice.File{Entry: "tls-cert", Part: "certificate", Path: "/etc/app/tls/cert.pem",
		UID: "0", GID: "0", Mode: fileutil.FileMode(0o444), Data: []byte("CERT"), Rotation: files[0].Rotation},
		files[0])
	assert.Equal(t, "1000", files[1].UID)
	assert.Equal(t, fileutil.FileMode(0o400), files[1].Mode)
	assert.True(t, files[1].Sensitive)
	assert.Equal(t, []byte("KEY"), files[1].Data)
	assert.Len(t, files[0].Rotation, 64)
}

func TestResolveLeavesOutWhatCannotBeUsed(t *testing.T) {
	for name, tc := range map[string]struct {
		entry   *entity.Setting
		sources []*entity.Setting
	}{
		"a disabled entry": {entry(t, "a", base.SettingStatusDisabled, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
		"a source the scope does not return mounts nothing": {entry(t, "a", base.SettingStatusActive,
			certFiles("cert_other")), []*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
		"a certificate not obtained yet": {entry(t, "a", base.SettingStatusActive, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "", "")}},
		"a key no entry may have": {entry(t, "tls", base.SettingStatusActive, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
	} {
		svc := fixture(t, []*entity.Setting{tc.entry}, tc.sources...)
		files, err := svc.Resolve(context.Background(), nil, testApp)
		assert.NoError(t, err, name)
		assert.Empty(t, files, name)
	}
	disabled := certSource(t, "cert_1", "CERT", "KEY")
	disabled.Status = base.SettingStatusDisabled
	svc := fixture(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))}, disabled)
	files, err := svc.Resolve(context.Background(), nil, testApp)
	assert.NoError(t, err)
	assert.Empty(t, files, "a disabled source, should the scope return it")
}

func TestResolveSkipsAFileThatIsWrongInItself(t *testing.T) {
	svc := fixture(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"}, Files: []*entity.AppSettingMountFile{
			{Part: "certificate", Path: "relative/cert.pem"},
			{Part: "certificate", Path: "/run/secrets/tls/cert.pem"},
			{Part: "htpasswd", Path: "/etc/htpasswd"},
			{Part: "privateKey", Path: "/etc/key.pem"},
		}})}, certSource(t, "cert_1", "CERT", "KEY"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/etc/key.pem"}, paths(files))
}

// Entries are taken in key order and the first claim to a path wins: which file
// lands there never depends on the order rows come back in.
func TestTwoEntriesClaimingAPathKeepTheFirstByKey(t *testing.T) {
	second := entry(t, "b", base.SettingStatusActive, certFiles("cert_2"))
	first := entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))
	svc := fixture(t, []*entity.Setting{second, first},
		certSource(t, "cert_1", "ONE", "K1"), certSource(t, "cert_2", "TWO", "K2"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/etc/app/tls/cert.pem", "/etc/app/tls/key.pem"}, paths(files))
	assert.Equal(t, []byte("ONE"), files[0].Data)
}
```

- [ ] **Step 3: Run.** `go test ./hivepaas_app/service/settingmountservice/...` should FAIL to compile.

- [ ] **Step 4: Implement.** `settingmountserviceimpl/service.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
)

type service struct {
	resLinkRepo repository.ResLinkRepo
	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	dockerManager docker.Manager
	taskQueue     queue.TaskQueue
	logger        logging.Logger

	// removalRetryDelay is how long Sweep waits before asking again to remove
	// an object a service still holds.
	removalRetryDelay time.Duration

	// The seams below are what tests replace: bunex options are opaque closures,
	// which a test double of the repositories cannot read.
	rotationKey func() []byte
	loadEntries func(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error)
	loadSources func(ctx context.Context, db database.IDB, app *entity.App, ids []string) ([]*entity.Setting, error)
	loadReaders func(ctx context.Context, db database.IDB, sourceIDs []string) ([]string, error)
}

func New(
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	dockerManager docker.Manager,
	taskQueue queue.TaskQueue,
	logger logging.Logger,
) settingmountservice.Service {
	s := &service{
		resLinkRepo:   resLinkRepo,
		settingRepo:   settingRepo,
		taskRepo:      taskRepo,
		dockerManager: dockerManager,
		taskQueue:     taskQueue,
		logger:        logger,

		removalRetryDelay: 2 * time.Second,
	}
	// Read on every use: the app secret can be replaced while running, which
	// renames every mounted object once, at its next refresh.
	s.rotationKey = func() []byte { return []byte(config.Current().Secret) }
	s.loadEntries = s.loadEntriesFromRepo
	s.loadSources = s.loadSourcesFromRepo
	s.loadReaders = s.loadReadersFromRepo
	return s
}

// loadEntriesFromRepo is the app's own entries, whatever their status: an entry
// is never inherited.
func (s *service) loadEntriesFromRepo(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error) {
	entries, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
		bunex.SelectWhere("setting.object_id = ?", appID),
	)
	return entries, hperrors.Wrap(err)
}

// loadSourcesFromRepo is the active sources among ids that the app's scope sees.
func (s *service) loadSourcesFromRepo(
	ctx context.Context, db database.IDB, app *entity.App, ids []string,
) ([]*entity.Setting, error) {
	sources, err := s.settingRepo.ListByIDs(ctx, db, app.GetObjectScope(), ids, true)
	return sources, hperrors.Wrap(err)
}
```

`resolve.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"slices"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const (
	defaultUID  = "0"
	defaultGID  = "0"
	defaultMode = fileutil.FileMode(0o444)
)

func (s *service) Resolve(ctx context.Context, db database.IDB, app *entity.App) ([]*settingmountservice.File, error) {
	entries, err := s.loadEntries(ctx, db, app.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries = gofn.Filter(entries, func(e *entity.Setting) bool {
		return e.Status == base.SettingStatusActive && settingmountservice.ValidEntryKey(e.Name)
	})
	// Key order, so that the first claim to a path is the same one every time.
	slices.SortFunc(entries, func(a, b *entity.Setting) int { return strings.Compare(a.Name, b.Name) })

	mounts := make(map[string]*entity.AppSettingMount, len(entries))
	var sourceIDs []string
	for _, e := range entries {
		mount, err := e.AsAppSettingMount()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		mounts[e.ID] = mount
		sourceIDs = append(sourceIDs, mount.Source.ID)
	}
	sourceIDs = gofn.ToSet(gofn.ToSliceSkippingZero(sourceIDs...))
	if len(sourceIDs) == 0 {
		return nil, nil
	}
	sources, err := s.loadSources(ctx, db, app, sourceIDs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Setting, len(sources))
	for _, source := range sources {
		byID[source.ID] = source
	}

	key := s.rotationKey()
	claimed := map[string]bool{}
	var files []*settingmountservice.File
	for _, e := range entries {
		entryFiles, err := entryFiles(key, e.Name, mounts[e.ID], byID[mounts[e.ID].Source.ID])
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, file := range entryFiles {
			if claimed[file.Path] {
				continue
			}
			claimed[file.Path] = true
			files = append(files, file)
		}
	}
	return files, nil
}

// entryFiles is what one entry mounts: nothing when its source is missing,
// disabled or unusable, and none of the files naming a part the source does not
// offer or a path no file may have.
func entryFiles(
	key []byte, entryKey string, mount *entity.AppSettingMount, source *entity.Setting,
) ([]*settingmountservice.File, error) {
	if source == nil || source.Status != base.SettingStatusActive ||
		!settingmountservice.IsSourceType(source.Type) {
		return nil, nil
	}
	values, err := settingmountservice.Values(source)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !settingmountservice.Usable(source.Type, values) {
		return nil, nil
	}
	var files []*settingmountservice.File
	seenParts := map[string]bool{}
	for _, f := range mount.Files {
		part := settingmountservice.PartOf(source.Type, f.Part)
		if part == nil || seenParts[f.Part] || !settingmountservice.ValidPath(f.Path) || part.Empty(values) {
			continue
		}
		seenParts[f.Part] = true
		data, err := part.RenderFrom(values)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		files = append(files, &settingmountservice.File{
			Entry: entryKey, Part: part.Name, Path: f.Path,
			UID:  gofn.Coalesce(f.UID, defaultUID),
			GID:  gofn.Coalesce(f.GID, defaultGID),
			Mode: gofn.Coalesce(f.Mode, defaultMode),
			Sensitive: part.Sensitive, Data: data,
			Rotation: settingmountservice.RotationKey(key, source.Type, part, values),
		})
	}
	return files, nil
}
```

Check that `gofn.ToSet` and `gofn.ToSliceSkippingZero` exist in the vendored gofn (`grep -n "^func ToSet\b\|^func ToSliceSkippingZero" vendor/github.com/tiendc/gofn/*.go`). If not, dedupe with a map. The service struct lists no `appRepo`: the executor loads apps itself (Task 6).

- [ ] **Step 5: Run** `go test ./hivepaas_app/service/settingmountservice/...`. It should PASS. Then run `go vet ./hivepaas_app/service/settingmountservice/...`. The package does not implement `Service` in full until Task 5, so add `var _ settingmountservice.Service = (*service)(nil)` only in Task 5.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/settingmountservice/
git commit -m "feat(settingmounts): resolve the files an app's entries give it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: `ApplyToService`, `Sweep`, `Refresh`, `RemoveApp`

**Files:**
- Create: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/apply.go`
- Create: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/sweep.go`
- Test: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/fake_docker_test.go`, `apply_test.go`

**Interfaces:**
- Consumes:
  - `Resolve` (Task 4);
  - `ObjectName`, `Labels`, `SecretTarget`, `ConfigTarget` and the label constants (Task 3);
  - from `docker.Manager`: `SecretList/Create/Remove`, `ConfigList/Create/Remove`, `ServiceInspect`, `ServiceUpdateFunc`.
- Produces: `ApplyToService`, `Sweep`, `Refresh` and `RemoveApp` as the interface says. In `ApplyToService`, a reference is "mounted" when it names an object carrying `LabelEntry`, whichever app it is labeled for. This is what drops references a preview or a clone copied from another app.

- [ ] **Step 1: Write the fake Docker.** `fake_docker_test.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker keeps secrets, configs and one service in memory, and filters by
// label as the daemon does: "key" is present, "key=value" matches.
type fakeDocker struct {
	docker.Manager
	secrets map[string]swarm.Secret
	configs map[string]swarm.Config
	service *swarm.Service
	nextID  int
	created []string
	updates int
	// inUse is how many more removals of an id fail, as a secret a service still
	// references does.
	inUse map[string]int
}

func newFakeDocker(spec swarm.ServiceSpec) *fakeDocker {
	if spec.TaskTemplate.ContainerSpec == nil {
		spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	return &fakeDocker{secrets: map[string]swarm.Secret{}, configs: map[string]swarm.Config{},
		service: &swarm.Service{ID: "svc_1", Spec: spec}, inUse: map[string]int{}}
}

func matches(labels map[string]string, filters client.Filters) bool {
	for want := range filters["label"] {
		key, value, withValue := strings.Cut(want, "=")
		got, ok := labels[key]
		if !ok || (withValue && got != value) {
			return false
		}
	}
	return true
}

func (f *fakeDocker) id(kind string) string {
	f.nextID++
	return fmt.Sprintf("%s_%d", kind, f.nextID)
}

func (f *fakeDocker) SecretList(_ context.Context, options ...docker.SecretListOption) (*client.SecretListResult, error) {
	opts := client.SecretListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.SecretListResult{}
	for _, secret := range f.secrets {
		if matches(secret.Spec.Labels, opts.Filters) {
			out.Items = append(out.Items, secret)
		}
	}
	return out, nil
}

func (f *fakeDocker) SecretCreate(
	_ context.Context, name string, data []byte, options ...docker.SecretCreateOption,
) (*client.SecretCreateResult, error) {
	for _, secret := range f.secrets {
		if secret.Spec.Name == name {
			return nil, hperrors.Wrap(hperrors.ErrInfraAlreadyExists)
		}
	}
	opts := client.SecretCreateOptions{}
	opts.Spec.Name, opts.Spec.Data = name, data
	for _, opt := range options {
		opt(&opts)
	}
	id := f.id("secret")
	f.secrets[id] = swarm.Secret{ID: id, Spec: opts.Spec}
	f.created = append(f.created, name)
	return &client.SecretCreateResult{ID: id}, nil
}

func (f *fakeDocker) SecretRemove(
	_ context.Context, id string, _ ...docker.SecretRemoveOption,
) (*client.SecretRemoveResult, error) {
	if f.inUse[id] > 0 {
		f.inUse[id]--
		return nil, hperrors.Wrap(hperrors.ErrInfraConflict)
	}
	delete(f.secrets, id)
	return &client.SecretRemoveResult{}, nil
}

func (f *fakeDocker) ConfigList(_ context.Context, options ...docker.ConfigListOption) (*client.ConfigListResult, error) {
	opts := client.ConfigListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.ConfigListResult{}
	for _, config := range f.configs {
		if matches(config.Spec.Labels, opts.Filters) {
			out.Items = append(out.Items, config)
		}
	}
	return out, nil
}

func (f *fakeDocker) ConfigCreate(
	_ context.Context, name string, data []byte, options ...docker.ConfigCreateOption,
) (*client.ConfigCreateResult, error) {
	for _, config := range f.configs {
		if config.Spec.Name == name {
			return nil, hperrors.Wrap(hperrors.ErrInfraAlreadyExists)
		}
	}
	opts := client.ConfigCreateOptions{}
	opts.Spec.Name, opts.Spec.Data = name, data
	for _, opt := range options {
		opt(&opts)
	}
	id := f.id("config")
	f.configs[id] = swarm.Config{ID: id, Spec: opts.Spec}
	f.created = append(f.created, name)
	return &client.ConfigCreateResult{ID: id}, nil
}

func (f *fakeDocker) ConfigRemove(
	_ context.Context, id string, _ ...docker.ConfigRemoveOption,
) (*client.ConfigRemoveResult, error) {
	if f.inUse[id] > 0 {
		f.inUse[id]--
		return nil, hperrors.Wrap(hperrors.ErrInfraConflict)
	}
	delete(f.configs, id)
	return &client.ConfigRemoveResult{}, nil
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, _ string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	return &client.ServiceInspectResult{Service: *f.copyService()}, nil
}

func (f *fakeDocker) ServiceUpdateFunc(
	_ context.Context, _ string, _ *swarm.Service, fn func(int, *swarm.Service) (bool, error),
	_ int, _ time.Duration, _ ...docker.ServiceUpdateOption,
) error {
	svc := f.copyService()
	changed, err := fn(0, svc)
	if err != nil {
		return err
	}
	if changed {
		f.service = svc
		f.updates++
	}
	return nil
}

// copyService is the service as a fresh inspect returns it.
func (f *fakeDocker) copyService() *swarm.Service {
	svc := *f.service
	contSpec := *svc.Spec.TaskTemplate.ContainerSpec
	contSpec.Secrets = append([]*swarm.SecretReference(nil), contSpec.Secrets...)
	contSpec.Configs = append([]*swarm.ConfigReference(nil), contSpec.Configs...)
	svc.Spec.TaskTemplate.ContainerSpec = &contSpec
	return &svc
}
```

Confirm that the option type names match `services/docker`: `SecretRemoveOption`, `ConfigRemoveOption`, `ServiceInspectOption` and `ServiceUpdateOption` (`grep -n "^type .*Option func" services/docker/*.go`). Use whatever names are there.

- [ ] **Step 2: Write the failing tests.** `apply_test.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	sms "github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// engine is a fixture over a fake Docker whose service already has an ordinary
// secret, db_password.
func engine(t *testing.T, entries []*entity.Setting, sources ...*entity.Setting) (*service, *fakeDocker) {
	t.Helper()
	svc := fixture(t, entries, sources...)
	fake := newFakeDocker(swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Secrets: []*swarm.SecretReference{{SecretID: "ordinary_1", SecretName: "shop_prod_api_db_password",
			File: &swarm.SecretReferenceFileTarget{Name: "db_password", UID: "0", GID: "0", Mode: 0o444}}},
	}}})
	svc.dockerManager = fake
	svc.logger = logging.GlobalLogger()
	return svc, fake
}

func activeCertEntry(t *testing.T) *entity.Setting {
	return entry(t, "tls-cert", base.SettingStatusActive, certFiles("cert_1"))
}

func mountedTargets(spec *swarm.ServiceSpec) (secrets, configs []string) {
	for _, ref := range spec.TaskTemplate.ContainerSpec.Secrets {
		secrets = append(secrets, ref.File.Name)
	}
	for _, ref := range spec.TaskTemplate.ContainerSpec.Configs {
		configs = append(configs, ref.File.Name)
	}
	return secrets, configs
}

func TestApplyCreatesLabeledObjectsAndReferencesThem(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))
	spec := fake.copyService().Spec

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	secrets, configs := mountedTargets(&spec)
	assert.Equal(t, []string{"db_password", "/etc/app/tls/key.pem"}, secrets, "the ordinary secret stays")
	assert.Equal(t, []string{"/etc/app/tls/cert.pem"}, configs)
	key := spec.TaskTemplate.ContainerSpec.Secrets[1]
	assert.Equal(t, "1000", key.File.UID)
	assert.EqualValues(t, 0o400, key.File.Mode)
	stored := fake.secrets[key.SecretID]
	assert.Equal(t, []byte("KEY"), stored.Spec.Data)
	assert.Equal(t, sms.Labels("app_1", "tls-cert", "privateKey"), stored.Spec.Labels)
	assert.Regexp(t, `^shop_prod_api_mount_tls-cert_privatekey_[0-9a-f]{8}$`, key.SecretName)
}

func TestTheSameDataTwiceCreatesNothingNew(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))

	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))
	created := len(fake.created)
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.Len(t, fake.created, created)
	assert.Equal(t, 1, fake.updates, "nothing changed, so the service is not updated, nor restarted")
}

// A certificate and its key are replaced in one update, and what they replace
// is removed once nothing references it.
func TestARenewedCertificateIsSwappedInOneUpdateAndTheOldOneRemoved(t *testing.T) {
	source := certSource(t, "cert_1", "CERT", "KEY")
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, source)
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.NoError(t, source.SetData(&entity.SSLCert{Certificate: "CERT2", PrivateKey: entity.NewEncryptedField("KEY2")}))
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.Equal(t, 2, fake.updates)
	assert.Len(t, fake.secrets, 1, "the old key is gone; the ordinary secret was never Docker's to list here")
	assert.Len(t, fake.configs, 1)
	for _, secret := range fake.secrets {
		assert.Equal(t, []byte("KEY2"), secret.Spec.Data)
	}
}

func TestAnOrdinarySecretKeepsItsTarget(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "privateKey", Path: "/run/secrets/db_password"}},
	})}, certSource(t, "cert_1", "CERT", "KEY"))
	spec := fake.copyService().Spec

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	assert.Len(t, spec.TaskTemplate.ContainerSpec.Secrets, 1)
	assert.Equal(t, "ordinary_1", spec.TaskTemplate.ContainerSpec.Secrets[0].SecretID)
	assert.Empty(t, fake.created)
}

// A disabled entry's files leave the service, and their objects leave Docker.
func TestAnEntryThatCannotBeUsedIsTakenOut(t *testing.T) {
	e := activeCertEntry(t)
	svc, fake := engine(t, []*entity.Setting{e}, certSource(t, "cert_1", "CERT", "KEY"))
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	e.Status = base.SettingStatusDisabled
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	secrets, configs := mountedTargets(&fake.service.Spec)
	assert.Equal(t, []string{"db_password"}, secrets)
	assert.Empty(t, configs)
	assert.Empty(t, fake.secrets)
	assert.Empty(t, fake.configs)
}

// A preview or a clone starts from another app's spec; the files that app
// mounts are its own, and this app resolves its own.
func TestReferencesToAnotherAppsMountedObjectsAreDropped(t *testing.T) {
	svc, fake := engine(t, nil)
	fake.secrets["parent_key"] = swarm.Secret{ID: "parent_key", Spec: swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: "parent_key", Labels: sms.Labels("app_parent", "tls-cert", "privateKey")}}}
	spec := fake.copyService().Spec
	spec.TaskTemplate.ContainerSpec.Secrets = append(spec.TaskTemplate.ContainerSpec.Secrets,
		&swarm.SecretReference{SecretID: "parent_key", SecretName: "parent_key",
			File: &swarm.SecretReferenceFileTarget{Name: "/etc/app/tls/key.pem"}})

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	secrets, _ := mountedTargets(&spec)
	assert.Equal(t, []string{"db_password"}, secrets)
	assert.NoError(t, svc.Sweep(context.Background(), testApp))
	assert.Contains(t, fake.secrets, "parent_key", "another app's object is its own to sweep")
}

func TestSweepGivesUpOnWhatStaysInUseAndRemovesTheRest(t *testing.T) {
	svc, fake := engine(t, nil)
	for _, id := range []string{"stuck", "free"} {
		fake.secrets[id] = swarm.Secret{ID: id, Spec: swarm.SecretSpec{
			Annotations: swarm.Annotations{Name: id, Labels: sms.Labels("app_1", "a", "privateKey")}}}
	}
	fake.inUse["stuck"] = 100

	err := svc.Sweep(context.Background(), testApp)

	assert.Error(t, err)
	assert.Contains(t, fake.secrets, "stuck")
	assert.NotContains(t, fake.secrets, "free")
}

func TestASourceThatFailsToRenderChangesNothing(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))
	svc.loadSources = func(context.Context, database.IDB, *entity.App, []string) ([]*entity.Setting, error) {
		return nil, errors.New("no data key")
	}
	spec := fake.copyService().Spec
	before := fake.copyService().Spec

	assert.Error(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))
	assert.Equal(t, before, spec)
	assert.Empty(t, fake.created)
}

func TestRemoveAppRemovesOnlyItsOwnMountedObjects(t *testing.T) {
	svc, fake := engine(t, nil)
	fake.secrets["mine"] = swarm.Secret{ID: "mine", Spec: swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: "mine", Labels: sms.Labels("app_1", "a", "privateKey")}}}
	fake.configs["theirs"] = swarm.Config{ID: "theirs", Spec: swarm.ConfigSpec{
		Annotations: swarm.Annotations{Name: "theirs", Labels: sms.Labels("app_2", "a", "certificate")}}}

	assert.NoError(t, svc.RemoveApp(context.Background(), "app_1"))

	assert.Empty(t, fake.secrets)
	assert.Contains(t, fake.configs, "theirs")
}
```

- [ ] **Step 3: Run.** `go test ./hivepaas_app/service/settingmountservice/...` should FAIL to compile.

- [ ] **Step 4: Implement.** `apply.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"errors"
	"reflect"
	"slices"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const serviceUpdateRetryMax = 2

var _ settingmountservice.Service = (*service)(nil)

// mounted is what Docker holds of mounted settings, by id and by name.
type mounted struct {
	secrets, configs     map[string]string // id -> name
	secretIDs, configIDs map[string]string // name -> id
}

func (s *service) ApplyToService(
	ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec,
) error {
	contSpec := spec.TaskTemplate.ContainerSpec
	if contSpec == nil {
		return nil
	}
	files, err := s.Resolve(ctx, db, app)
	if err != nil {
		return hperrors.Wrap(err)
	}
	have, err := s.listMounted(ctx, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// What stays: every reference that is not a mounted setting's. Its targets
	// are taken, and a file claiming one is left out.
	secrets := slices.DeleteFunc(slices.Clone(contSpec.Secrets), func(ref *swarm.SecretReference) bool {
		_, ok := have.secrets[ref.SecretID]
		return ok
	})
	configs := slices.DeleteFunc(slices.Clone(contSpec.Configs), func(ref *swarm.ConfigReference) bool {
		_, ok := have.configs[ref.ConfigID]
		return ok
	})
	taken := map[string]bool{}
	for _, ref := range secrets {
		if ref.File != nil {
			taken[settingmountservice.SecretTarget(ref.File.Name)] = true
		}
	}
	for _, ref := range configs {
		if ref.File != nil {
			taken[settingmountservice.ConfigTarget(ref.File.Name)] = true
		}
	}

	for _, file := range files {
		if taken[file.Path] {
			continue
		}
		name := settingmountservice.ObjectName(app.GlobalKey, file.Entry, file.Part, file.Rotation)
		if file.Sensitive {
			id, err := s.ensureSecret(ctx, have, app, file, name)
			if err != nil {
				return hperrors.Wrap(err)
			}
			secrets = append(secrets, &swarm.SecretReference{SecretID: id, SecretName: name,
				File: &swarm.SecretReferenceFileTarget{Name: file.Path, UID: file.UID, GID: file.GID,
					Mode: file.Mode.ToFileMode()}})
			continue
		}
		id, err := s.ensureConfig(ctx, have, app, file, name)
		if err != nil {
			return hperrors.Wrap(err)
		}
		configs = append(configs, &swarm.ConfigReference{ConfigID: id, ConfigName: name,
			File: &swarm.ConfigReferenceFileTarget{Name: file.Path, UID: file.UID, GID: file.GID,
				Mode: file.Mode.ToFileMode()}})
	}
	contSpec.Secrets, contSpec.Configs = secrets, configs
	return nil
}

// ensureSecret is the id of the secret named name, created when Docker has
// none. A create that finds the name taken - a refresh running beside a
// deployment - reads it back rather than failing.
func (s *service) ensureSecret(
	ctx context.Context, have *mounted, app *entity.App, file *settingmountservice.File, name string,
) (string, error) {
	if id, ok := have.secretIDs[name]; ok {
		return id, nil
	}
	resp, err := s.dockerManager.SecretCreate(ctx, name, file.Data, func(opts *client.SecretCreateOptions) {
		opts.Spec.Labels = settingmountservice.Labels(app.ID, file.Entry, file.Part)
	})
	if err == nil {
		have.secretIDs[name] = resp.ID
		return resp.ID, nil
	}
	if !errors.Is(err, hperrors.ErrInfraConflict) && !errors.Is(err, hperrors.ErrInfraAlreadyExists) {
		return "", hperrors.Wrap(err)
	}
	again, listErr := s.listMounted(ctx, settingmountservice.LabelEntry)
	if listErr != nil {
		return "", hperrors.Wrap(listErr)
	}
	if id, ok := again.secretIDs[name]; ok {
		return id, nil
	}
	return "", hperrors.Wrap(err)
}
```

Write `ensureConfig` the same way over `ConfigCreate` and `configIDs`. `listMounted` lists both kinds with label filters, each `"key"` (present) or `"key=value"` (equal):

```go
func (s *service) listMounted(ctx context.Context, labels ...string) (*mounted, error) {
	have := &mounted{secrets: map[string]string{}, configs: map[string]string{},
		secretIDs: map[string]string{}, configIDs: map[string]string{}}
	secrets, err := s.dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		for _, label := range labels {
			docker.FilterAdd(&opts.Filters, "label", label)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, secret := range secrets.Items {
		have.secrets[secret.ID], have.secretIDs[secret.Spec.Name] = secret.Spec.Name, secret.ID
	}
	configs, err := s.dockerManager.ConfigList(ctx, func(opts *client.ConfigListOptions) {
		for _, label := range labels {
			docker.FilterAdd(&opts.Filters, "label", label)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, config := range configs.Items {
		have.configs[config.ID], have.configIDs[config.Spec.Name] = config.Spec.Name, config.ID
	}
	return have, nil
}
```

```go
// Refresh brings the app's service to its mounted files in one update, which
// restarts it only when a reference changed, and then sweeps.
func (s *service) Refresh(ctx context.Context, db database.IDB, app *entity.App) error {
	if app.ServiceID == "" {
		return nil
	}
	err := s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, svc *swarm.Service) (bool, error) {
			contSpec := svc.Spec.TaskTemplate.ContainerSpec
			if contSpec == nil {
				return false, nil
			}
			secrets, configs := slices.Clone(contSpec.Secrets), slices.Clone(contSpec.Configs)
			if err := s.ApplyToService(ctx, db, app, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}
			changed := !reflect.DeepEqual(secrets, contSpec.Secrets) || !reflect.DeepEqual(configs, contSpec.Configs)
			return changed, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = s.Sweep(ctx, app); err != nil {
		s.logger.Warnf("setting mounts of app %s: %v", app.ID, err)
	}
	return nil
}
```

`reflect.DeepEqual` on slices of pointers compares the pointees. A spec that has lost nothing and gained nothing is equal even though the slices are new, and that is what the test "the same data twice" pins.

`sweep.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const removalRetryMax = 2

// Sweep removes the app's mounted objects that its service no longer
// references. Docker refuses to remove one a service still names; that one is
// retried, then left to the next sweep - a deployment is not failed for it.
func (s *service) Sweep(ctx context.Context, app *entity.App) error {
	have, err := s.listMounted(ctx, settingmountservice.LabelAppID+"="+app.ID, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}
	referenced := map[string]bool{}
	if app.ServiceID != "" {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if err == nil && inspect.Service.Spec.TaskTemplate.ContainerSpec != nil {
			for _, ref := range inspect.Service.Spec.TaskTemplate.ContainerSpec.Secrets {
				referenced[ref.SecretID] = true
			}
			for _, ref := range inspect.Service.Spec.TaskTemplate.ContainerSpec.Configs {
				referenced[ref.ConfigID] = true
			}
		}
	}
	return s.remove(ctx, have, referenced)
}

// RemoveApp removes every mounted object of an app, once its service is gone.
func (s *service) RemoveApp(ctx context.Context, appID string) error {
	have, err := s.listMounted(ctx, settingmountservice.LabelAppID+"="+appID, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return s.remove(ctx, have, nil)
}

func (s *service) remove(ctx context.Context, have *mounted, keep map[string]bool) (errs error) {
	for id := range have.secrets {
		if keep[id] {
			continue
		}
		errs = errors.Join(errs, s.retryRemoval(ctx, func() error {
			_, err := s.dockerManager.SecretRemove(ctx, id)
			return err
		}))
	}
	for id := range have.configs {
		if keep[id] {
			continue
		}
		errs = errors.Join(errs, s.retryRemoval(ctx, func() error {
			_, err := s.dockerManager.ConfigRemove(ctx, id)
			return err
		}))
	}
	return errs
}

func (s *service) retryRemoval(ctx context.Context, remove func() error) error {
	err := gofn.ExecRetryCtx(ctx, func() error {
		if err := remove(); err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return err
		}
		return nil
	}, removalRetryMax, s.removalRetryDelay)
	return hperrors.Wrap(err)
}
```

Check the signature of `gofn.ExecRetryCtx` in the vendor directory. `clustersecretserviceimpl/secret_remove.go` calls it as `(ctx, fn, retryMax, retryDelay, opts...)`. Match that.

- [ ] **Step 5: Run** `go test ./hivepaas_app/service/settingmountservice/...`. It should PASS. Then run `golangci-lint run ./hivepaas_app/service/settingmountservice/...`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/settingmountservice/
git commit -m "feat(settingmounts): put mounted files on a service in one update, and sweep what it replaced

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: The refresh task

**Files:**
- Modify: `hivepaas_app/base/task.go`:
  - `TaskTypeSettingMountRefresh TaskType = "task:setting-mount-refresh"` with a comment;
  - add it to `AllTaskTypes`.
- Create: `hivepaas_app/entity/task_setting_mount_refresh.go`
- Create: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/refresh_task.go`
- Test: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/refresh_task_test.go`
- Create: `hivepaas_app/tasks/tasksettingmountrefresh/executor.go`
- Test: `hivepaas_app/tasks/tasksettingmountrefresh/executor_test.go`
- Modify: `hivepaas_app/tasks/queue/queueimpl/task_timeout.go`:
  - `timeoutSettingMountRefresh = 15 * time.Minute`;
  - the map entry.
- Modify: `hivepaas_app/tasks/initializer/initializer.go`: `_ *tasksettingmountrefresh.Executor`.
- Modify: `hivepaas_app/registry/provides.go`:
  - `settingmountserviceimpl.New` beside `dockerapiserviceimpl.New`;
  - `tasksettingmountrefresh.NewExecutor` beside `tasksslobtain.NewExecutor`;
  - imports.

**Interfaces:**
- Consumes: `taskRepo.Insert(ctx, db, task)`, `taskQueue.ScheduleTask(ctx, tasks...)`, `resLinkRepo.List`, `settingRepo.ListByIDs`, `appRepo.GetByID(ctx, db, "", id)`.
- Produces:
  - `entity.TaskSettingMountRefreshArgs{AppIDs []string}` and `entity.TaskSettingMountRefreshOutput{Applied int; Failed map[string]string}`;
  - `(*Task).ArgsAsSettingMountRefresh()` and `(*Task).OutputAsSettingMountRefresh()`;
  - `RecordRefresh`, `Schedule`.

- [ ] **Step 1: Entity.** `entity/task_setting_mount_refresh.go`:

```go
package entity

// TaskSettingMountRefreshArgs are the apps whose mounted settings to bring up to
// date. Only their ids: what they mount is read when the task runs, which may be
// minutes later after a retry.
type TaskSettingMountRefreshArgs struct {
	AppIDs []string `json:"appIds"`
}

type TaskSettingMountRefreshOutput struct {
	Applied int               `json:"applied"`
	Failed  map[string]string `json:"failed,omitempty"`
}

func (t *Task) ArgsAsSettingMountRefresh() (*TaskSettingMountRefreshArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSettingMountRefreshArgs { return &TaskSettingMountRefreshArgs{} })
}

func (t *Task) OutputAsSettingMountRefresh() (*TaskSettingMountRefreshOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSettingMountRefreshOutput { return &TaskSettingMountRefreshOutput{} })
}
```

`base/task.go`, after `TaskTypeSSLObtain`:

```go
	// TaskTypeSettingMountRefresh brings the files apps mount from settings up to
	// date after one of those settings changed. It is recorded in the
	// transaction that changes the setting. See service/settingmountservice.
	TaskTypeSettingMountRefresh TaskType = "task:setting-mount-refresh"
```

- [ ] **Step 2: Write the failing tests.** `refresh_task_test.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type fakeTaskRepo struct {
	repository.TaskRepo
	inserted []*entity.Task
}

func (f *fakeTaskRepo) Insert(_ context.Context, _ database.IDB, task *entity.Task, _ ...bunex.InsertQueryOption) error {
	f.inserted = append(f.inserted, task)
	return nil
}

func recorder(t *testing.T, readers map[string][]string) (*service, *fakeTaskRepo) {
	t.Helper()
	tasks := &fakeTaskRepo{}
	svc := &service{taskRepo: tasks}
	svc.loadReaders = func(_ context.Context, _ database.IDB, ids []string) ([]string, error) {
		var apps []string
		for _, id := range ids {
			apps = append(apps, readers[id]...)
		}
		return apps, nil
	}
	return svc, tasks
}

func TestRecordRefreshNamesTheAppsThatReadTheSettings(t *testing.T) {
	svc, tasks := recorder(t, map[string][]string{"cert_1": {"app_2", "app_1"}})

	task, err := svc.RecordRefresh(context.Background(), nil,
		&entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert},
		&entity.Setting{ID: "entry_1", Type: base.SettingTypeAppSettingMount, ObjectID: "app_1"},
		&entity.Setting{ID: "secret_1", Type: base.SettingTypeSecret, ObjectID: "app_9"})

	assert.NoError(t, err)
	if assert.NotNil(t, task) && assert.Len(t, tasks.inserted, 1) {
		assert.Equal(t, base.TaskTypeSettingMountRefresh, task.Type)
		assert.Equal(t, base.TaskStatusNotStarted, task.Status)
		args, err := task.ArgsAsSettingMountRefresh()
		assert.NoError(t, err)
		assert.Equal(t, []string{"app_1", "app_2"}, args.AppIDs, "sorted, once each")
	}
}

func TestRecordRefreshRecordsNothingWhenNoAppReads(t *testing.T) {
	svc, tasks := recorder(t, nil)

	task, err := svc.RecordRefresh(context.Background(), nil,
		&entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert},
		&entity.Setting{ID: "secret_1", Type: base.SettingTypeSecret})

	assert.NoError(t, err)
	assert.Nil(t, task)
	assert.Empty(t, tasks.inserted)
}
```

`tasks/tasksettingmountrefresh/executor_test.go`:

```go
package tasksettingmountrefresh

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeApps struct{ repository.AppRepo }

func (fakeApps) GetByID(_ context.Context, _ database.IDB, _, id string, _ ...bunex.SelectQueryOption) (*entity.App, error) {
	if id == "app_gone" {
		return nil, hperrors.NewNotFound("App")
	}
	return &entity.App{ID: id}, nil
}

type fakeMounts struct {
	settingmountservice.Service
	refreshed []string
}

func (f *fakeMounts) Refresh(_ context.Context, _ database.IDB, app *entity.App) error {
	f.refreshed = append(f.refreshed, app.ID)
	if app.ID == "app_broken" {
		return errors.New("docker said no")
	}
	return nil
}

// Every app is tried: one that fails is written down and retried with the task,
// one deleted since is skipped.
func TestTheTaskRefreshesEveryAppAndReportsThoseThatFailed(t *testing.T) {
	mounts := &fakeMounts{}
	e := &Executor{appRepo: fakeApps{}, settingMountService: mounts, logger: logging.GlobalLogger()}
	task := &entity.Task{Type: base.TaskTypeSettingMountRefresh}
	assert.NoError(t, task.SetArgs(&entity.TaskSettingMountRefreshArgs{
		AppIDs: []string{"app_1", "app_broken", "app_gone", "app_2"}}))

	err := e.execute(context.Background(), nil, &queue.TaskExecData{Task: task})

	assert.Error(t, err)
	assert.Equal(t, []string{"app_1", "app_broken", "app_2"}, mounts.refreshed)
	out, _ := task.OutputAsSettingMountRefresh()
	assert.Equal(t, 2, out.Applied)
	assert.Contains(t, out.Failed, "app_broken")
}
```

Check `appRepo.GetByID`'s exact signature (`grep -n "GetByID(ctx" hivepaas_app/repository/app_repo.go`) and `hperrors.NewNotFound`, and match the fake to them.

- [ ] **Step 3: Run** both packages. They should FAIL to compile.

- [ ] **Step 4: Implement.** `refresh_task.go`:

```go
package settingmountserviceimpl

import (
	"context"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const (
	refreshMaxRetry   = 3
	refreshRetryDelay = timeutil.Duration(time.Minute)
)

func (s *service) RecordRefresh(
	ctx context.Context, db database.IDB, settings ...*entity.Setting,
) (*entity.Task, error) {
	var appIDs, sourceIDs []string
	for _, setting := range settings {
		switch {
		case setting == nil:
		case setting.Type == base.SettingTypeAppSettingMount:
			appIDs = append(appIDs, setting.ObjectID)
		case settingmountservice.IsSourceType(setting.Type):
			sourceIDs = append(sourceIDs, setting.ID)
		}
	}
	if len(sourceIDs) > 0 {
		readers, err := s.loadReaders(ctx, db, sourceIDs)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		appIDs = append(appIDs, readers...)
	}
	appIDs = gofn.ToSet(gofn.ToSliceSkippingZero(appIDs...))
	if len(appIDs) == 0 {
		return nil, nil
	}
	slices.Sort(appIDs)

	now := timeutil.NowUTC()
	task := &entity.Task{
		ID:     gofn.Must(ulid.NewStringULID()),
		Scope:  base.ObjectScopeGlobal,
		Type:   base.TaskTypeSettingMountRefresh,
		Status: base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityDefault,
			MaxRetry:   refreshMaxRetry,
			RetryDelay: refreshRetryDelay,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := task.SetArgs(&entity.TaskSettingMountRefreshArgs{AppIDs: appIDs}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := s.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

func (s *service) Schedule(ctx context.Context, tasks ...*entity.Task) {
	tasks = gofn.ToSliceSkippingNil(tasks...)
	if len(tasks) == 0 {
		return
	}
	if err := s.taskQueue.ScheduleTask(ctx, tasks...); err != nil {
		s.logger.Warnf("setting mount refresh left to the queue's scan: %v", err)
	}
}

// loadReadersFromRepo is the apps whose entries mount one of sourceIDs, through
// the links every setting reference writes.
func (s *service) loadReadersFromRepo(ctx context.Context, db database.IDB, sourceIDs []string) ([]string, error) {
	links, _, err := s.resLinkRepo.List(ctx, db, nil,
		bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("res_link.dst_id IN (?)", bunex.List(sourceIDs)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(links) == 0 {
		return nil, nil
	}
	entryIDs := make([]string, 0, len(links))
	for _, link := range links {
		entryIDs = append(entryIDs, link.SrcID)
	}
	entries, err := s.settingRepo.ListByIDs(ctx, db, nil, entryIDs, false,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		appIDs = append(appIDs, e.ObjectID)
	}
	return appIDs, nil
}
```

Check that `bunex.List` exists; `repository/setting_repo.go` uses it. Also check `gofn.ToSliceSkippingNil`; without it, filter with `gofn.Filter`.

`tasks/tasksettingmountrefresh/executor.go`:

```go
// Package tasksettingmountrefresh brings the files apps mount from settings up
// to date after one of those settings changed: a certificate renewed, a password
// replaced, an entry disabled.
package tasksettingmountrefresh

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	appRepo             repository.AppRepo
	settingMountService settingmountservice.Service
	logger              logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRepo repository.AppRepo,
	settingMountService settingmountservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{appRepo: appRepo, settingMountService: settingMountService, logger: logger}
	taskQueue.RegisterExecutor(base.TaskTypeSettingMountRefresh, e.execute)
	return e
}

func (e *Executor) execute(ctx context.Context, db database.Tx, task *queue.TaskExecData) error {
	args, err := task.Task.ArgsAsSettingMountRefresh()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("setting mount refresh task has no args")
	}

	output := &entity.TaskSettingMountRefreshOutput{}
	for _, appID := range args.AppIDs {
		app, err := e.appRepo.GetByID(ctx, db, "", appID)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				continue // deleted since; its objects went with it
			}
			output.fail(appID, err)
			continue
		}
		if err = e.settingMountService.Refresh(ctx, db, app); err != nil {
			e.logger.Errorf("failed to refresh the setting mounts of app %s: %v", appID, err)
			output.fail(appID, err)
			continue
		}
		output.Applied++
	}
	task.Task.MustSetOutput(output)

	// An app that failed is retried with the task; the others cost one inspect
	// each and no update, as nothing of theirs changed.
	if len(output.Failed) > 0 {
		return hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("setting mounts could not be refreshed on %d app(s)", len(output.Failed))
	}
	return nil
}
```

Put `fail` in the entity file:

```go
func (o *TaskSettingMountRefreshOutput) fail(...)
```

It cannot be unexported across packages, so write it in the executor as a local helper instead:

```go
func fail(out *entity.TaskSettingMountRefreshOutput, appID string, err error) {
	if out.Failed == nil {
		out.Failed = map[string]string{}
	}
	out.Failed[appID] = err.Error()
}
```

Replace `output.fail(appID, err)` with `fail(output, appID, err)`.

Then make the wiring changes listed under **Files**.

- [ ] **Step 5: Run** `go build ./... && go test ./hivepaas_app/service/settingmountservice/... ./hivepaas_app/tasks/... ./hivepaas_app/registry/...`. It should PASS. If the registry has a dependency-graph test, it builds the new graph and catches a cycle.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/base/task.go hivepaas_app/entity/task_setting_mount_refresh.go \
  hivepaas_app/service/settingmountservice/ hivepaas_app/tasks/tasksettingmountrefresh/ \
  hivepaas_app/tasks/queue/queueimpl/task_timeout.go hivepaas_app/tasks/initializer/initializer.go \
  hivepaas_app/registry/provides.go
git commit -m "feat(settingmounts): a refresh task for the apps that read a changed setting

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: What records the refresh task

**Files:**
- Modify: `hivepaas_app/service/settingeventservice/types.go`: `Tasks []*entity.Task` on `CreateEvent`, `UpdateEvent` and `DeleteEvent`.
- Modify: `hivepaas_app/service/settingeventservice/service.go`: `ScheduleTasks(ctx context.Context, tasks ...*entity.Task)`.
- Modify: `settingeventserviceimpl/service.go` (the `settingMountService` dependency), `event_on_create.go`, `event_on_update.go` (`OnUpdate` and `OnUpdateStatus`, wherever it is), `event_on_delete.go`.
- Create: `settingeventserviceimpl/mount_refresh.go`
- Test: `settingeventserviceimpl/mount_refresh_test.go`
- Modify: the seven use-case sites:
  - `usecase/settings/setting_create.go`;
  - `setting_update.go`;
  - `setting_update_unique.go`;
  - `setting_update_status.go`;
  - `setting_update_status_unique.go`;
  - `setting_delete.go`;
  - `setting_delete_unique.go`.
- Modify: `service/sslrenewalservice/sslrenewalserviceimpl/service.go` and `ssl_renewal.go`
- Modify: `tasks/tasksslobtain/executor.go`
- Modify: `service/specservice/specserviceimpl/service.go` (New), `import_write.go` (`write`)
- Test: `service/specservice/specserviceimpl/fakes_apply_test.go`, `export_test.go` (the fixture's `New` call), `import_apply_test.go`

**Interfaces:**
- Consumes: `RecordRefresh`, `Schedule` (Task 6).
- Produces: `settingeventservice.Service.ScheduleTasks`.

- [ ] **Step 1: Write the failing tests.** `settingeventserviceimpl/mount_refresh_test.go`:

```go
package settingeventserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

type fakeMounts struct {
	settingmountservice.Service
	recorded []*entity.Setting
}

func (f *fakeMounts) RecordRefresh(
	_ context.Context, _ database.IDB, settings ...*entity.Setting,
) (*entity.Task, error) {
	f.recorded = append(f.recorded, settings...)
	return &entity.Task{ID: "task_1"}, nil
}

// Every path the settings screens write through records a refresh for what it
// wrote, and hands the task back to be scheduled after commit.
func TestSettingEventsRecordAMountRefresh(t *testing.T) {
	mounts := &fakeMounts{}
	s := &service{settingMountService: mounts}
	cert := &entity.Setting{ID: "cert_1", Type: base.SettingTypeSSLCert}
	ctx := context.Background()

	created := &settingeventservice.CreateEvent{Setting: cert}
	updated := &settingeventservice.UpdateEvent{Setting: cert}
	status := &settingeventservice.UpdateEvent{Setting: cert}
	deleted := &settingeventservice.DeleteEvent{Setting: cert}
	assert.NoError(t, s.OnCreate(ctx, nil, created))
	assert.NoError(t, s.OnUpdate(ctx, nil, updated))
	assert.NoError(t, s.OnUpdateStatus(ctx, nil, status))
	assert.NoError(t, s.OnDelete(ctx, nil, deleted))

	assert.Len(t, mounts.recorded, 4)
	for _, tasks := range [][]*entity.Task{created.Tasks, updated.Tasks, status.Tasks, deleted.Tasks} {
		assert.Len(t, tasks, 1)
	}
}
```

Read `event_on_create.go`, `event_on_delete.go` and `event_on_update_meta.go` first. If `OnCreate` or `OnDelete` publishes to `systemEventBus` unconditionally, set `systemEventBus` to a stub in the test, or keep the setting type outside the branches that publish (SSL cert is outside them today).

In `specserviceimpl/import_apply_test.go`:

```go
// Import records a refresh for what it wrote, in its own transaction; it is
// scheduled with the rest of phase one's tasks, after the commit.
func TestApplyRecordsAMountRefreshForWhatItWrote(t *testing.T) {
	svc, bundle := planFixture(t)
	mounts := svc.settingMountService.(*fakeSettingMounts)

	resp := apply(t, svc, bundle, applyReq(t, svc, bundle))

	assert.NotEmpty(t, mounts.recorded)
	assert.True(t, slices.ContainsFunc(resp.Tasks, func(task *entity.Task) bool {
		return task.Type == base.TaskTypeSettingMountRefresh
	}))
}
```

In `fakes_apply_test.go`:

```go
type fakeSettingMounts struct {
	settingmountservice.Service
	recorded []*entity.Setting
}

func (f *fakeSettingMounts) RecordRefresh(
	_ context.Context, _ database.IDB, settings ...*entity.Setting,
) (*entity.Task, error) {
	f.recorded = append(f.recorded, settings...)
	if len(settings) == 0 {
		return nil, nil
	}
	return &entity.Task{ID: "task_mount_refresh", Type: base.TaskTypeSettingMountRefresh}, nil
}
```

In `export_test.go`'s `New(...)` call, add `&fakeSettingMounts{},` after `&fakeNetworkService{},`.

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/settingeventservice/... ./hivepaas_app/service/specservice/...`. It should FAIL to compile.

- [ ] **Step 3: Implement the event service.** `mount_refresh.go`:

```go
package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// recordMountRefresh records, in the event's transaction, a refresh of the apps
// that mount the setting - or of the entry's own app - into tasks.
func (s *service) recordMountRefresh(
	ctx context.Context, db database.IDB, tasks *[]*entity.Task, setting *entity.Setting,
) error {
	task, err := s.settingMountService.RecordRefresh(ctx, db, setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if task != nil {
		*tasks = append(*tasks, task)
	}
	return nil
}

func (s *service) ScheduleTasks(ctx context.Context, tasks ...*entity.Task) {
	s.settingMountService.Schedule(ctx, tasks...)
}
```

`OnUpdate` then becomes:

```go
func (s *service) OnUpdate(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.UpdateEvent,
) (err error) {
	// Reload periodic jobs in workers as the update may relate
	if event.Setting.IsTypeIn(base.SettingTypePeriodicJob, base.SettingTypeIMService, base.SettingTypeEmail) {
		_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	}

	return s.recordMountRefresh(ctx, db, &event.Tasks, event.Setting)
}
```

Make the same final call in `OnUpdateStatus`, `OnCreate` and `OnDelete`, with each function's existing body kept above it. Add `settingMountService settingmountservice.Service` to `New` and to the struct.

- [ ] **Step 4: Schedule after commit at the seven use-case sites.** Each builds its event inside the transaction's closure. Hoist the event variable out, and schedule once the transaction returns without error. For `setting_update.go`:

```go
	var event *settingeventservice.UpdateEvent
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		// ... unchanged ...

		// Fire update event
		event = &settingeventservice.UpdateEvent{
			Setting:    persistingData.Setting,
			OldSetting: data.Setting,
		}
		err = uc.SettingEventService.OnUpdate(ctx, db, event)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The tasks the event recorded exist once the transaction has committed.
	uc.SettingEventService.ScheduleTasks(ctx, event.Tasks...)
```

Apply the same shape to the other six:
- `CreateEvent` in `setting_create.go`;
- `UpdateEvent` with `OnUpdateStatus` in the two status files;
- `UpdateEvent` in `setting_update_unique.go`;
- `DeleteEvent` in the two delete files.

A retried transaction rebuilds the event, so only the committed attempt's tasks are scheduled.

- [ ] **Step 5: SSL renewal.** Add `settingMountService settingmountservice.Service` to `sslrenewalserviceimpl.New` and to the struct. In `sslSaveUpdatedSettings`, declare `var refresh *entity.Task` before `transaction.Execute`. After `UpsertMulti` succeeds inside the closure:

```go
		// The apps that mount a renewed certificate get it; renewing is not
		// granting, so nothing is asked.
		refresh, err = s.settingMountService.RecordRefresh(ctx, db, persistingSettings...)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
```

After `WriteCertFiles`, call `s.settingMountService.Schedule(ctx, refresh)`. `Schedule` skips nil tasks.

- [ ] **Step 6: SSL obtain.** Add `settingMountService settingmountservice.Service` to `tasksslobtain.NewExecutor` and to `Executor`. In `execute`, after `saveObtained` succeeds:

```go
	refresh, err := e.settingMountService.RecordRefresh(ctx, db, setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	task.OnPostTx(func() { e.settingMountService.Schedule(context.WithoutCancel(ctx), refresh) })
```

`task` here is the `*queue.TaskExecData` parameter. The queue calls `OnPostTx` callbacks once the task's transaction has committed.

- [ ] **Step 7: Import.**
  - Add `settingMountService settingmountservice.Service` to `specserviceimpl.New`, after `networkService`, and to the struct.
  - In `writer.write`, after `provisionApps` and before `setOutcomes`:

```go
	// Written settings someone mounts are refreshed once the import commits,
	// scheduled with the other tasks phase one made.
	refresh, err := w.p.s.settingMountService.RecordRefresh(ctx, w.p.db, w.written...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if refresh != nil {
		w.tasks = append(w.tasks, refresh)
	}
```

`usecase/specuc/import_apply.go` already schedules `applied.Tasks` after `AfterCommit`.

- [ ] **Step 8: Run** `go build ./... && go test ./hivepaas_app/service/settingeventservice/... ./hivepaas_app/service/specservice/... ./hivepaas_app/usecase/settings/... ./hivepaas_app/tasks/... ./hivepaas_app/service/sslrenewalservice/...`. It should PASS. Renewal and obtain have no unit fixture for their transactions. Cover them by the build here and by the Linux check in Task 9.

- [ ] **Step 9: Commit**

```bash
git add hivepaas_app/service/settingeventservice/ hivepaas_app/usecase/settings/setting_create.go \
  hivepaas_app/usecase/settings/setting_update.go hivepaas_app/usecase/settings/setting_update_unique.go \
  hivepaas_app/usecase/settings/setting_update_status.go hivepaas_app/usecase/settings/setting_update_status_unique.go \
  hivepaas_app/usecase/settings/setting_delete.go hivepaas_app/usecase/settings/setting_delete_unique.go \
  hivepaas_app/service/sslrenewalservice/sslrenewalserviceimpl/ hivepaas_app/tasks/tasksslobtain/executor.go \
  hivepaas_app/service/specservice/specserviceimpl/
git commit -m "feat(settingmounts): every write of a mounted setting records a refresh

The settings screens, SSL renewal, SSL obtain and import record it in their
own transaction and schedule it once that commits.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 8: Every deployment puts it right; app deletion cleans up

**Files:**
- Modify: `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/service.go` (the dependency)
- Modify: `image_deploy_apply_svc.go`, `repo_deploy_apply_svc.go`
- Modify: `hivepaas_app/service/appservice/appserviceimpl/service.go` (the dependency), `deletion.go`

**Interfaces:**
- Consumes: `ApplyToService`, `Sweep`, `RemoveApp`.

- [ ] **Step 1: Deployments.** In both `*StepServiceApply` functions, directly after the Docker API call inside the update function:

```go
			// Mounted settings too: a refresh that failed is put right here.
			if err := s.settingMountService.ApplyToService(ctx, db, data.App, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}
```

After `ServiceUpdateFunc` returns without error:

```go
	// What the update replaced goes once nothing references it; what is still
	// held is left to the next sweep rather than failing the deployment.
	_ = s.settingMountService.Sweep(ctx, data.App)
```

Add `settingMountService settingmountservice.Service` to `New` (after `settingService`) and to the struct.

- [ ] **Step 2: Deletion.** In `deleteAppInDocker`, after `_ = s.dockerAPIService.RemoveApp(ctx, app.ID)`:

```go
	_ = s.settingMountService.RemoveApp(ctx, app.ID)
```

Add the dependency to `appserviceimpl.New` and to the struct. First check for a cycle: `settingmountserviceimpl` must not import `appservice`. It does not. Also check that nothing `settingmountserviceimpl.New` takes is built from `appservice` (`go build ./...` and starting the backend both catch it).

- [ ] **Step 3: Run** `go build ./... && go test ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/appservice/...`. It should PASS.

- [ ] **Step 4: Commit**

```bash
git add hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/ hivepaas_app/service/appservice/appserviceimpl/
git commit -m "feat(settingmounts): deployments apply mounted settings, and app deletion removes them

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 9: Against a real daemon, the gates, and the merge

**Files:**
- Test: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/real_daemon_test.go`
- Modify: `docs/superpowers/specs/2026-09-25-setting-mounts-design.md`:
  - §2: the rotation key is keyed, an HMAC with the app secret;
  - §5: import records in phase one and schedules after commit;
  - §12: secrets outside `/run/secrets` verified on Docker 29.8.

- [ ] **Step 1: Write the test.** It runs only with `HP_TEST_DOCKER=1` and a swarm, and is skipped otherwise:

```go
package settingmountserviceimpl

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TestAgainstARealDaemon mounts a certificate and its key outside /run/secrets,
// renews them, and checks the container reads the new pair after one update and
// that the old objects are gone. It needs a swarm manager and alpine:3.
func TestAgainstARealDaemon(t *testing.T) {
	if os.Getenv("HP_TEST_DOCKER") == "" {
		t.Skip("set HP_TEST_DOCKER=1 to run against this machine's Docker daemon")
	}
	ctx := context.Background()
	dockerManager, err := docker.New()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	id, _ := ulid.NewStringULID()
	app := &entity.App{ID: "itest-" + id, GlobalKey: "itest_" + id}

	source := certSource(t, "cert_1", "CERT-1", "KEY-1")
	svc := fixture(t, []*entity.Setting{entry(t, "tls-cert", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"}, Files: []*entity.AppSettingMountFile{
			{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
			{Part: "privateKey", Path: "/etc/app/tls/key.pem", Mode: 0o400},
		}})}, source)
	svc.dockerManager, svc.logger = dockerManager, logging.GlobalLogger()
	svc.removalRetryDelay = time.Second

	spec := swarm.ServiceSpec{Annotations: swarm.Annotations{Name: app.GlobalKey},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
			Image: "alpine:3", Command: []string{"sleep", "3600"}}}}
	assert.NoError(t, svc.ApplyToService(ctx, nil, app, &spec))
	created, err := dockerManager.ServiceCreate(ctx, &spec)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	app.ServiceID = created.ID
	t.Cleanup(func() {
		_, _ = dockerManager.ServiceRemove(context.Background(), app.ServiceID)
		time.Sleep(3 * time.Second)
		_ = svc.RemoveApp(context.Background(), app.ID)
	})

	readKey := func() string {
		t.Helper()
		_, err := dockerManager.ServiceWaitUntilRunning(ctx, app.ServiceID, true, 2*time.Second, time.Second)
		assert.NoError(t, err)
		containers, err := dockerManager.ServiceContainerList(ctx, app.ServiceID)
		if !assert.NoError(t, err) || !assert.NotEmpty(t, containers.Items) {
			t.FailNow()
		}
		out, err := dockerManager.ContainerCopyFrom(ctx, containers.Items[0].ID, "/etc/app/tls/key.pem")
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		defer out.Content.Close()
		tr := tar.NewReader(out.Content)
		_, err = tr.Next()
		assert.NoError(t, err)
		data, _ := io.ReadAll(tr)
		return string(data)
	}
	assert.Equal(t, "KEY-1", readKey())

	assert.NoError(t, source.SetData(&entity.SSLCert{Certificate: "CERT-2", PrivateKey: entity.NewEncryptedField("KEY-2")}))
	assert.NoError(t, svc.Refresh(ctx, nil, app))
	time.Sleep(5 * time.Second) // the old task stops, and with it the last hold on the old secret
	assert.NoError(t, svc.Sweep(ctx, app))
	assert.Equal(t, "KEY-2", readKey())

	left, err := dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		docker.FilterAdd(&opts.Filters, "label", "hivepaas.app.id="+app.ID)
	})
	assert.NoError(t, err)
	assert.Len(t, left.Items, 1, "only the new key is left")
}
```

If `ServiceWaitUntilRunning` returns before the new task replaces the old one, poll `ServiceContainerList` until the container's `Created` time is after the refresh. If `ContainerCopyFrom` cannot read from a secret's tmpfs, read the file with `ContainerExecWait` running `cat` instead. Check both signatures in `services/docker/manager.go` before relying on them.

- [ ] **Step 2: Run it on this machine,** where Docker Desktop has swarm active: `HP_TEST_DOCKER=1 go test ./hivepaas_app/service/settingmountservice/settingmountserviceimpl/ -run TestAgainstARealDaemon -v`. It should PASS. Then run it without the variable, which should SKIP.

- [ ] **Step 3: Update the spec** as listed under **Files**, keeping its wording style.

- [ ] **Step 4: Gates.** Run `go build ./...`, `golangci-lint run ./...` and `go test ./...`. All should be clean or PASS.

- [ ] **Step 5: Commit, merge, delete the branch**

```bash
git add hivepaas_app/service/settingmountservice/settingmountserviceimpl/real_daemon_test.go \
  docs/superpowers/specs/2026-09-25-setting-mounts-design.md
git commit -m "test(settingmounts): a renewed certificate reaches a real container in one update

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main && git merge --no-ff feat/setting-mounts && git branch -d feat/setting-mounts
go test ./...
```

- [ ] **Step 6: Tell the user what to check on the Linux server,** after rebuilding the backend image:
  1. Make an entry through the database or a debug call (the screen comes in plan 4) that mounts a certificate. Deploy, and check the files exist in the container.
  2. Renew the certificate. `task:setting-mount-refresh` runs, the service restarts once, and the old objects are gone.
  3. Disable the entry. The files leave the container.
  4. Delete the app. `docker secret ls --filter label=hivepaas.settingMount.entry` and `docker config ls` with the same filter show nothing of it.
  5. Apps using the Docker API were given the old labels. After redeploying, the agent labels their new objects under the new names. Objects still carrying `hivepaas.docker-api.*` are not reclaimed by the new code, so remove them by hand on that server: they are test data only, and nothing released used those names.
