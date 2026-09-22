# Image Naming and Tagging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A built image is always called `<project key>-<app key>:<env key>-<commit sha7>`, the
operator cannot name it, and a deployment may add tags that are never remembered.

**Architecture:** The name and the tag become functions of the app, computed on the entity and
used by the one place that tags a build. `imageName` and `imageTags` leave the deployment
settings entirely; what one deployment asked for - its extra tags, and `NoCache` with them -
moves into that deployment's task arguments, and reaches a build on another node through a new
field on the agent's build request.

**Tech Stack:** Go 1.27, bun, testify, protobuf/gRPC for the build agent; React + TypeScript,
zod and react-hook-form on the dashboard.

**Spec:** [docs/superpowers/specs/2026-09-22-image-naming-and-tagging-design.md](../specs/2026-09-22-image-naming-and-tagging-design.md)

## Global Constraints

- Before calling a task done: `go build ./...`, `golangci-lint run ./...` over the **whole**
  repo (120-character lines, US spelling), `go test ./...`. Run `make gen-swag` in any task
  that changes a DTO and `make gen-proto` in the task that changes the proto;
  `docs/openapi/swagger.json` and the generated `*.pb.go` are committed.
- Repository name: `<project key>-<app key>`, each segment cut to **50** characters, and when
  any segment was cut the name ends with `-` plus the first **6** characters of a sha256 of the
  app's global key.
- Tag: `<env key>-<commit sha7>`. The environment part is cut to **20** characters. Every
  environment is prefixed, production included.
- A deployment may add at most **5** tags. Each is matched against the OCI tag grammar
  `^[A-Za-z0-9_][A-Za-z0-9._-]{0,127}$` and is at most **100** characters before the
  environment prefix is added.
- Limits: 255 for the whole repository name including `<address>/<username>/`, 128 for a tag.
- The commit reference stays first in the list, because the service spec runs `ImageTags[0]`.
- A value that belongs to one deployment lives in that deployment's **task arguments**, never
  in the app's deployment settings. `NoCache` is moved there by this plan for the same reason.
- Only references containing `/` are pushed, which is unchanged.
- End every commit message with `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`.

---

## File Structure

**Backend**

| File | Change |
|---|---|
| `hivepaas_app/entity/app_image.go` (new) | `ImageRepoName`, `ImageTag`, `ImageTagPrefix` and the normalize/truncate rules |
| `hivepaas_app/entity/app.go` | `GetAutoImageName` removed |
| `hivepaas_app/base/image.go` | `ImageNameMaxLen` replaced by `ImageRepoNameMaxLen`, `ImageTagMaxLen`, `ImageCustomTagMaxLen`, `ImageMaxCustomTags` |
| `hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/helper.go` | `calcBuildImageTags` becomes a pure `buildImageReferences` |
| `hivepaas_app/service/imagebuildservice/types.go` | `ImageBuildReq.ImageTags` added, `ImageName` removed |
| `hivepaas_app/entity/app_deployment_settings.go` | `ImageName` and `ImageTags` removed from `DeploymentRepoSource`, `NoCache` removed from `AppDeploymentSettings` |
| `hivepaas_app/entity/app_deployment_settings_migration.go` | version 2: drops all three stored fields |
| `hivepaas_app/entity/task_app_deploy.go` | `NoCache` and `ImageTags` on the deploy task's arguments |
| `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_settings_{get,update}.go` | fields removed; `image` block added to the response |
| `hivepaas_app/usecase/appactionuc/appactiondto/deploy.go` | `ImageTags` moves up to the request, validated, no longer applied to the setting |
| `hivepaas_app/usecase/appactionuc/deploy.go` | tags written onto the deployment only |
| `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/repo_deploy_build_image.go` | passes the tags into the build request |
| `hivepaas_app/interface/agent/proto/image_build.proto` + client and server mapping | `image_tags` field |

**Dashboard**

| File | Change |
|---|---|
| `.../deployment-settings/app-deployment-settings.api.contracts.ts`, `.api.validator.ts`, `.api.ts` | `imageName`/`imageTags` out, `image` block in |
| `.../domain/apps/deployment-settings/app-deployment-settings.entity.ts` | same |
| `.../deployment-settings/schemas/app-config-deployment-settings.schema.ts` | both fields out |
| `.../deployment-settings/form/app-config-deployment-settings.form.com.tsx`, `route/...route.com.tsx` | both fields out |
| `.../deployment-settings/building-blocks/build-configuration-fields.com.tsx` | input replaced by the read-only line and the note |
| `.../deployment-settings/building-blocks/registry-repo-note.com.tsx` (new) | the note and the address rules |

---

### Task 1: The name and the tag

**Files:**
- Create: `hivepaas_app/entity/app_image.go`
- Create: `hivepaas_app/entity/app_image_test.go`
- Modify: `hivepaas_app/base/image.go`

**Deviation taken while executing:** `GetAutoImageName` stays until Task 2. Its only caller is
the build helper that Task 2 rewrites, and removing it here would leave the tree unbuildable
between two tasks that are each supposed to end green.

**Interfaces:**
- Produces: `(*entity.App).ImageRepoName() (string, error)`,
  `(*entity.App).ImageTag(commitHash string) (string, error)`,
  `(*entity.App).ImageTagPrefix() (string, error)`;
  `base.ImageRepoNameMaxLen = 255`, `base.ImageTagMaxLen = 128`,
  `base.ImageCustomTagMaxLen = 100`, `base.ImageMaxCustomTags = 5`.

- [ ] **Step 1: Write the failing test**

`hivepaas_app/entity/app_image_test.go`:

```go
package entity_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func appIn(projectKey, envKey, appKey string) *entity.App {
	return &entity.App{
		Key:        appKey,
		GlobalKey:  projectKey + "/" + envKey + "/" + appKey,
		Project:    &entity.Project{Key: projectKey},
		ProjectEnv: &entity.ProjectEnv{Key: envKey},
	}
}

func TestImageRepoNameJoinsProjectAndApp(t *testing.T) {
	name, err := appIn("shop", "dev", "api").ImageRepoName()

	assert.NoError(t, err)
	assert.Equal(t, "shop-api", name)
}

// The environment is in the tag, so one app is one repository however many
// environments it runs in.
func TestImageRepoNameIsTheSameInEveryEnvironment(t *testing.T) {
	dev, err := appIn("shop", "dev", "api").ImageRepoName()
	assert.NoError(t, err)
	prod, err := appIn("shop", "prod", "api").ImageRepoName()
	assert.NoError(t, err)

	assert.Equal(t, dev, prod)
}

// The same app key in two projects was one repository before this change, which
// mixed two projects' images and let a busy one evict a quiet one's.
func TestImageRepoNameSeparatesProjects(t *testing.T) {
	first, _ := appIn("shop", "dev", "api").ImageRepoName()
	second, _ := appIn("blog", "dev", "api").ImageRepoName()

	assert.NotEqual(t, first, second)
}

func TestImageRepoNameNeedsItsProject(t *testing.T) {
	app := &entity.App{Key: "api", ProjectEnv: &entity.ProjectEnv{Key: "dev"}}

	_, err := app.ImageRepoName()
	assert.Error(t, err)
}

// A project and an app may each be named with 100 characters, and the address
// and the account are still to go in front of the result.
func TestImageRepoNameTruncatesAndStaysUnique(t *testing.T) {
	long := strings.Repeat("a", 80)
	first, err := appIn(long, "dev", "one-"+strings.Repeat("b", 80)).ImageRepoName()
	assert.NoError(t, err)
	second, err := appIn(long, "dev", "one-"+strings.Repeat("c", 80)).ImageRepoName()
	assert.NoError(t, err)

	assert.LessOrEqual(t, len(first), 110)
	assert.NotEqual(t, first, second, "two cut names must not collide")
}

func TestImageRepoNameIsValidOCI(t *testing.T) {
	name, err := appIn("shop__one", "dev", "api--two").ImageRepoName()

	assert.NoError(t, err)
	assert.Equal(t, "shop_one-api-two", name)
	assert.NotContains(t, name, "__")
	assert.NotContains(t, name, "--")
	assert.False(t, strings.HasPrefix(name, "-"))
	assert.False(t, strings.HasSuffix(name, "-"))
}

func TestImageTagCarriesTheEnvironment(t *testing.T) {
	tag, err := appIn("shop", "dev", "api").ImageTag("9f3c1de0ab")

	assert.NoError(t, err)
	assert.Equal(t, "dev-9f3c1de", tag)
}

// Production is prefixed like everything else: nothing marks an environment as
// production, so a bare tag would be a convention only some installations get.
func TestImageTagPrefixesProductionToo(t *testing.T) {
	tag, err := appIn("shop", "prod", "api").ImageTag("9f3c1de0ab")

	assert.NoError(t, err)
	assert.Equal(t, "prod-9f3c1de", tag)
}

func TestImageTagPrefixIsCutToTwentyCharacters(t *testing.T) {
	prefix, err := appIn("shop", strings.Repeat("e", 40), "api").ImageTagPrefix()

	assert.NoError(t, err)
	assert.Len(t, prefix, 20)
}

func TestImageTagRefusesAShortCommit(t *testing.T) {
	_, err := appIn("shop", "dev", "api").ImageTag("9f3")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/entity/ -run TestImage -v`
Expected: FAIL - `app.ImageRepoName undefined`.

- [ ] **Step 3: Write the entity code**

`hivepaas_app/entity/app_image.go`:

```go
package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// repoSegmentMaxLen is what one of the two parts of a repository name may
	// hold. A project and an app may each be named with 100 characters, and
	// <address>/<username>/ still goes in front of the result.
	repoSegmentMaxLen = 50
	// envSegmentMaxLen leaves room in a 128-character tag for a commit, or for a
	// custom tag of the length base.ImageCustomTagMaxLen allows.
	envSegmentMaxLen = 20
	// truncatedSuffixLen is how much of the global key's digest is appended when
	// a segment was cut, so that two long names that share a prefix stay apart.
	truncatedSuffixLen = 6
	commitHashLen      = 7
)

// ImageRepoName is the repository a build is pushed to: one per app, shared by
// the app's environments, unique across the installation.
//
// It needs Project loaded. Both build paths load it - the deployment helper and
// the agent's own loader - and an app without it is a programming error rather
// than a name to guess at, so it is an error and not a shorter name.
func (app *App) ImageRepoName() (string, error) {
	if app.Project == nil || app.Project.Key == "" {
		return "", hperrors.Wrap(hperrors.ErrMissing).
			WithMsgLog("app %v has no project loaded, so its image has no name", app.ID)
	}

	project, projectCut := cutSegment(app.Project.Key, repoSegmentMaxLen)
	appPart, appCut := cutSegment(app.Key, repoSegmentMaxLen)

	name := normalizeImagePart(project + "-" + appPart)
	if projectCut || appCut {
		name += "-" + digestOf(app.GlobalKey)
	}
	return name, nil
}

// ImageTagPrefix is the environment part every tag of this app carries, the
// commit's and any the deployment added.
func (app *App) ImageTagPrefix() (string, error) {
	if app.ProjectEnv == nil || app.ProjectEnv.Key == "" {
		return "", hperrors.Wrap(hperrors.ErrMissing).
			WithMsgLog("app %v has no environment loaded, so its image has no tag", app.ID)
	}

	env, _ := cutSegment(app.ProjectEnv.Key, envSegmentMaxLen)
	return normalizeImagePart(env), nil
}

// ImageTag is what one build is tagged with.
func (app *App) ImageTag(commitHash string) (string, error) {
	if len(commitHash) < commitHashLen {
		return "", hperrors.Wrap(hperrors.ErrValueInvalid).
			WithMsgLog("commit hash %q is too short to tag an image with", commitHash)
	}

	prefix, err := app.ImageTagPrefix()
	if err != nil {
		return "", err
	}
	return prefix + "-" + commitHash[:commitHashLen], nil
}

// cutSegment shortens one part of a name and says whether it had to.
func cutSegment(value string, limit int) (string, bool) {
	if len(value) <= limit {
		return value, false
	}
	return value[:limit], true
}

// normalizeImagePart makes a joined name or tag legal: the OCI grammar takes a
// separator between alphanumerics and nothing else, so a doubled or dangling one
// has to go. Keys arrive slugged, and this is what handles the ones that were
// slugged under older rules.
func normalizeImagePart(value string) string {
	value = strings.NewReplacer("__", "_", "--", "-").Replace(value)
	for strings.Contains(value, "__") || strings.Contains(value, "--") {
		value = strings.NewReplacer("__", "_", "--", "-").Replace(value)
	}
	return strings.Trim(value, "-_.")
}

func digestOf(globalKey string) string {
	sum := sha256.Sum256([]byte(globalKey))
	return hex.EncodeToString(sum[:])[:truncatedSuffixLen]
}

// ImageReference is what a build pushes: the repository under the account of the
// registry it is being pushed to, and one tag.
func ImageReference(address, username, repoName, tag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", address, username, repoName, tag)
}

var _ = base.ImageRepoNameMaxLen
```

Replace `hivepaas_app/base/image.go`:

```go
package base

const (
	// ImageRepoNameMaxLen is the OCI limit for a whole repository name, which
	// <address>/<username>/<name> has to fit inside.
	ImageRepoNameMaxLen = 255
	// ImageTagMaxLen is the OCI limit for a tag.
	ImageTagMaxLen = 128
	// ImageCustomTagMaxLen is what a deployment's own tag may hold before the
	// environment prefix is added to it.
	ImageCustomTagMaxLen = 100
	// ImageMaxCustomTags is how many a deployment may add. A release marker or
	// two is the case this exists for.
	ImageMaxCustomTags = 5
)
```

Delete the `var _ = base.ImageRepoNameMaxLen` line once something else in the package uses the
constants; it is there only to keep the import while the file is written.

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/entity/ -run TestImage -v`
Expected: PASS, all ten.

- [ ] **Step 5: Build, lint and commit**

```bash
go build ./... && golangci-lint run ./... && go test ./hivepaas_app/entity/...
git add hivepaas_app/entity hivepaas_app/base/image.go
git commit -m "$(cat <<'EOF'
feat(image): name an image after its project and app, tag it with its environment

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: The build uses them

**Files:**
- Modify: `hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/helper.go`
- Modify: `hivepaas_app/service/imagebuildservice/types.go`
- Create: `hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/helper_test.go`

**Interfaces:**
- Consumes: `(*entity.App).ImageRepoName`, `ImageTag`, `ImageTagPrefix` (Task 1).
- Produces: `buildImageReferences(app *entity.App, commitHash string, extraTags []string,
  regAuth *entity.RegistryAuth) ([]string, error)`; `ImageBuildReq.ImageTags []string` replacing
  `ImageBuildReq.ImageName`.

- [ ] **Step 1: Write the failing test**

`helper_test.go`:

```go
package imagebuildserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func buildApp() *entity.App {
	return &entity.App{
		ID: "app-1", Key: "api", GlobalKey: "shop/dev/api",
		Project:    &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Key: "dev"},
	}
}

func registryAuth() *entity.RegistryAuth {
	return &entity.RegistryAuth{Address: "registry.example.com", Username: "hivepaas"}
}

func TestReferencesWithARegistry(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", nil, registryAuth())

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"registry.example.com/hivepaas/shop-api:dev-9f3c1de",
		"shop-api:dev-9f3c1de",
	}, refs)
}

// The service spec runs ImageTags[0], so the commit reference has to stay first
// however many tags the deployment added.
func TestReferencesPutTheCommitFirst(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", []string{"v1.4.0", "stable"},
		registryAuth())

	assert.NoError(t, err)
	assert.Equal(t, "registry.example.com/hivepaas/shop-api:dev-9f3c1de", refs[0])
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:dev-v1.4.0")
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:dev-stable")
}

// Two environments share one repository, so a custom tag of one must not
// overwrite the other's.
func TestReferencesPrefixCustomTagsWithTheEnvironment(t *testing.T) {
	app := buildApp()
	app.ProjectEnv = &entity.ProjectEnv{Key: "prod"}

	refs, err := buildImageReferences(app, "9f3c1de0ab", []string{"v1.4.0"}, registryAuth())

	assert.NoError(t, err)
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:prod-v1.4.0")
	assert.NotContains(t, refs, "registry.example.com/hivepaas/shop-api:v1.4.0")
}

// Without a registry nothing can be pushed, so only the local reference is built
// - which is what a single-node installation runs.
func TestReferencesWithoutARegistry(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", []string{"v1.4.0"}, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"shop-api:dev-9f3c1de", "shop-api:dev-v1.4.0"}, refs)
}

func TestReferencesNeedAnApp(t *testing.T) {
	_, err := buildImageReferences(&entity.App{Key: "api"}, "9f3c1de0ab", nil, registryAuth())
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/service/imagebuildservice/... -run TestReferences -v`
Expected: FAIL - `undefined: buildImageReferences`.

- [ ] **Step 3: Rewrite the helper**

`helper.go` replaces `calcBuildImageTags` entirely:

```go
package imagebuildserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// buildImageReferences is every name the built image is tagged with, in the order
// they are applied.
//
// The first is what the service spec runs, so it is the commit's and it is the
// one carrying the registry. The local reference is last and is never pushed:
// push skips anything without a "/", which is how a build on a node keeps a name
// it can run from without sending it anywhere.
func buildImageReferences(
	app *entity.App,
	commitHash string,
	extraTags []string,
	regAuth *entity.RegistryAuth,
) ([]string, error) {
	repoName, err := app.ImageRepoName()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	commitTag, err := app.ImageTag(commitHash)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	prefix, err := app.ImageTagPrefix()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	tags := make([]string, 0, len(extraTags)+1)
	tags = append(tags, commitTag)
	for _, extra := range extraTags {
		tags = append(tags, prefix+"-"+extra)
	}

	refs := make([]string, 0, len(tags)*2) //nolint:mnd // the registry's and the local one
	if regAuth != nil {
		for _, tag := range tags {
			refs = append(refs, entity.ImageReference(regAuth.Address, regAuth.Username, repoName, tag))
		}
	}
	for _, tag := range tags {
		refs = append(refs, repoName+":"+tag)
	}
	return refs, nil
}
```

In `build_image.go`, the call becomes:

```go
	var regAuth *entity.RegistryAuth
	if data.PushToRegistry.ID != "" {
		regAuthSetting := data.RefObjects.RefSettings[data.PushToRegistry.ID]
		if regAuthSetting == nil {
			return hperrors.NewMissing("Registry auth to push image")
		}
		regAuth = regAuthSetting.MustAsRegistryAuth()
	}
	data.ImageTags, err = buildImageReferences(data.App, data.CommitHash, data.ImageTags, regAuth)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Resp.ImageTags = data.ImageTags
```

Remove `GetAutoImageName` from `hivepaas_app/entity/app.go`: this task's rewrite takes its last
caller away.

In `types.go`, `ImageName string` becomes `ImageTags []string`, with the comment:

```go
	// ImageTags are the tags this deployment asked for, without the environment
	// prefix. They are not stored anywhere: a release marker belongs to one build.
	ImageTags []string
```

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/service/imagebuildservice/... -v`
Expected: PASS, all five plus whatever the package already had.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/imagebuildservice
git commit -m "$(cat <<'EOF'
feat(image): build every reference from the app rather than from settings

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: The fields leave the settings

**Files:**
- Modify: `hivepaas_app/entity/app_deployment_settings.go`,
  `hivepaas_app/entity/app_deployment_settings_migration.go`
- Modify: `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_settings_update.go`,
  `.../deployment_settings_get.go`
- Create: `hivepaas_app/entity/app_deployment_settings_migration_test.go`

**Interfaces:**
- Produces: `CurrentAppDeploymentSettingsVersion = 2`;
  `DeploymentSettingsResp.Image *DeploymentImageNamingResp{RepoName, TagPrefix string}`.
- `AppDeploymentSettings` loses `NoCache` as well: it is a value one deployment asked for, and
  Task 4 moves it to the task's arguments with the tags.

- [ ] **Step 1: Write the failing test**

`app_deployment_settings_migration_test.go`:

```go
package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A row written before this change carries a name and tags that nothing reads any
// more. The migration drops them, so that a spec bundle exported afterwards does
// not carry them either.
func TestDeploymentSettingsMigrationDropsNamingFields(t *testing.T) {
	stored := &entity.Setting{
		ID:      "s1",
		Type:    base.SettingTypeAppDeployment,
		Version: 1,
		Data: `{
			"activeMethod": "repo",
			"repoSource": {
				"repoType": "git", "repoURL": "https://example.com/x.git", "repoRef": "main",
				"imageName": "custom-name", "imageTags": ["api:latest"]
			},
			"noCache": true
		}`,
	}

	data, err := stored.AsAppDeploymentSettings()
	assert.NoError(t, err)

	changed, err := data.Migrate(stored)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, entity.CurrentAppDeploymentSettingsVersion, stored.Version)

	assert.NotContains(t, stored.Data, "imageName")
	assert.NotContains(t, stored.Data, "custom-name")
	assert.NotContains(t, stored.Data, "api:latest")
	assert.NotContains(t, stored.Data, "noCache")
	assert.Contains(t, stored.Data, "repoRef")
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/entity/ -run TestDeploymentSettingsMigration -v`
Expected: FAIL - the data still contains `imageName`, because version 1 is current and the
migration returns early.

- [ ] **Step 3: Remove the fields and write the migration**

In `app_deployment_settings.go`: `CurrentAppDeploymentSettingsVersion = 2`; delete
`ImageName` and `ImageTags` from `DeploymentRepoSource`, and delete `NoCache` from
`AppDeploymentSettings`.

`NoCache` goes for the same reason the tags do. It is written onto the deployment's copy of the
settings after the app's own setting has been serialized - which works, but means a struct
whose fields are sometimes configuration and sometimes one deployment's argument, and nothing
in the type says which is which. Task 4 gives both a place that says so.

In `app_deployment_settings_migration.go`:

```go
	// Version 2 removed repoSource.imageName, repoSource.imageTags and noCache.
	// The name is a function of the app now, the tags were configuration no build
	// ever read, and noCache belongs to one deployment rather than to the app.
	// All three are gone from the struct, so parsing has already dropped them -
	// writing the data back is what takes them out of the row.
```

(The unmarshal into the new struct drops unknown fields, so `MustSetData` at the end of
`Migrate` is what persists their removal. No field-by-field code is needed.)

In `deployment_settings_update.go`, delete `ImageName`, `ImageTags` from
`DeploymentRepoSourceReq`, their two lines in `ToEntity`, and the `imageName` validator; delete
the now unused `imageNameMaxLen` constant.

In `deployment_settings_get.go`, delete both fields from `DeploymentRepoSourceResp` and add to
`DeploymentSettingsResp`:

```go
	// Image is what a build of this app will be called. The dashboard shows it
	// and composes the full reference from the registry the form has selected,
	// so that the naming rules live in one place - here.
	Image *DeploymentImageNamingResp `json:"image"`
```

```go
type DeploymentImageNamingResp struct {
	RepoName  string `json:"repoName"`
	TagPrefix string `json:"tagPrefix"`
}
```

filled in the transform from `app.ImageRepoName()` and `app.ImageTagPrefix()`, left nil when
either returns an error.

- [ ] **Step 4: Run the tests and regenerate the API document**

```bash
go test ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/... -v
make gen-swag
```
Expected: PASS, and `swagger.json` loses both fields and gains the `image` block.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app docs/openapi
git commit -m "$(cat <<'EOF'
feat(image): take the image name and tags out of the deployment settings

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: What one deployment asked for lives in its task

**Files:**
- Modify: `hivepaas_app/entity/task_app_deploy.go`
- Modify: `hivepaas_app/service/appdeploymentservice/service.go`,
  `.../appdeploymentserviceimpl/deployment_create.go`, `.../deployment.go`,
  `.../repo_deploy_build_image.go:51`, `.../repo_deploy_checkout_src.go:27`
- Modify the three other callers of `CreateDeploymentAndTask`:
  `hivepaas_app/service/apppreviewservice/apppreviewserviceimpl/create_preview.go:268`,
  `hivepaas_app/service/appprovisionservice/appprovisionserviceimpl/provision.go:150`,
  `hivepaas_app/usecase/webhookuc/app_deployment.go:77`
- Modify: `hivepaas_app/usecase/appactionuc/appactiondto/deploy.go`,
  `hivepaas_app/usecase/appactionuc/deploy.go:190-197`
- Create: `hivepaas_app/usecase/appactionuc/appactiondto/deploy_test.go`,
  `hivepaas_app/entity/task_app_deploy_test.go`

**Interfaces:**
- Consumes: `ImageBuildReq.ImageTags` (Task 2).
- Produces: `entity.TaskAppDeployArgs{Deployment ObjectID, NoCache bool, ImageTags []string}`;
  `appdeploymentservice.DeploymentArgs{NoCache bool, ImageTags []string}`;
  `CreateDeploymentAndTask(app *entity.App, deploymentSettings *entity.AppDeploymentSettings,
  args DeploymentArgs) (*entity.Deployment, *entity.Task, error)`;
  `appDeploymentData.DeployArgs *entity.TaskAppDeployArgs`; `DeployAppReq.ImageTags []string`.

- [ ] **Step 1: Write the failing tests**

`hivepaas_app/entity/task_app_deploy_test.go`:

```go
package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// What one deployment asked for travels with its task rather than with the app's
// settings, so it has to survive being written and read back.
func TestDeployArgsCarryWhatOneDeploymentAskedFor(t *testing.T) {
	task := &entity.Task{ID: "t1"}
	task.MustSetArgs(&entity.TaskAppDeployArgs{
		Deployment: entity.ObjectID{ID: "d1"},
		NoCache:    true,
		ImageTags:  []string{"v1.4.0"},
	})

	// Read it the way the executor does: from the stored string, not the cache.
	stored := &entity.Task{ID: task.ID, Args: task.Args}
	args, err := stored.ArgsAsAppDeploy()

	assert.NoError(t, err)
	assert.Equal(t, "d1", args.Deployment.ID)
	assert.True(t, args.NoCache)
	assert.Equal(t, []string{"v1.4.0"}, args.ImageTags)
}
```

`hivepaas_app/usecase/appactionuc/appactiondto/deploy_test.go`:

```go
package appactiondto_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

func reqWithTags(tags ...string) *appactiondto.DeployAppReq {
	req := appactiondto.NewDeployAppReq()
	req.ImageTags = tags
	return req
}

func TestDeployAcceptsPlainTags(t *testing.T) {
	assert.Empty(t, reqWithTags("v1.4.0", "stable").Validate())
}

// A tag is a tag, not a reference: the repository is the app's, and the
// environment prefix is added by the build.
func TestDeployRefusesReferencesAndPaths(t *testing.T) {
	for _, tag := range []string{"api:v1", "team/api:v1", "", "-leading", strings.Repeat("x", 101)} {
		assert.NotEmpty(t, reqWithTags(tag).Validate(), "tag %q must be refused", tag)
	}
}

func TestDeployRefusesTooManyTags(t *testing.T) {
	assert.NotEmpty(t, reqWithTags("a", "b", "c", "d", "e", "f").Validate())
}
```

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/usecase/appactionuc/... -run 'TestDeploy' -v`
Expected: FAIL - `args.NoCache undefined` and `req.ImageTags undefined`.

- [ ] **Step 3: Put both values on the task's arguments**

`hivepaas_app/entity/task_app_deploy.go`:

```go
type TaskAppDeployArgs struct {
	Deployment ObjectID `json:"deployment"`

	// NoCache and ImageTags are what this one deployment asked for. They are here
	// rather than in the app's deployment settings because that is what they are:
	// arguments to one run, not configuration of the app.
	NoCache bool `json:"noCache,omitempty"`
	// ImageTags carry no environment prefix; the build adds it.
	ImageTags []string `json:"imageTags,omitempty"`
}
```

`hivepaas_app/service/appdeploymentservice/service.go`:

```go
// DeploymentArgs is what one deployment asked for, as opposed to what the app is
// configured with.
type DeploymentArgs struct {
	NoCache   bool
	ImageTags []string
}

	CreateDeploymentAndTask(app *entity.App, deploymentSettings *entity.AppDeploymentSettings,
		args DeploymentArgs) (*entity.Deployment, *entity.Task, error)
```

`deployment_create.go` takes the third parameter and writes it through:

```go
	err := deploymentTask.SetArgs(&entity.TaskAppDeployArgs{
		Deployment: entity.ObjectID{ID: deployment.ID},
		NoCache:    args.NoCache,
		ImageTags:  args.ImageTags,
	})
```

The three callers that ask for nothing special pass an empty value:

```go
	// create_preview.go, provision.go, app_deployment.go
	..., appdeploymentservice.DeploymentArgs{})
```

`hivepaas_app/usecase/appactionuc/deploy.go` replaces the two lines that mutated the settings:

```go
	deployment, deploymentTask, err := uc.appDeploymentService.CreateDeploymentAndTask(
		app, data.NewDeploymentSettings,
		appdeploymentservice.DeploymentArgs{NoCache: req.NoCache, ImageTags: req.ImageTags},
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
```

(The `deployment.Settings.NoCache = req.NoCache` line goes with them.)

`deployment.go` keeps the parsed arguments on the data it passes around:

```go
type appDeploymentData struct {
	*appdeploymentservice.AppDeploymentReq
	App                *entity.App
	Deployment         *entity.Deployment
	DeployArgs         *entity.TaskAppDeployArgs
	DeploymentCanceled bool
	Step               string
	NotifMsgData       *notificationservice.TemplateDataAppDeployment
}
```

assigned in `loadDeploymentData`, where the arguments are already parsed:

```go
	args, err := task.ArgsAsAppDeploy()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.DeployArgs = args
```

and the two readers use it:

```go
	// repo_deploy_checkout_src.go and repo_deploy_build_image.go
	if data.DeployArgs.NoCache || (data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache) {
```

with the build request carrying the tags:

```go
		ImageTags: data.DeployArgs.ImageTags,
```

- [ ] **Step 4: Validate the tags in the request**

In `appactiondto/deploy.go`, delete `ImageTags` from `DeploymentRepoSourceReq` and the
`setting.ImageTags = req.ImageTags` line in its `ApplyTo` - that line is what made a one-off tag
permanent. Add to `DeployAppReq`:

```go
	// ImageTags are extra tags for this deployment only, without the environment
	// prefix, which the build adds. Nothing stores them.
	ImageTags []string `json:"imageTags"`
```

and validate with the helpers this repo already uses:

```go
// imageTagPattern is the OCI tag grammar, which has no place for a "/" or a ":",
// so a reference such as "api:v1" is refused here rather than pushed somewhere
// surprising.
var imageTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)

func (req *DeployAppReq) validateImageTags() (res []vld.Validator) {
	res = append(res, vld.SliceLen(&req.ImageTags, 0, base.ImageMaxCustomTags).OnError(
		vld.SetField("imageTags", nil),
	))
	for i := range req.ImageTags {
		field := fmt.Sprintf("imageTags[%d]", i)
		res = append(res, vld.StrLen(&req.ImageTags[i], 1, base.ImageCustomTagMaxLen).OnError(
			vld.SetField(field, nil),
		))
		res = append(res, vld.StrMatch(&req.ImageTags[i], imageTagPattern).OnError(
			vld.SetField(field, nil),
		))
	}
	return res
}
```

called from `DeployAppReq.Validate`.

- [ ] **Step 5: Run the tests**

Run: `go test ./hivepaas_app/entity/ ./hivepaas_app/usecase/appactionuc/... ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/apppreviewservice/... -v`
Expected: PASS.

- [ ] **Step 6: Regenerate and commit**

```bash
make gen-swag && go build ./... && golangci-lint run ./...
git add hivepaas_app docs/openapi
git commit -m "$(cat <<'EOF'
feat(image): give one deployment's arguments a place of their own

Extra image tags, and NoCache with them, move out of the app's deployment
settings and into the deploy task's arguments. They were being written onto the
deployment's copy of the settings after the app's own setting had been
serialized, which worked but left a struct whose fields were sometimes
configuration and sometimes one run's argument.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Tags reach a build on another node

**Files:**
- Modify: `hivepaas_app/interface/agent/proto/image_build.proto`
- Modify: `hivepaas_app/interface/agent/client/imagebuildservice/service.go:99`,
  `hivepaas_app/interface/agent/server/imagebuildservice/image_build.go:69`
- Create: `hivepaas_app/interface/agent/server/imagebuildservice/image_build_mapping_test.go`

**Interfaces:**
- Consumes: `ImageBuildReq.ImageTags` (Task 2).
- Produces: `ImageBuildReq.image_tags` field 12 in the proto, carried both ways.

- [ ] **Step 1: Write the failing test**

`image_build_mapping_test.go`:

```go
package imagebuildservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// A build that runs on another node has to tag the image the same way, so the
// deployment's tags have to survive the wire.
func TestImageTagsSurviveTheProto(t *testing.T) {
	req := &proto.ImageBuildReq{
		TaskId: "t1", AppId: "a1", CommitHash: "9f3c1de0ab",
		ImageTags: []string{"v1.4.0", "stable"},
	}

	assert.Equal(t, []string{"v1.4.0", "stable"}, req.GetImageTags())
}
```

- [ ] **Step 2: Run the test to watch it fail**

Run: `go test ./hivepaas_app/interface/agent/... -run TestImageTags -v`
Expected: FAIL - `unknown field ImageTags in proto.ImageBuildReq`.

- [ ] **Step 3: Change the proto and regenerate**

In `image_build.proto`, replace `string image_name = 5;` with a reserved slot and add the new
field, so that an old agent talking to a new server cannot mistake one for the other:

```proto
message ImageBuildReq {
  string task_id = 1;
  string app_id = 2;
  string commit_hash = 3;
  DeploymentDockerfile dockerfile = 4;
  reserved 5;                  // was image_name, now a function of the app
  reserved "image_name";
  string push_to_registry_id = 6;
  ImageBuildSettings image_build_settings = 7;
  bool no_cache = 8;
  string build_id = 9;
  string checkout_dir = 10;
  string temp_dir = 11;
  repeated string image_tags = 12;
}
```

Run: `make gen-proto`

Then the two mappings: in the client, `ImageName: req.ImageName` becomes
`ImageTags: req.ImageTags`; in the server, `ImageName: req.GetImageName()` becomes
`ImageTags: req.GetImageTags()`.

- [ ] **Step 4: Run the tests**

Run: `go test ./hivepaas_app/interface/agent/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
go build ./... && golangci-lint run ./... && go test ./...
git add hivepaas_app
git commit -m "$(cat <<'EOF'
feat(image): carry a deployment's tags to a build on another node

An agent older than the server ignores the new field and pushes only the commit
tag; the build log lists what was pushed, so that shows up rather than passing
silently.

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: The dashboard's data layer

**Files:**
- Modify: `src/application/modules/projects/api/services/project-apps-services/deployment-settings/app-deployment-settings.api.contracts.ts`,
  `.api.validator.ts`, `.api.ts`
- Modify: `src/application/modules/projects/domain/apps/deployment-settings/app-deployment-settings.entity.ts`

**Interfaces:**
- Consumes: the `image` block from Task 3.
- Produces: `AppDeploymentSettings["image"]` of type `{ repoName: string; tagPrefix: string } | null`;
  `imageName` and `imageTags` gone from the entity, the contracts and the payload.

- [ ] **Step 1: Remove both fields**

Delete `imageName` and `imageTags` from the entity, from the update payload type, from the
validator's schema, and delete `splitImageTags` from `app-deployment-settings.api.ts` along with
the line that called it.

- [ ] **Step 2: Add the image block**

In the validator's response schema:

```ts
    image: z
        .object({ repoName: z.string(), tagPrefix: z.string() })
        .nullish()
        .transform(value => value ?? null),
```

and the matching field on the entity:

```ts
    /** What a build of this app is called. The rules live on the server. */
    image: { repoName: string; tagPrefix: string } | null;
```

- [ ] **Step 3: Verify**

```bash
cd ../hivepaas-dashboard && npx tsc --noEmit && npm run lint
```
Expected: errors only where the form still reads the removed fields, which Task 7 fixes.

- [ ] **Step 4: Commit** (after Task 7, since the tree does not typecheck between them)

---

### Task 7: The dashboard's build section

**Files:**
- Modify: `.../deployment-settings/schemas/app-config-deployment-settings.schema.ts`,
  `.../form/app-config-deployment-settings.form.com.tsx`,
  `.../route/app-config-deployment-settings.route.com.tsx`,
  `.../building-blocks/build-configuration-fields.com.tsx`
- Create: `.../building-blocks/registry-repo-note.com.tsx`

**Interfaces:**
- Consumes: `AppDeploymentSettings["image"]` (Task 6), the registry auth list, which carries
  `address` and `username`.
- Produces: `RegistryRepoNote({ repoName, registryAuthId })`, and
  `registryCreatesRepositories(address: string): boolean`.

- [ ] **Step 1: Remove the input and the form fields**

Delete `imageName` and `imageTags` from the zod schema, from `toFormValues` in the form
component, from the payload in the route, and delete the *Image Repository Name* `InfoBlock`
and its `useController` from `build-configuration-fields.com.tsx`.

- [ ] **Step 2: Write the note**

`registry-repo-note.com.tsx`:

```tsx
/**
 * Whether pushing to this registry creates the repository.
 *
 * An address that is not recognized is treated as one that does not: being told
 * to check something that turns out to be true costs a glance, and staying quiet
 * costs a whole build, because the push is the last step.
 */
export function registryCreatesRepositories(address: string): boolean {
    const host = address.trim().toLowerCase().split("/")[0];

    if (host === "docker.io" || host === "index.docker.io" || host === "ghcr.io" || host === "quay.io") {
        return true;
    }
    return host.endsWith(".azurecr.io");
}
```

The component renders nothing when no registry is selected or when
`registryCreatesRepositories` is true, and otherwise:

```tsx
<div className={cn(dashedBorderBox)}>
    <span className="font-semibold text-orange-500">Create this repository first:</span>{" "}
    <span className="font-mono">{reference}</span>. This registry does not create repositories
    on push, so a build would fail at its last step.
</div>
```

- [ ] **Step 3: Show what the image is called**

In `build-configuration-fields.com.tsx`, where the input was:

```tsx
<InfoBlock
    titleWidth={220}
    title={
        <LabelWithInfo
            label="Image"
            content="What a build of this app is called. It is derived from the project and the app, so two apps can never share a repository."
        />
    }
>
    <div className="flex flex-col gap-1">
        <span className="font-mono text-sm">{image?.repoName ?? "—"}</span>
        {reference && (
            <div className="flex items-center gap-2">
                <span className="font-mono text-xs text-muted-foreground">{reference}</span>
                <button
                    type="button"
                    className="text-blue-500 cursor-pointer hover:underline select-none"
                    onClick={() => {
                        // The repo has no copy component; this is how the API key
                        // dialog does it.
                        void navigator.clipboard
                            .writeText(reference)
                            .then(() => {
                                toast.success("Image reference copied to clipboard");
                            })
                            .catch(() => {
                                toast.error("Failed to copy the image reference");
                            });
                    }}
                >
                    Copy
                </button>
            </div>
        )}
        <p className="text-xs text-muted-foreground">
            Tagged {image?.tagPrefix ?? "<env>"}-&lt;commit&gt; for this environment.
        </p>
    </div>
</InfoBlock>
<RegistryRepoNote
    repoName={image?.repoName ?? ""}
    registryAuthId={pushToRegistry.value?.id ?? ""}
/>
```

where `reference` is `address/username/repoName` composed from the registry auth the form
currently has selected - not the saved one, so that the note follows what the operator is about
to save.

- [ ] **Step 4: Verify**

```bash
npx tsc --noEmit && npm run lint && npm run build
```
Expected: clean.

- [ ] **Step 5: Commit both dashboard tasks**

```bash
git add src/application
git commit -m "$(cat <<'EOF'
feat(image): show what a build is called instead of asking for a name

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: End to end, on the development cluster

**Files:** none. What this produces is a green run or a bug to fix in the task that owns it.

**What ran, and what could not:** the settings side is verified against a running server - an
app's deployment settings answer with the computed name, a deploy carrying `["v1.4.0"]` put the
tags in the task's arguments, and the app's own setting gained neither them nor `noCache`. The
build side is not: the deployment task was executed by the other backend on this machine, the
one running main, so the image came out named the old way. Running the steps below needs the
instance that executes tasks to be built from this branch.

- [ ] **Step 1: Two projects, one app key**

Create `api` in two projects, build both with the system registry selected. Expect two
repositories, `<project-a>-api` and `<project-b>-api`, each with one tag.

- [ ] **Step 2: Two environments, one app**

Deploy the same app in `dev` and in another environment. Expect **one** repository with
`dev-<commit>` and `<other>-<commit>`.

- [ ] **Step 3: A deployment's own tags**

```bash
curl -su <key>:<secret> -X POST -H 'Content-Type: application/json' \
  -d '{"repoSource":{"repoRef":"main"},"imageTags":["v1.4.0"]}' \
  http://localhost:10000/_/projects/<projectID>/dev/apps/<appID>/deploy
```

Expect `dev-<commit>` and `dev-v1.4.0` on one image, the service running the commit one, and
the app's deployment settings unchanged afterwards - check with a GET that no tags are stored.

- [ ] **Step 4: The note**

In the dashboard, select a registry whose address is an ECR one and confirm the note appears
with the full reference; select the system registry and confirm it does not.

- [ ] **Step 5: A build on another node**

With a second node in the cluster, force the build onto it and confirm the tags are the same.
If the agent on that node is older, confirm the build log shows only the commit tag rather than
failing.

---

## Notes from the self-review

- **Spec coverage.** §3 is Task 1, §4 Task 3, §5 Tasks 2, 4 and 5, §6 Tasks 6 and 7, §7 needs
  no code, §8 is the tests inside each task plus Task 8.
- **One thing the spec leaves open** and this plan decides: where a deployment's tags live now
  that `DeploymentRepoSource` has none. They go into the deploy task's arguments, beside the
  deployment id - and `NoCache` moves there with them, since it was the existing example of a
  one-run value kept in a configuration struct. `CreateDeploymentAndTask` grows a third
  parameter, which its four callers pass an empty value to unless they have something to say.
- **Order matters between Tasks 6 and 7:** the dashboard does not typecheck between them, so
  they share a commit.
- **Not in this plan, on purpose:** retention by repository pattern, ECR's expiring token, and
  a check that the repository exists - all in §9 of the spec.
