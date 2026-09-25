# Link App Env Vars Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A "Link App" dialog on an app's Env Variables screen that suggests `${<app>.<VAR>}` variables for another app of the env, computed by the backend.

**Architecture:**
- **Backend.**
  - A new shared system variable, `HIVEPAAS_PASSWORD_URLENCODED`.
  - A pure package, `service/envvarservice/envlink`, holds the recipe table: link targets, and the suggestion groups for a target.
  - Two read endpoints under `/apps/{appID}/env-vars/` call into it.
- **Dashboard.**
  - The env-vars API slice grows two queries.
  - A dialog component adds the chosen rows to the form without saving.

**Tech Stack:** Go (gin, bun, testify), React 19, TanStack Query, React Hook Form, shadcn UI.

**Spec:** `docs/superpowers/specs/2026-09-25-link-app-env-vars-design.md`

## Global Constraints

- **Repositories.**
  - Backend: this repository.
  - Dashboard: `../hivepaas-dashboard`.
  - Each on its own branch, `feat/link-app-env-vars`.
  - Merge locally, delete the branch, do not push, and stage only the files named in the task.
- **Commit trailer:** `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Backend gates:** `go build ./...`, `golangci-lint run ./...` (the whole repo: lll 120, misspell, gci, gosec), `go test ./...`, and `make gen-swag` for the new DTOs.
- **Dashboard gates:** `npm run lint:ci`, `npm run build`. There is no test runner; the dashboard is checked in Task 5 against the backend.
- **New variable:** `HIVEPAAS_PASSWORD_URLENCODED`.
  - The password with every byte outside `A-Z a-z 0-9 - . _ ~` percent-encoded, as `%XX` in upper-case hex.
  - Published and shared by kinds `database` and `cache`.
  - Reserved: a person cannot define it.
  - Masked like the password.
- **Values are references only.** No suggestion carries a resolved value.
- **Endpoints**, read permission:
  - `GET /projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/link-targets`;
  - `GET .../env-vars/link-suggestions?targetAppId=<id>`.
- **Refusal:** a target that is not in the app's env, the app itself, or a preview returns `ERR_APP_NOT_FOUND`.
- **Dialog:** 1000px wide, opened by a "Link App" button before "Show Final Values", on the app screen only. It adds rows to the form and saves nothing.

## Review Focus

1. **An app key that starts with a digit or has hyphens** (`2fa-app`) must still give valid variable names for the address group: `APP_2FA_APP_URL`, never `2FA-APP_URL`. *Task 2, `TestEnvKeyOfAKey`.*
2. **A cache with authentication off** must not get a `redis://:@host` URL: its URL has no credentials. *Task 2, `TestRedisWithoutAPasswordHasNoCredentialsInItsURL`.*
3. **A target from another env, the app itself, or a preview** is refused, not suggested for. *Task 2, `TestTargetsLeaveOutTheAppItselfAndPreviews`; the use case looks targets up through `Targets` only.*
4. **A prefix that makes a key collide with one the form has, or two selected rows with one key,** blocks "Add" until resolved, rather than silently overwriting. *Task 4, `rowIssues`; checked in Task 5's run.*
5. **A password with `@ : / ? # %` or non-ASCII bytes** is encoded byte by byte. *Task 1, `TestPercentEncode`.*

---

### Task 1: The URL-encoded password variable

**Files:**
- Modify: `hivepaas_app/base/env_var.go` (the constant; `AppKindSharedEnvVars`; `mapAppUnallowedVar`; `mapAppSecretVar`)
- Modify: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env.go` (`kindEnvVars`, a new `percentEncode`)
- Test: `hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go`

**Interfaces:**
- Produces: `base.AppSystemEnvVarPasswordURLEncoded = "HIVEPAAS_PASSWORD_URLENCODED"`.

- [ ] **Step 1: Branch.**

```bash
git checkout main && git checkout -b feat/link-app-env-vars
```

- [ ] **Step 2: Write the failing tests.** Append to `app_system_env_test.go`:

```go
func TestPercentEncode(t *testing.T) {
	assert.Equal(t, "", percentEncode(""))
	assert.Equal(t, "Abc-._~09", percentEncode("Abc-._~09"), "the unreserved set is kept")
	assert.Equal(t, "p%40ss%3Aw%2Frd%3F%23%25", percentEncode("p@ss:w/rd?#%"))
	assert.Equal(t, "%C3%BC%20x", percentEncode("ü x"), "byte by byte, a space included")
}

func TestKindEnvVars_TheURLEncodedPasswordIsShared(t *testing.T) {
	for _, kind := range []*entity.AppKindSettings{
		{Category: base.AppCategoryDatabase, Engine: "postgres", Database: &entity.AppKindDatabase{
			Username: "app", Password: entity.NewEncryptedField("p@ss:w"),
			RootPassword: entity.NewEncryptedField("r"), SSLMode: base.DatabaseSslModeDisable,
		}},
		{Category: base.AppCategoryCache, Engine: "redis", Cache: &entity.AppKindCache{
			Password: entity.NewEncryptedField("p@ss:w"),
		}},
	} {
		envs, err := kindEnvVars(kind)
		assert.NoError(t, err)
		byKey := envByKey(t, envs)
		assert.Equal(t, "p%40ss%3Aw", byKey[base.AppSystemEnvVarPasswordURLEncoded].Value, string(kind.Category))
		assert.True(t, byKey[base.AppSystemEnvVarPasswordURLEncoded].IsShared, string(kind.Category))
	}
	assert.False(t, base.IsAppRuntimeEnvAllowed(base.AppSystemEnvVarPasswordURLEncoded), "a person cannot define it")
	assert.True(t, base.IsAppSecretEnv(base.AppSystemEnvVarPasswordURLEncoded), "it is masked as the password is")
}
```

- [ ] **Step 3: Run them.** `go test ./hivepaas_app/service/envvarservice/envvarserviceimpl/ -run 'TestPercentEncode|TestKindEnvVars'`. Expected: FAIL to build, with `undefined: percentEncode` and `undefined: base.AppSystemEnvVarPasswordURLEncoded`.

- [ ] **Step 4: The constant and the lists.** In `base/env_var.go`:
  - After `AppSystemEnvVarPassword`, add:

    ```go
    	// AppSystemEnvVarPasswordURLEncoded is the password percent-encoded, for a
    	// connection string: a password with @ or : would otherwise break the URL.
    	AppSystemEnvVarPasswordURLEncoded = "HIVEPAAS_PASSWORD_URLENCODED" //nolint:gosec // G101: env name
    ```

  - In `AppKindSharedEnvVars`, `AppCategoryDatabase` returns `AppSystemEnvVarUser, AppSystemEnvVarPassword, AppSystemEnvVarPasswordURLEncoded, AppSystemEnvVarDatabaseName, AppSystemEnvVarSSLMode`.
  - `AppCategoryCache` returns `AppSystemEnvVarPassword, AppSystemEnvVarPasswordURLEncoded`.
  - Add `AppSystemEnvVarPasswordURLEncoded: {},` to `mapAppUnallowedVar`, after `AppSystemEnvVarPassword`, and to `mapAppSecretVar`.

- [ ] **Step 5: Publish it.** In `app_system_env.go`:
  - In `kindEnvVars`, the database case, after `sharedEnv(base.AppSystemEnvVarPassword, gofn.Must(db.Password.GetPlain())),`, add:

    ```go
    			sharedEnv(base.AppSystemEnvVarPasswordURLEncoded, percentEncode(gofn.Must(db.Password.GetPlain()))),
    ```

  - The cache case gets the same, with `cache.Password`.
  - At the end of the file:

```go
// percentEncode escapes every byte outside RFC 3986's unreserved set, which is
// safe in any part of a URL: user info, path or query.
func percentEncode(s string) string {
	const hexDigits = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&0x0f])
	}
	return b.String()
}
```

Add `"strings"` to the imports if it is missing.

- [ ] **Step 6: Run the package's tests.** `go test ./hivepaas_app/service/envvarservice/... ./hivepaas_app/service/apptemplateservice/...`. Expected: PASS. The existing test that matches `AppKindSharedEnvVars` against what `kindEnvVars` shares passes with both sides updated.

- [ ] **Step 7: Commit.**

```bash
git add hivepaas_app/base/env_var.go hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env.go hivepaas_app/service/envvarservice/envvarserviceimpl/app_system_env_test.go
git commit -m "feat(envvars): databases and caches share their password URL-encoded

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The recipe table

**Files:**
- Create: `hivepaas_app/service/envvarservice/envlink/envlink.go`, `hivepaas_app/service/envvarservice/envlink/recipes.go`
- Test: `hivepaas_app/service/envvarservice/envlink/envlink_test.go`

**Interfaces:**
- Consumes: `base.AppSystemEnvVarPasswordURLEncoded` (Task 1).
- Produces:
  - `envlink.Target{ID, Key, Name string; Category base.AppCategory; Engine string}`;
  - `envlink.Facts{HasPort, HasPassword bool; SharedVars []string}`;
  - `envlink.Var{Key, Value, Description string}`;
  - `envlink.Group{ID, Title, Description string; Recommended bool; Warnings []string; Vars []*Var}`;
  - `envlink.EngineFamily(engine string) string`;
  - `envlink.Targets(self *entity.App, apps []*entity.App, kinds map[string]*entity.AppKindSettings) []*Target`;
  - `envlink.FactsOf(shared []*envvarservice.EnvVar) *Facts`;
  - `envlink.Suggest(target *Target, facts *Facts) []*Group`.

- [ ] **Step 1: Write the failing tests.** `envlink_test.go`:

```go
package envlink

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func groupIDs(groups []*Group) []string {
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids
}

func varsOf(g *Group) map[string]string {
	out := map[string]string{}
	for _, v := range g.Vars {
		out[v.Key] = v.Value
	}
	return out
}

var ready = &Facts{HasPort: true, HasPassword: true}

func TestEngineFamilies(t *testing.T) {
	for engine, want := range map[string]string{
		"postgres": FamilyPostgres, "PostgreSQL": FamilyPostgres, "timescaledb": FamilyPostgres,
		"mariadb": FamilyMySQL, "mongo": FamilyMongoDB, "valkey": FamilyRedis, "memcached": FamilyMemcached,
		"clickhouse": "", "": "",
	} {
		assert.Equal(t, want, EngineFamily(engine), engine)
	}
}

func TestPostgresGetsAURLTheirPartsAndLibpq(t *testing.T) {
	groups := Suggest(&Target{Key: "pg", Name: "PG", Category: base.AppCategoryDatabase, Engine: "postgres"}, ready)
	assert.Equal(t, []string{"connection-url", "individual", "libpq"}, groupIDs(groups))
	assert.True(t, groups[0].Recommended)
	assert.Equal(t,
		"postgres://${pg.HIVEPAAS_USER}:${pg.HIVEPAAS_PASSWORD_URLENCODED}@${pg.HIVEPAAS_HOST}:${pg.HIVEPAAS_PORT}"+
			"/${pg.HIVEPAAS_DATABASE_NAME}?sslmode=${pg.HIVEPAAS_SSL_MODE}",
		varsOf(groups[0])["DATABASE_URL"])
	assert.Equal(t, "${pg.HIVEPAAS_PASSWORD}", varsOf(groups[1])["DB_PASSWORD"], "a part is the raw password")
	assert.Equal(t, "${pg.HIVEPAAS_HOST}", varsOf(groups[2])["PGHOST"])
}

func TestEachCategoryAndFamilyHasARecipe(t *testing.T) {
	for name, tc := range map[string]struct {
		target *Target
		ids    []string
	}{
		"mysql":        {&Target{Key: "db", Category: base.AppCategoryDatabase, Engine: "mariadb"}, []string{"connection-url", "individual"}},
		"mongodb":      {&Target{Key: "db", Category: base.AppCategoryDatabase, Engine: "mongodb"}, []string{"connection-url", "individual"}},
		"other db":     {&Target{Key: "db", Category: base.AppCategoryDatabase, Engine: "clickhouse"}, []string{"individual"}},
		"redis":        {&Target{Key: "c", Category: base.AppCategoryCache, Engine: "dragonfly"}, []string{"connection-url", "individual"}},
		"memcached":    {&Target{Key: "c", Category: base.AppCategoryCache, Engine: "memcached"}, []string{"servers"}},
		"other cache":  {&Target{Key: "c", Category: base.AppCategoryCache, Engine: "hazelcast"}, []string{"individual"}},
		"storage":      {&Target{Key: "s3", Category: base.AppCategoryStorage, Engine: "minio"}, []string{"s3"}},
		"webapp":       {&Target{Key: "api", Category: base.AppCategoryWebapp}, []string{"addresses"}},
		"without kind": {&Target{Key: "api"}, []string{"addresses"}},
	} {
		groups := Suggest(tc.target, ready)
		assert.Equal(t, tc.ids, groupIDs(groups), name)
		assert.True(t, groups[0].Recommended, name)
	}
}

// Every variable a recipe names is one the target's kind shares, or one every
// app shares: a suggestion never refers to something that is not there.
func TestRecipesNameOnlySharedVariables(t *testing.T) {
	for _, target := range []*Target{
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "postgres"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "mysql"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "mongodb"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "x"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "redis"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "memcached"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "x"},
		{Key: "a", Category: base.AppCategoryStorage},
		{Key: "a", Category: base.AppCategoryWebapp},
	} {
		allowed := map[string]bool{}
		for _, name := range append(append([]string{}, base.AppCommonSharedEnvVars...),
			base.AppKindSharedEnvVars(target.Category)...) {
			allowed[name] = true
		}
		for _, g := range Suggest(target, ready) {
			for _, v := range g.Vars {
				for _, name := range referencedNames(v.Value, target.Key) {
					assert.True(t, allowed[name], "%s/%s names %s", target.Category, target.Engine, name)
				}
			}
		}
	}
}

func TestRedisWithoutAPasswordHasNoCredentialsInItsURL(t *testing.T) {
	groups := Suggest(&Target{Key: "cache", Category: base.AppCategoryCache, Engine: "redis"},
		&Facts{HasPort: true, HasPassword: false})
	assert.Equal(t, "redis://${cache.HIVEPAAS_HOST}:${cache.HIVEPAAS_PORT}/0", varsOf(groups[0])["REDIS_URL"])
}

func TestATargetWithoutAPortIsWarnedAbout(t *testing.T) {
	groups := Suggest(&Target{Key: "pg", Name: "PG", Category: base.AppCategoryDatabase, Engine: "postgres"},
		&Facts{HasPort: false, HasPassword: true})
	for _, g := range groups {
		assert.NotEmpty(t, g.Warnings, g.ID)
		assert.Contains(t, g.Warnings[0], "PG has no container port")
	}
}

func TestATargetWithoutAKindSaysWhy(t *testing.T) {
	groups := Suggest(&Target{Key: "api", Name: "API"}, ready)
	assert.Contains(t, groups[0].Warnings, "API declares no kind, so only its addresses and shared variables are offered.")
}

func TestTheTargetsSharedVariablesComeLast(t *testing.T) {
	groups := Suggest(&Target{Key: "api", Name: "API", Category: base.AppCategoryWebapp},
		&Facts{HasPort: true, SharedVars: []string{"TOKEN_URL", "API_VERSION"}})
	last := groups[len(groups)-1]
	assert.Equal(t, "shared", last.ID)
	assert.Equal(t, "Shared variables of API", last.Title)
	assert.Equal(t, []string{"API_VERSION", "TOKEN_URL"}, []string{last.Vars[0].Key, last.Vars[1].Key})
	assert.Equal(t, "${api.API_VERSION}", last.Vars[0].Value)
}

func TestEnvKeyOfAKey(t *testing.T) {
	assert.Equal(t, "API", envKeyOf("api"))
	assert.Equal(t, "MY_API", envKeyOf("my-api"))
	assert.Equal(t, "APP_2FA_APP", envKeyOf("2fa-app"))
}

func TestTargetsLeaveOutTheAppItselfAndPreviews(t *testing.T) {
	self := &entity.App{ID: "a1", Key: "web", Name: "Web", ProjectEnvID: "e1"}
	apps := []*entity.App{
		self,
		{ID: "a2", Key: "pg", Name: "Postgres", ProjectEnvID: "e1"},
		{ID: "a3", Key: "web-pr-1", Name: "PR", ProjectEnvID: "e1", ParentID: "a1"},
		{ID: "a4", Key: "cache", Name: "Cache", ProjectEnvID: "e1"},
		{ID: "a5", Key: "other", Name: "Other env", ProjectEnvID: "e2"},
	}
	kinds := map[string]*entity.AppKindSettings{"a2": {Category: base.AppCategoryDatabase, Engine: "postgres"}}

	targets := Targets(self, apps, kinds)

	assert.Len(t, targets, 2)
	assert.Equal(t, "Cache", targets[0].Name, "ordered by name")
	assert.Equal(t, &Target{ID: "a2", Key: "pg", Name: "Postgres", Category: base.AppCategoryDatabase,
		Engine: "postgres"}, targets[1])
}

func TestFactsOfASharedEnvironment(t *testing.T) {
	system := func(key, value string) *envvarservice.EnvVar {
		return &envvarservice.EnvVar{EnvVar: &entity.EnvVar{Key: key, Value: value, IsShared: true, IsSystem: true}}
	}
	facts := FactsOf([]*envvarservice.EnvVar{
		system(base.AppSystemEnvVarPort, "5432"),
		system(base.AppSystemEnvVarPassword, ""),
		{EnvVar: &entity.EnvVar{Key: "API_VERSION", Value: "2", IsShared: true}},
	})
	assert.Equal(t, &Facts{HasPort: true, HasPassword: false, SharedVars: []string{"API_VERSION"}}, facts)
}
```

- [ ] **Step 2: Run them.** `go test ./hivepaas_app/service/envvarservice/envlink/`. Expected: FAIL to build (`no non-test Go files` or undefined names).

- [ ] **Step 3: The types, targets and facts.** `envlink.go`:

```go
// Package envlink suggests the variables that link one app to another of its
// env: references to what the target shares, put together the way a client of
// its engine wants them. It holds no values: every suggestion is a reference,
// resolved when the app is built.
package envlink

import (
	"regexp"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// Target is an app another app of its env may link to.
type Target struct {
	ID       string
	Key      string
	Name     string
	Category base.AppCategory
	Engine   string
}

// Facts are what the suggestions depend on besides the target's kind.
type Facts struct {
	HasPort     bool
	HasPassword bool
	// SharedVars are the keys of the variables the target's own settings share,
	// not its system variables.
	SharedVars []string
}

// Var is one suggested variable.
type Var struct {
	Key         string
	Value       string
	Description string
}

// Group is suggested variables that go together.
type Group struct {
	ID          string
	Title       string
	Description string
	Recommended bool
	Warnings    []string
	Vars        []*Var
}

// Targets are the apps of self's env another may link to: not self, not a
// preview, ordered by name.
func Targets(self *entity.App, apps []*entity.App, kinds map[string]*entity.AppKindSettings) []*Target {
	targets := make([]*Target, 0, len(apps))
	for _, app := range apps {
		if app.ID == self.ID || app.ParentID != "" || app.ProjectEnvID != self.ProjectEnvID {
			continue
		}
		target := &Target{ID: app.ID, Key: app.Key, Name: app.Name}
		if kind := kinds[app.ID]; kind != nil {
			target.Category, target.Engine = kind.Category, kind.Engine
		}
		targets = append(targets, target)
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets
}

// FactsOf reads the facts from the target's shared environment.
func FactsOf(shared []*envvarservice.EnvVar) *Facts {
	facts := &Facts{}
	for _, env := range shared {
		switch {
		case env.IsSystem && env.Key == base.AppSystemEnvVarPort:
			facts.HasPort = env.Value != ""
		case env.IsSystem && env.Key == base.AppSystemEnvVarPassword:
			facts.HasPassword = env.Value != ""
		case !env.IsSystem:
			facts.SharedVars = append(facts.SharedVars, env.Key)
		}
	}
	return facts
}

var notEnvKeyChars = regexp.MustCompile(`[^A-Z0-9_]+`)

// envKeyOf is an app key made a variable name: upper case, anything else an
// underscore, and never starting with a digit.
func envKeyOf(key string) string {
	name := notEnvKeyChars.ReplaceAllString(strings.ToUpper(key), "_")
	if name == "" || (name[0] >= '0' && name[0] <= '9') {
		name = "APP_" + name
	}
	return name
}

var reReference = regexp.MustCompile(`\$\{([a-zA-Z0-9_-]+)\.([A-Za-z_][A-Za-z0-9_]*)\}`)

// referencedNames are the variables of app a value refers to.
func referencedNames(value, app string) []string {
	var names []string
	for _, m := range reReference.FindAllStringSubmatch(value, -1) {
		if m[1] == app {
			names = append(names, m[2])
		}
	}
	return names
}
```

- [ ] **Step 4: The recipes.** `recipes.go`:

```go
package envlink

import (
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	FamilyPostgres  = "postgres"
	FamilyMySQL     = "mysql"
	FamilyMongoDB   = "mongodb"
	FamilyRedis     = "redis"
	FamilyMemcached = "memcached"
)

var families = map[string]string{
	"postgres": FamilyPostgres, "postgresql": FamilyPostgres, "timescaledb": FamilyPostgres,
	"postgis": FamilyPostgres, "pgvector": FamilyPostgres,
	"mysql": FamilyMySQL, "mariadb": FamilyMySQL, "percona": FamilyMySQL,
	"mongodb": FamilyMongoDB, "mongo": FamilyMongoDB, "ferretdb": FamilyMongoDB,
	"redis": FamilyRedis, "valkey": FamilyRedis, "keydb": FamilyRedis, "dragonfly": FamilyRedis,
	"memcached": FamilyMemcached,
}

// EngineFamily is the family an engine belongs to, or "" for one the table does
// not know.
func EngineFamily(engine string) string {
	return families[strings.ToLower(strings.TrimSpace(engine))]
}

// refs writes references to one app's variables.
type refs struct{ app string }

func (r refs) of(name string) string { return "${" + r.app + "." + name + "}" }

// Suggest is what to add to link to target, most useful first, the target's own
// shared variables last.
func Suggest(target *Target, facts *Facts) []*Group {
	r := refs{app: target.Key}
	var groups []*Group
	switch target.Category {
	case base.AppCategoryDatabase:
		groups = databaseGroups(r, EngineFamily(target.Engine))
	case base.AppCategoryCache:
		groups = cacheGroups(r, EngineFamily(target.Engine), facts.HasPassword)
	case base.AppCategoryStorage:
		groups = []*Group{storageGroup(r)}
	case base.AppCategoryWebapp:
		groups = []*Group{addressGroup(r, target)}
	default:
		address := addressGroup(r, target)
		address.Warnings = append(address.Warnings,
			target.Name+" declares no kind, so only its addresses and shared variables are offered.")
		groups = []*Group{address}
	}
	if !facts.HasPort {
		portRef := r.of(base.AppSystemEnvVarPort)
		for _, g := range groups {
			for _, v := range g.Vars {
				if strings.Contains(v.Value, portRef) {
					g.Warnings = append(g.Warnings, target.Name+" has no container port; set one in its "+
						"Routing Settings, or the port below is empty.")
					break
				}
			}
		}
	}
	if shared := sharedGroup(r, target, facts.SharedVars); shared != nil {
		groups = append(groups, shared)
	}
	return groups
}

func urlGroup(key, value string) *Group {
	return &Group{
		ID: "connection-url", Title: "Connection string", Recommended: true,
		Description: "One URL most client libraries accept.",
		Vars:        []*Var{{Key: key, Value: value, Description: "Connection URL"}},
	}
}

func individualGroup(recommended bool, vars ...*Var) *Group {
	return &Group{
		ID: "individual", Title: "Individual variables", Recommended: recommended,
		Description: "Each part on its own, for a client configured part by part.",
		Vars:        vars,
	}
}

func databaseGroups(r refs, family string) []*Group {
	host, port := r.of(base.AppSystemEnvVarHost), r.of(base.AppSystemEnvVarPort)
	user, password := r.of(base.AppSystemEnvVarUser), r.of(base.AppSystemEnvVarPassword)
	name, sslMode := r.of(base.AppSystemEnvVarDatabaseName), r.of(base.AppSystemEnvVarSSLMode)
	auth := user + ":" + r.of(base.AppSystemEnvVarPasswordURLEncoded) + "@" + host + ":" + port

	switch family {
	case FamilyPostgres:
		return []*Group{
			urlGroup("DATABASE_URL", "postgres://"+auth+"/"+name+"?sslmode="+sslMode),
			individualGroup(false,
				&Var{Key: "DB_HOST", Value: host}, &Var{Key: "DB_PORT", Value: port},
				&Var{Key: "DB_USER", Value: user}, &Var{Key: "DB_PASSWORD", Value: password},
				&Var{Key: "DB_NAME", Value: name}, &Var{Key: "DB_SSL_MODE", Value: sslMode}),
			{
				ID: "libpq", Title: "libpq variables",
				Description: "What libpq and most Postgres drivers read with no configuration.",
				Vars: []*Var{
					{Key: "PGHOST", Value: host}, {Key: "PGPORT", Value: port}, {Key: "PGUSER", Value: user},
					{Key: "PGPASSWORD", Value: password}, {Key: "PGDATABASE", Value: name},
					{Key: "PGSSLMODE", Value: sslMode},
				},
			},
		}
	case FamilyMySQL:
		return []*Group{
			urlGroup("DATABASE_URL", "mysql://"+auth+"/"+name),
			individualGroup(false,
				&Var{Key: "DB_HOST", Value: host}, &Var{Key: "DB_PORT", Value: port},
				&Var{Key: "DB_USER", Value: user}, &Var{Key: "DB_PASSWORD", Value: password},
				&Var{Key: "DB_NAME", Value: name}),
		}
	case FamilyMongoDB:
		return []*Group{
			urlGroup("MONGODB_URI", "mongodb://"+auth+"/"+name+"?authSource=admin"),
			individualGroup(false,
				&Var{Key: "MONGO_HOST", Value: host}, &Var{Key: "MONGO_PORT", Value: port},
				&Var{Key: "MONGO_USER", Value: user}, &Var{Key: "MONGO_PASSWORD", Value: password},
				&Var{Key: "MONGO_DATABASE", Value: name}),
		}
	}
	return []*Group{individualGroup(true,
		&Var{Key: "DB_HOST", Value: host}, &Var{Key: "DB_PORT", Value: port},
		&Var{Key: "DB_USER", Value: user}, &Var{Key: "DB_PASSWORD", Value: password},
		&Var{Key: "DB_NAME", Value: name})}
}

func cacheGroups(r refs, family string, hasPassword bool) []*Group {
	host, port := r.of(base.AppSystemEnvVarHost), r.of(base.AppSystemEnvVarPort)
	password := r.of(base.AppSystemEnvVarPassword)

	switch family {
	case FamilyRedis:
		// A cache with authentication off gets a URL without credentials: some
		// clients refuse an empty password in one.
		auth := ""
		if hasPassword {
			auth = ":" + r.of(base.AppSystemEnvVarPasswordURLEncoded) + "@"
		}
		return []*Group{
			urlGroup("REDIS_URL", "redis://"+auth+host+":"+port+"/0"),
			individualGroup(false, &Var{Key: "REDIS_HOST", Value: host}, &Var{Key: "REDIS_PORT", Value: port},
				&Var{Key: "REDIS_PASSWORD", Value: password}),
		}
	case FamilyMemcached:
		return []*Group{{
			ID: "servers", Title: "Servers", Recommended: true,
			Description: "The address list memcached clients take.",
			Vars:        []*Var{{Key: "MEMCACHED_SERVERS", Value: host + ":" + port}},
		}}
	}
	return []*Group{individualGroup(true, &Var{Key: "CACHE_HOST", Value: host},
		&Var{Key: "CACHE_PORT", Value: port}, &Var{Key: "CACHE_PASSWORD", Value: password})}
}

func storageGroup(r refs) *Group {
	return &Group{
		ID: "s3", Title: "S3 client", Recommended: true,
		Description: "What S3 SDKs read: the endpoint inside the cluster, the key pair, the bucket and the region.",
		Vars: []*Var{
			{Key: "S3_ENDPOINT", Value: "http://" + r.of(base.AppSystemEnvVarHost) + ":" + r.of(base.AppSystemEnvVarPort)},
			{Key: "AWS_ACCESS_KEY_ID", Value: r.of(base.AppSystemEnvVarKeyID)},
			{Key: "AWS_SECRET_ACCESS_KEY", Value: r.of(base.AppSystemEnvVarSecret)},
			{Key: "S3_BUCKET", Value: r.of(base.AppSystemEnvVarBucket)},
			{Key: "AWS_REGION", Value: r.of(base.AppSystemEnvVarRegion)},
		},
	}
}

func addressGroup(r refs, target *Target) *Group {
	name := envKeyOf(target.Key)
	return &Group{
		ID: "addresses", Title: "Addresses", Recommended: true,
		Description: "Where to reach the app: inside the cluster, and from outside.",
		Vars: []*Var{
			{Key: name + "_URL", Value: "http://" + r.of(base.AppSystemEnvVarHost) + ":" + r.of(base.AppSystemEnvVarPort),
				Description: "Inside the cluster"},
			{Key: name + "_PUBLIC_URL", Value: r.of(base.AppSystemEnvVarAppURL),
				Description: "From outside; empty while the app has no domain"},
		},
	}
}

func sharedGroup(r refs, target *Target, keys []string) *Group {
	if len(keys) == 0 {
		return nil
	}
	sorted := append([]string{}, keys...)
	sort.Strings(sorted)
	vars := make([]*Var, 0, len(sorted))
	for _, key := range sorted {
		vars = append(vars, &Var{Key: key, Value: r.of(key)})
	}
	return &Group{
		ID: "shared", Title: "Shared variables of " + target.Name,
		Description: "Variables " + target.Name + " declares shared.",
		Vars:        vars,
	}
}
```

- [ ] **Step 5: Run the tests.** `go test ./hivepaas_app/service/envvarservice/envlink/`. Expected: PASS.

- [ ] **Step 6: Lint.** `golangci-lint run ./hivepaas_app/service/envvarservice/...`. Expected: 0 issues. Split any line over 120 characters.

- [ ] **Step 7: Commit.**

```bash
git add hivepaas_app/service/envvarservice/envlink
git commit -m "feat(envvars): the recipes that link an app to another of its env

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The endpoints

**Files:**
- Create: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/env_vars_link.go`, `hivepaas_app/usecase/appsettingsuc/env_vars_link.go`, `hivepaas_app/interface/api/handler/appsettingshandler/env_var_link.go`
- Modify: `hivepaas_app/interface/api/server/router_apps.go` (two routes after `/compute`)
- Test: `hivepaas_app/interface/api/server/router_env_link_test.go`
- Regenerate: `docs/openapi/swagger.json` (`make gen-swag`)

**Interfaces:**
- Consumes: Task 2's `envlink.Targets`, `envlink.FactsOf`, `envlink.Suggest`.
- Produces, for the dashboard:
  - wire `link-targets`: `{data: [{id, key, name, category, engine}]}`;
  - wire `link-suggestions`: `{data: {target: {…}, groups: [{id, title, description, recommended, warnings: string[], vars: [{key, value, description}]}]}}`.

- [ ] **Step 1: Write the failing route test.** `router_env_link_test.go`. It uses `projectRoutes` from `router_config_files_test.go`, which registers the app routes too.

```go
package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppEnvVarsHaveLinkRoutes(t *testing.T) {
	routes := projectRoutes(t)
	base := "/projects/:projectID/:projectEnv/apps/:appID/env-vars"
	assert.True(t, routes[http.MethodGet+" "+base+"/link-targets"])
	assert.True(t, routes[http.MethodGet+" "+base+"/link-suggestions"])
}
```

- [ ] **Step 2: Run it.** `go test ./hivepaas_app/interface/api/server/ -run TestAppEnvVarsHaveLinkRoutes`. Expected: FAIL, both assertions false.

- [ ] **Step 3: DTOs.** `appsettingsdto/env_vars_link.go`:

```go
package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envlink"
)

type ListEnvLinkTargetsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewListEnvLinkTargetsReq() *ListEnvLinkTargetsReq {
	return &ListEnvLinkTargetsReq{}
}

func (req *ListEnvLinkTargetsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListEnvLinkTargetsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data []*EnvLinkTargetResp `json:"data"`
}

type EnvLinkTargetResp struct {
	ID       string           `json:"id"`
	Key      string           `json:"key"`
	Name     string           `json:"name"`
	Category base.AppCategory `json:"category"`
	Engine   string           `json:"engine"`
}

type GetEnvLinkSuggestionsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	TargetAppID  string `json:"-" mapstructure:"targetAppId"`
}

func NewGetEnvLinkSuggestionsReq() *GetEnvLinkSuggestionsReq {
	return &GetEnvLinkSuggestionsReq{}
}

func (req *GetEnvLinkSuggestionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.TargetAppID, true, "targetAppId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetEnvLinkSuggestionsResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *EnvLinkSuggestionsResp `json:"data"`
}

type EnvLinkSuggestionsResp struct {
	Target *EnvLinkTargetResp  `json:"target"`
	Groups []*EnvLinkGroupResp `json:"groups"`
}

type EnvLinkGroupResp struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Recommended bool              `json:"recommended"`
	Warnings    []string          `json:"warnings"`
	Vars        []*EnvLinkVarResp `json:"vars"`
}

type EnvLinkVarResp struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func TransformEnvLinkTarget(target *envlink.Target) *EnvLinkTargetResp {
	return &EnvLinkTargetResp{ID: target.ID, Key: target.Key, Name: target.Name,
		Category: target.Category, Engine: target.Engine}
}

func TransformEnvLinkGroups(groups []*envlink.Group) []*EnvLinkGroupResp {
	resp := make([]*EnvLinkGroupResp, 0, len(groups))
	for _, g := range groups {
		vars := make([]*EnvLinkVarResp, 0, len(g.Vars))
		for _, v := range g.Vars {
			vars = append(vars, &EnvLinkVarResp{Key: v.Key, Value: v.Value, Description: v.Description})
		}
		resp = append(resp, &EnvLinkGroupResp{ID: g.ID, Title: g.Title, Description: g.Description,
			Recommended: g.Recommended, Warnings: append([]string{}, g.Warnings...), Vars: vars})
	}
	return resp
}
```

- [ ] **Step 4: The use case.** `appsettingsuc/env_vars_link.go`:

```go
package appsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envlink"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ListEnvLinkTargets(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.ListEnvLinkTargetsReq,
) (*appsettingsdto.ListEnvLinkTargetsResp, error) {
	self, apps, kinds, err := uc.loadEnvLinkApps(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.ListEnvLinkTargetsResp{
		Data: gofn.MapSlice(envlink.Targets(self, apps, kinds), appsettingsdto.TransformEnvLinkTarget),
	}, nil
}

func (uc *UC) GetEnvLinkSuggestions(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetEnvLinkSuggestionsReq,
) (*appsettingsdto.GetEnvLinkSuggestionsResp, error) {
	self, apps, kinds, err := uc.loadEnvLinkApps(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// A target is one Targets offers: of the env, not the app, not a preview.
	target, found := gofn.Find(envlink.Targets(self, apps, kinds), func(t *envlink.Target) bool {
		return t.ID == req.TargetAppID
	})
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", req.TargetAppID)
	}
	targetApp, _ := gofn.Find(apps, func(a *entity.App) bool { return a.ID == target.ID })

	// Read, never returned: only whether the port and the password are set.
	shared, err := uc.envVarService.BuildSharedEnvVarsInApp(ctx, uc.db, targetApp, envvarservice.EnvBuildOptions{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appsettingsdto.GetEnvLinkSuggestionsResp{Data: &appsettingsdto.EnvLinkSuggestionsResp{
		Target: appsettingsdto.TransformEnvLinkTarget(target),
		Groups: appsettingsdto.TransformEnvLinkGroups(envlink.Suggest(target, envlink.FactsOf(shared))),
	}}, nil
}

// loadEnvLinkApps is the app, the apps of its env with what building their
// environment needs, and their kinds by app id.
func (uc *UC) loadEnvLinkApps(
	ctx context.Context, projectID, appID string,
) (*entity.App, []*entity.App, map[string]*entity.AppKindSettings, error) {
	self, err := uc.appService.LoadApp(ctx, uc.db, projectID, appID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	apps, _, err := uc.appRepo.List(ctx, uc.db, projectID, nil,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectWhere("app.project_env_id = ?", self.ProjectEnvID),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	kinds := make(map[string]*entity.AppKindSettings, len(apps))
	if len(apps) == 0 {
		return self, apps, kinds, nil
	}
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.object_id IN (?)", gofn.MapSlice(apps, func(a *entity.App) string { return a.ID })...),
	)
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	for _, setting := range settings {
		kind, err := setting.AsAppKindSettings()
		if err != nil {
			return nil, nil, nil, hperrors.Wrap(err)
		}
		kinds[setting.ObjectID] = kind
	}
	return self, apps, kinds, nil
}
```

- [ ] **Step 5: The handler and routes.** `appsettingshandler/env_var_link.go`:

```go
package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// ListEnvLinkTargets Lists the apps an app may link its env vars to
// @Summary Lists the apps an app may link its env vars to
// @Description Lists the apps an app may link its env vars to
// @Tags    app_settings
// @Produce json
// @Id      listAppEnvLinkTargets
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appsettingsdto.ListEnvLinkTargetsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/link-targets [get]
func (h *Handler) ListEnvLinkTargets(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewListEnvLinkTargetsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.ListEnvLinkTargets(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetEnvLinkSuggestions Suggests the env vars that link an app to another
// @Summary Suggests the env vars that link an app to another
// @Description Suggests the env vars that link an app to another
// @Tags    app_settings
// @Produce json
// @Id      getAppEnvLinkSuggestions
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   targetAppId query string true "the app to link to"
// @Success 200 {object} appsettingsdto.GetEnvLinkSuggestionsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/link-suggestions [get]
func (h *Handler) GetEnvLinkSuggestions(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewGetEnvLinkSuggestionsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.GetEnvLinkSuggestions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
```

In `router_apps.go`, after `envVarGroup.POST("/compute", appSettingsHandler.BuildEnvVars)`:

```go
		envVarGroup.GET("/link-targets", appSettingsHandler.ListEnvLinkTargets)
		envVarGroup.GET("/link-suggestions", appSettingsHandler.GetEnvLinkSuggestions)
```

The handler holds the concrete `*appsettingsuc.UC`, so nothing else needs the new methods declared.

- [ ] **Step 6: Run the route test and the gates.**
  - `go test ./hivepaas_app/interface/api/server/`. Expected: PASS.
  - `go build ./... && golangci-lint run ./... && go test ./...`. Expected: all clean.
  - `make gen-swag`. Expected: `docs/openapi/swagger.json` gains both paths.

- [ ] **Step 7: Commit.**

```bash
git add hivepaas_app/usecase/appsettingsuc/appsettingsdto/env_vars_link.go hivepaas_app/usecase/appsettingsuc/env_vars_link.go hivepaas_app/interface/api/handler/appsettingshandler/env_var_link.go hivepaas_app/interface/api/server/router_apps.go hivepaas_app/interface/api/server/router_env_link_test.go docs/openapi/swagger.json
git commit -m "feat(envvars): endpoints that suggest the env vars linking an app to another

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: The Link App dialog

**Files (dashboard, `src/application/modules/projects`):**
- Modify:
  - `api/services/project-apps-services/env-vars/project-app-env-vars.api.contracts.ts`;
  - `.../project-app-env-vars.api.validator.ts`;
  - `.../project-app-env-vars.api.ts`;
  - `api/hooks/project-apps/use-project-app-env-vars.api.ts`;
  - `data/queries/project-apps/project-app-env-vars.queries.ts`;
  - `data/constants/projects.query-keys.ts`.
- Create:
  - `routes/single-project/single-app/configuration/env-variables/link-app/link-app.utils.ts`;
  - `.../link-app/link-app-dialog.com.tsx`;
  - `.../link-app/index.ts`.
- Modify:
  - `module-shared/form/env-vars/env-vars.form.com.tsx` (an `onLinkApp` prop);
  - `routes/single-project/single-app/configuration/env-variables/form/app-config-env-vars.form.com.tsx` (open the dialog, add rows).

**Interfaces:**
- Consumes: Task 3's wire format.
- Produces:
  - `ProjectAppEnvVarsQueries.useFindLinkTargets`;
  - `ProjectAppEnvVarsQueries.useFindLinkSuggestions`;
  - `LinkAppDialog`.

- [ ] **Step 1: Branch.** `cd ../hivepaas-dashboard && git checkout main && git checkout -b feat/link-app-env-vars`

- [ ] **Step 2: Contracts.** Append to `project-app-env-vars.api.contracts.ts`:

```ts
export type EnvLinkTarget = {
    id: string;
    key: string;
    name: string;
    category: string;
    engine: string;
};

export type EnvLinkVar = {
    key: string;
    value: string;
    description: string;
};

export type EnvLinkGroup = {
    id: string;
    title: string;
    description: string;
    recommended: boolean;
    warnings: string[];
    vars: EnvLinkVar[];
};

export type ProjectAppEnvVars_FindLinkTargets_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type ProjectAppEnvVars_FindLinkTargets_Res = ApiResponseBase<EnvLinkTarget[]>;

export type ProjectAppEnvVars_FindLinkSuggestions_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    targetAppID: string;
}>;
export type ProjectAppEnvVars_FindLinkSuggestions_Res = ApiResponseBase<{
    target: EnvLinkTarget;
    groups: EnvLinkGroup[];
}>;
```

- [ ] **Step 3: Validator.** In `project-app-env-vars.api.validator.ts`, add these schemas above the class:

```ts
const EnvLinkTargetSchema = z.object({
    id: z.string(),
    key: z.string(),
    name: z.string(),
    category: z.string().optional().default(""),
    engine: z.string().optional().default(""),
});

const FindLinkTargetsSchema = z.object({
    data: z.array(EnvLinkTargetSchema).nullable().transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullable().optional().default(null),
});

const FindLinkSuggestionsSchema = z.object({
    data: z.object({
        target: EnvLinkTargetSchema,
        groups: z.array(
            z.object({
                id: z.string(),
                title: z.string(),
                description: z.string().optional().default(""),
                recommended: z.boolean().optional().default(false),
                warnings: z.array(z.string()).nullable().optional().transform(value => value ?? []),
                vars: z.array(
                    z.object({
                        key: z.string(),
                        value: z.string(),
                        description: z.string().optional().default(""),
                    }),
                ),
            }),
        ),
    }),
    meta: BaseMetaApiSchema.nullable().optional().default(null),
});
```

and these methods inside the class:

```ts
    findLinkTargets = (response: AxiosResponse): ProjectAppEnvVars_FindLinkTargets_Res => {
        return parseApiResponse({ response, schema: FindLinkTargetsSchema });
    };

    findLinkSuggestions = (response: AxiosResponse): ProjectAppEnvVars_FindLinkSuggestions_Res => {
        return parseApiResponse({ response, schema: FindLinkSuggestionsSchema });
    };
```

Import the two `_Res` types beside the existing ones, and `BaseMetaApiSchema` if it is not imported yet.

- [ ] **Step 4: API.** In `project-app-env-vars.api.ts`, add these methods to the class, following `findOne`:

```ts
    async findLinkTargets(
        request: ProjectAppEnvVars_FindLinkTargets_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_FindLinkTargets_Res, Error>> {
        const { projectID, env, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/env-vars/link-targets`, { signal }),
            ).pipe(
                map(this.validator.findLinkTargets),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findLinkSuggestions(
        request: ProjectAppEnvVars_FindLinkSuggestions_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_FindLinkSuggestions_Res, Error>> {
        const { projectID, env, appID, targetAppID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/env-vars/link-suggestions`, {
                    params: { targetAppId: targetAppID },
                    signal,
                }),
            ).pipe(
                map(this.validator.findLinkSuggestions),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
```

- [ ] **Step 5: Hook, keys and queries.**
  - **Hook.** In `use-project-app-env-vars.api.ts`, add these to `queries`, after `findOne`, and import the two `_Req` types:

```ts
                findLinkTargets: async (data: ProjectAppEnvVars_FindLinkTargets_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.envVars.$.findLinkTargets({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                findLinkSuggestions: async (
                    data: ProjectAppEnvVars_FindLinkSuggestions_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.envVars.$.findLinkSuggestions({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
```

  - **Query keys.** In `projects.query-keys.ts`, after `projects.apps.env-vars.$.find-one`:

```ts
    "projects.apps.env-vars.$.find-link-targets": "projects.apps.env-vars.$.find-link-targets",
    "projects.apps.env-vars.$.find-link-suggestions": "projects.apps.env-vars.$.find-link-suggestions",
```

  - **Queries.** In `project-app-env-vars.queries.ts`:

```ts
type FindLinkTargetsReq = ProjectAppEnvVars_FindLinkTargets_Req["data"];
type FindLinkTargetsOptions = Omit<UseQueryOptions<ProjectAppEnvVars_FindLinkTargets_Res>, "queryKey" | "queryFn">;

function useFindLinkTargets(request: FindLinkTargetsReq, options: FindLinkTargetsOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-link-targets"], request],
        queryFn: ({ signal }) => queries.findLinkTargets(request, signal),
        ...options,
    });
}

type FindLinkSuggestionsReq = ProjectAppEnvVars_FindLinkSuggestions_Req["data"];
type FindLinkSuggestionsOptions = Omit<
    UseQueryOptions<ProjectAppEnvVars_FindLinkSuggestions_Res>,
    "queryKey" | "queryFn"
>;

function useFindLinkSuggestions(request: FindLinkSuggestionsReq, options: FindLinkSuggestionsOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-link-suggestions"], request],
        queryFn: ({ signal }) => queries.findLinkSuggestions(request, signal),
        ...options,
    });
}
```

and `ProjectAppEnvVarsQueries` exports `useFindOne, useFindLinkTargets, useFindLinkSuggestions`.

- [ ] **Step 6: The row helpers.** `link-app/link-app.utils.ts`:

```ts
import type { EnvLinkGroup } from "~/projects/api/services";

export type LinkSection = "runtime" | "buildtime";

export interface LinkRow {
    id: string;
    groupId: string;
    key: string;
    value: string;
    description: string;
    selected: boolean;
    /** Replace the form's variable of the same key. */
    replace: boolean;
}

export type RowIssue = "" | "invalid" | "exists" | "duplicate";

const ENV_KEY_PATTERN = /^[A-Za-z_][A-Za-z0-9_]*$/;

/** One row per suggested variable; a recommended group's rows start selected. */
export function rowsFromGroups(groups: EnvLinkGroup[]): LinkRow[] {
    return groups.flatMap(group =>
        group.vars.map((suggested, index) => ({
            id: `${group.id}:${index}`,
            groupId: group.id,
            key: suggested.key,
            value: suggested.value,
            description: suggested.description,
            selected: group.recommended,
            replace: false,
        })),
    );
}

export function finalKey(prefix: string, key: string): string {
    return `${prefix.trim()}${key.trim()}`;
}

/**
 * What stops a selected row from being added: a key that is not a variable
 * name, one the form already has in the section (unless the row replaces it),
 * or one another selected row also has.
 */
export function rowIssues(rows: LinkRow[], prefix: string, existing: Set<string>): Record<string, RowIssue> {
    const counts = new Map<string, number>();
    for (const row of rows) {
        if (row.selected) {
            const key = finalKey(prefix, row.key);
            counts.set(key, (counts.get(key) ?? 0) + 1);
        }
    }
    const issues: Record<string, RowIssue> = {};
    for (const row of rows) {
        const key = finalKey(prefix, row.key);
        issues[row.id] = !row.selected
            ? ""
            : !ENV_KEY_PATTERN.test(key)
              ? "invalid"
              : (counts.get(key) ?? 0) > 1
                ? "duplicate"
                : existing.has(key) && !row.replace
                  ? "exists"
                  : "";
    }
    return issues;
}
```

- [ ] **Step 7: The dialog.** `link-app/link-app-dialog.com.tsx`:

```tsx
import { useEffect, useMemo, useState } from "react";

import { Badge } from "@components/ui/badge";
import { AlertTriangle } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { EnvLinkTarget } from "~/projects/api/services";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";

import { AppLoader, Combobox } from "@application/shared/components";

import {
    Button,
    Checkbox,
    Dialog,
    DialogBody,
    DialogFixedContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import { type LinkRow, type LinkSection, finalKey, rowIssues, rowsFromGroups } from "./link-app.utils";

const ISSUE_TEXT = {
    invalid: "Not a variable name",
    duplicate: "Another selected row has this name",
    exists: "The form already has this variable",
} as const;

export function LinkAppDialog({ open, initialSection, onOpenChange, existingKeys, onAdd }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId && env && appId, "the app's route params must be defined");

    const [targetID, setTargetID] = useState<string | null>(null);
    const [prefix, setPrefix] = useState("");
    const [section, setSection] = useState<LinkSection>(initialSection);
    const [rows, setRows] = useState<LinkRow[]>([]);

    useEffect(() => {
        if (open) {
            setSection(initialSection);
        } else {
            setTargetID(null);
            setPrefix("");
            setRows([]);
        }
    }, [open, initialSection]);

    const targetsQuery = ProjectAppEnvVarsQueries.useFindLinkTargets(
        { projectID: projectId, env, appID: appId },
        { enabled: open },
    );
    const suggestionsQuery = ProjectAppEnvVarsQueries.useFindLinkSuggestions(
        { projectID: projectId, env, appID: appId, targetAppID: targetID ?? "" },
        { enabled: open && Boolean(targetID) },
    );
    const groups = useMemo(() => suggestionsQuery.data?.data.groups ?? [], [suggestionsQuery.data]);

    useEffect(() => {
        setRows(rowsFromGroups(groups));
    }, [groups]);

    const existing = existingKeys(section);
    const issues = rowIssues(rows, prefix, existing);
    const selected = rows.filter(row => row.selected);
    const blocked = selected.some(row => issues[row.id] !== "");

    function updateRow(id: string, change: Partial<LinkRow>) {
        setRows(current => current.map(row => (row.id === id ? { ...row, ...change } : row)));
    }

    const targets = targetsQuery.data?.data ?? [];
    const targetOptions = targets.map(target => ({ value: { id: target.id, name: target.name }, label: target.name }));

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="flex h-[80vh] w-[1000px] max-w-[calc(100vw-1rem)] flex-col">
                <DialogHeader>
                    <DialogTitle>Link to another app</DialogTitle>
                </DialogHeader>

                <DialogBody className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_200px_auto]">
                        <Combobox
                            options={targetOptions}
                            value={targetID}
                            onChange={(_, option) => {
                                setTargetID(option?.id ? String(option.id) : null);
                            }}
                            placeholder="Choose an app of this env"
                            emptyText="No other app in this env"
                            valueKey="id"
                            loading={targetsQuery.isFetching}
                            renderOption={option => (
                                <TargetOption target={targets.find(target => target.id === option.value.id)} />
                            )}
                        />
                        <Input
                            aria-label="Prefix"
                            placeholder="Prefix, e.g. ANALYTICS_"
                            value={prefix}
                            onChange={event => {
                                setPrefix(event.target.value);
                            }}
                        />
                        <Tabs
                            value={section}
                            onValueChange={value => {
                                setSection(value as LinkSection);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value="runtime">Runtime</TabsTrigger>
                                <TabsTrigger value="buildtime">Build-time</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </div>

                    {!targetID && (
                        <p className="text-sm text-muted-foreground">
                            Choose an app to see the variables that connect to it. Each is a reference such as{" "}
                            <code>{"${db.HIVEPAAS_HOST}"}</code>, resolved when this app is built.
                        </p>
                    )}

                    {targetID && suggestionsQuery.isFetching && (
                        <div className="flex flex-1 items-center justify-center">
                            <AppLoader />
                        </div>
                    )}

                    {targetID &&
                        !suggestionsQuery.isFetching &&
                        groups.map(group => {
                            const groupRows = rows.filter(row => row.groupId === group.id);
                            const allSelected = groupRows.length > 0 && groupRows.every(row => row.selected);
                            return (
                                <div
                                    key={group.id}
                                    className="flex flex-col gap-2 rounded-md border p-3"
                                >
                                    <div className="flex items-center gap-2">
                                        <Checkbox
                                            checked={allSelected}
                                            onCheckedChange={checked => {
                                                setRows(current =>
                                                    current.map(row =>
                                                        row.groupId === group.id
                                                            ? { ...row, selected: checked === true }
                                                            : row,
                                                    ),
                                                );
                                            }}
                                        />
                                        <span className="font-medium">{group.title}</span>
                                        {group.recommended && <Badge variant="outline">Recommended</Badge>}
                                    </div>
                                    {group.description && (
                                        <p className="text-sm text-muted-foreground">{group.description}</p>
                                    )}
                                    {group.warnings.map(warning => (
                                        <p
                                            key={warning}
                                            className="flex items-center gap-1.5 text-xs text-amber-600"
                                        >
                                            <AlertTriangle className="size-3.5 shrink-0" />
                                            {warning}
                                        </p>
                                    ))}
                                    {groupRows.map(row => {
                                        const issue = issues[row.id];
                                        return (
                                            <div
                                                key={row.id}
                                                className="grid grid-cols-[auto_240px_1fr] items-start gap-2"
                                            >
                                                <Checkbox
                                                    className="mt-2.5"
                                                    checked={row.selected}
                                                    onCheckedChange={checked => {
                                                        updateRow(row.id, { selected: checked === true });
                                                    }}
                                                />
                                                <div className="flex flex-col gap-1">
                                                    <Input
                                                        aria-label={`${row.key} name`}
                                                        value={row.key}
                                                        aria-invalid={issue !== ""}
                                                        onChange={event => {
                                                            updateRow(row.id, { key: event.target.value });
                                                        }}
                                                    />
                                                    {prefix.trim() && (
                                                        <span className="text-xs text-muted-foreground">
                                                            {finalKey(prefix, row.key)}
                                                        </span>
                                                    )}
                                                    {issue !== "" && (
                                                        <span className="text-xs text-destructive">
                                                            {ISSUE_TEXT[issue]}
                                                        </span>
                                                    )}
                                                    {(issue === "exists" || row.replace) && row.selected && (
                                                        <label className="flex items-center gap-1.5 text-xs">
                                                            <Checkbox
                                                                checked={row.replace}
                                                                onCheckedChange={checked => {
                                                                    updateRow(row.id, { replace: checked === true });
                                                                }}
                                                            />
                                                            Replace it
                                                        </label>
                                                    )}
                                                </div>
                                                <div className="flex flex-col gap-1 pt-2">
                                                    <code className="break-all text-xs">{row.value}</code>
                                                    {row.description && (
                                                        <span className="text-xs text-muted-foreground">
                                                            {row.description}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            );
                        })}
                </DialogBody>

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="button"
                        disabled={selected.length === 0 || blocked}
                        onClick={() => {
                            onAdd(
                                section,
                                selected.map(row => ({ key: finalKey(prefix, row.key), value: row.value })),
                            );
                            onOpenChange(false);
                        }}
                    >
                        Add {selected.length} variable{selected.length === 1 ? "" : "s"}
                    </Button>
                </DialogFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

/** A target in the list: its name, its key, and its kind when it has one. */
function TargetOption({ target }: { target?: EnvLinkTarget }) {
    if (!target) {
        return null;
    }
    return (
        <span className="flex items-center gap-2">
            <span>{target.name}</span>
            <span className="text-xs text-muted-foreground">{target.key}</span>
            {target.category && (
                <Badge variant="outline">
                    {target.engine ? `${target.category} · ${target.engine}` : target.category}
                </Badge>
            )}
        </span>
    );
}

interface Props {
    open: boolean;
    initialSection: LinkSection;
    onOpenChange: (open: boolean) => void;
    /** The keys the form holds in a section. */
    existingKeys: (section: LinkSection) => Set<string>;
    /** Adds the rows; a key the section has is replaced in place. */
    onAdd: (section: LinkSection, vars: { key: string; value: string }[]) => void;
}
```

`link-app/index.ts`: `export * from "./link-app-dialog.com";` and `export * from "./link-app.utils";`.

Before writing the imports, check what `@/components/ui` exports: `grep -n "DialogFixedContent\|DialogFooter\|Tabs\b" src/components/ui/index.ts`. Import anything it lacks from its own file, as `Badge` is.

- [ ] **Step 8: The button.** In `env-vars.form.com.tsx`:
  - Add the prop `onLinkApp?: () => void`.
  - Build `extraActions` as a fragment: the Link App button when `onLinkApp` is set, then the existing Show Final Values button.

```tsx
    const extraActions =
        onShowFinalValues || onLinkApp ? (
            <div className="flex items-center gap-2">
                {onLinkApp && (
                    <Button
                        type="button"
                        variant="outline"
                        onClick={onLinkApp}
                        className="w-fit"
                    >
                        <Link2 className="size-4" />
                        Link App
                    </Button>
                )}
                {onShowFinalValues && (
                    <Button
                        type="button"
                        variant="outline"
                        onClick={onShowFinalValues}
                        className="w-fit"
                    >
                        Show Final Values
                    </Button>
                )}
            </div>
        ) : undefined;
```

Import `Link2` from `lucide-react`.

- [ ] **Step 9: Wire it into the app's form.** In `app-config-env-vars.form.com.tsx`:
  - **State.** Add `const [linkSection, setLinkSection] = useState<LinkSection | null>(null);`.
  - **Buttons.** Pass `onLinkApp={() => setLinkSection("buildtime")}` to the buildtime `EnvVarsBaseForm`, and `onLinkApp={() => setLinkSection("runtime")}` to the runtime one. Do not pass it to the shared one.
  - **Handlers.** Add:

```tsx
    function existingKeys(section: LinkSection): Set<string> {
        return new Set(methods.getValues(section).map(envVar => envVar.key.trim()));
    }

    function addLinkedVars(section: LinkSection, vars: { key: string; value: string }[]) {
        const current = [...methods.getValues(section)];
        for (const linked of vars) {
            const item = { key: linked.key, value: linked.value, isLiteral: false, isSystem: false, isReadOnly: false };
            const at = current.findIndex(envVar => envVar.key.trim() === linked.key);
            if (at >= 0) {
                current[at] = item;
            } else {
                current.push(item);
            }
        }
        methods.setValue(section, current, { shouldDirty: true });
        toast.success(`${vars.length} variable${vars.length === 1 ? "" : "s"} added - save to keep them`);
    }
```

  - **The dialog.** Render it beside `FinalEnvValuesDialog`:

```tsx
            <LinkAppDialog
                open={linkSection !== null}
                initialSection={linkSection ?? "runtime"}
                onOpenChange={open => {
                    if (!open) {
                        setLinkSection(null);
                    }
                }}
                existingKeys={existingKeys}
                onAdd={addLinkedVars}
            />
```

Import `LinkAppDialog` and `type LinkSection` from `../link-app`.

- [ ] **Step 10: Gates.** `npm run lint:ci` and `npm run build`. Expected: both clean. Run `npx eslint --fix` on the new files first.

- [ ] **Step 11: Commit.**

```bash
git add src/application/modules/projects
git commit -m "feat(apps): Link App suggests the env vars that connect to another app

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: Against the backend, and merge

- [ ] **Step 1: Run it.** Start this repository's backend on port 10099 (`HP_CONFIG_FILE=config/config.local.toml HP_ENV=development HP_HTTP_SERVER_PORT=10099 HP_RUN_MODE=app`). The user's backend is on 10000; leave it alone. Point the dashboard's dev server at it (`VITE_HP_API_PROXY_TARGET=http://localhost:10099 npx vite --port 4322`) and check:
  1. **API:** `link-targets` of an app leaves out the app and its previews.
  2. **API:** `link-suggestions` for a target of another env, or for the app itself, answers `ERR_APP_NOT_FOUND`.
  3. **Dialog:** "Link App" sits before "Show Final Values" in the runtime and build-time sections, not in the shared one.
  4. **Dialog:** choosing a database or webapp target lists its groups, with the recommended group selected.
  5. **Dialog:** a prefix changes the final names; two selected rows with one name, or a name the form has, disable Add until renamed or "Replace it" is ticked.
  6. **Dialog:** Add puts the rows in the form, unsaved. Save, then "Show Final Values", shows them resolved (masked).
  7. **Cleanup:** delete what the check created.
- [ ] **Step 2: Gates** in both repositories, as the Global Constraints list them.
- [ ] **Step 3: Merge.** Merge each repository's `feat/link-app-env-vars` into main with `--no-ff`, locally, then delete the branch.
