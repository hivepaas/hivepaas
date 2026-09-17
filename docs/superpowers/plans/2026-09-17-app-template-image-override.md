# App Template Image Override and Registry Scan Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a user create an app from a template with a newer image of their own choosing - same repository, any tag but `latest` - after seeing which tags the registry actually publishes, and give the templates repository a tool that bumps those pins by itself.

**Architecture:** Three pure additions to `templatemodel` decide what an override is allowed to be and which of a repository's tags are worth offering; a small `services/registry` client reads tags over the standard V2 API; `apptemplateservice` joins the two behind one method with a short-lived cache; the create path applies the override to the rendered document only - never to the stored base - and records it on the app; `tools/apptemplate bump` reuses the same pieces to rewrite the templates repository.

**Tech Stack:** Go 1.27, `net/http` with the V2 bearer-token challenge flow, `pkg/imageref` for tag ordering, gin handlers, testify, `httptest` for a fake registry.

**Spec:** `docs/superpowers/specs/2026-09-17-app-template-image-override-design.md`

## Global Constraints

- **Layering** (`docs/ARCHITECTURE.md`): `handler → dto → usecase → service → repository → entity → base`. A service never imports anything under `hivepaas_app/usecase/`. `templatemodel` stays pure: no database, no network, no service imports.
- **Error codes** are declared in the task that first raises them - `tools/errcodelint` refuses a code nothing references - each with its English message in the matching `hivepaas_app/pkg/translation/messages/en/errors.*.en.toml`.
- **Before every commit:** `make fmt`, `go build ./...`, `make lint-local` (`make lint` does not work in the devtools container here), `go test ./...`.
- **Style:** 120-character lines, US spelling (`normalize`, not `normalise` - `misspell` checks comments too), no `//nolint:exhaustive` on this repo's enums.
- **`make gen-swag`** whenever a DTO changes; `docs/openapi/swagger.json` is committed.
- **The override rule, verbatim from the spec:** same repository as the template's pinned image; a tag is required and may not be `latest`; a digest is allowed only alongside a tag; a moving tag and a different major line are allowed but classified so the dashboard can warn differently.
- **The scan rule:** the repository is derived on the server from template + version + variant. No endpoint accepts an image or a URL from the client - that would turn it into a probe for the cluster's internal network.
- **Limits:** at most 5000 tags read from a registry, at most 50 returned, cache 10 minutes.
- **Deferred work** is marked `// TODO: app templates phase 2|later - <what>. See docs/superpowers/specs/2026-09-17-app-template-image-override-design.md §10.`
- **Commits:** the steps below commit per task, on a branch off `main`.

---

## File Structure

**New**

| path | responsibility |
|---|---|
| `hivepaas_app/service/apptemplateservice/templatemodel/image.go` | what an override may be, and which tags are candidates (Task 1) |
| `services/registry/reference.go` | docker reference → registry host + repository path (Task 2) |
| `services/registry/registry.go` | the V2 tags client: token challenge, pagination, limits (Task 2) |
| `hivepaas_app/hperrors/errors_registry.go` | `ERR_REGISTRY_*` (Task 2) |
| `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags.go` | `ImageTags`: template → image → tags → candidates (Task 3) |
| `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags_cache.go` | per-repository cache with a TTL constant (Task 3) |
| `hivepaas_app/usecase/apptemplateuc/image_tags.go` + `apptemplatedto/image_tags_get.go` | the scan endpoint (Task 5) |
| `hivepaas_app/interface/api/handler/apptemplatehandler/image_tags_get.go` | transport (Task 5) |
| `tools/apptemplate/bump.go` | the `bump` subcommand (Task 6) |
| `../app-templates/.github/workflows/bump.yml` | the scheduled pull request (Task 6) |

**Modified**

| path | change |
|---|---|
| `hivepaas_app/pkg/imageref/imageref.go` | export `CompareTags` (Task 1) |
| `hivepaas_app/hperrors/errors_app_template.go` + `errors.app_template.en.toml` | `ERR_APP_TEMPLATE_IMAGE_NOT_ALLOWED` (Task 1) |
| `hivepaas_app/service/apptemplateservice/{service,types}.go`, `apptemplateserviceimpl/service.go` | `ImageTags`, the registry client dependency (Task 3) |
| `hivepaas_app/service/apptemplateservice/templaterender/render.go` | apply the override to the document, never to the base (Task 4) |
| `hivepaas_app/entity/setting_app_template.go` | `ImageOverride` (Task 4) |
| `hivepaas_app/usecase/apptemplateuc/{create.go,audit.go}` + `apptemplatedto/{create,binding_get}.go` | carry, record and report the override (Tasks 4-5) |
| `hivepaas_app/interface/api/server/router_projects.go`, `registry/provides.go` | route and wiring (Tasks 3, 5) |

---

## Task 1: What an override may be - `templatemodel`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/templatemodel/image.go`
- Modify: `hivepaas_app/pkg/imageref/imageref.go`
- Modify: `hivepaas_app/hperrors/errors_app_template.go`
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml`
- Test: `hivepaas_app/service/apptemplateservice/templatemodel/image_test.go`
- Test: `hivepaas_app/pkg/imageref/imageref_test.go`

**Interfaces:**
- Consumes: `imageref.Parse`, `imageref.MajorVersion`, `imageref.IsUpgrade`; `IsPinnedImage` and the package-private `pinnedTagPattern` in `validate.go`.
- Produces:
  - `imageref.CompareTags(a, b string) (int, bool)`
  - `type ImageOverrideClass string` with `ImageOverrideSameLine`, `ImageOverrideOtherMajor`, `ImageOverrideMoving`, and `AllImageOverrideClasses`
  - `func ClassifyImageOverride(templateImage, override string) (ImageOverrideClass, error)`
  - `type TagCandidate struct { Tag string; Class ImageOverrideClass; Newer bool }`
  - `func SelectTagCandidates(templateImage string, tags []string, maxCandidates int) []TagCandidate`
  - `hperrors.ErrAppTemplateImageNotAllowed`

- [ ] **Step 1: Export the tag comparison**

`compareTags` already orders `18.6` against `18.7` and says when it cannot. Ordering candidates needs it from another package, so give it an exported name and keep the unexported one as the implementation.

In `hivepaas_app/pkg/imageref/imageref.go`, above `func compareTags`:

```go
// CompareTags orders two tags the way IsUpgrade does, reporting whether it could.
//
// It is exported for callers that sort a registry's tags rather than compare two
// references: sorting needs an ordering, and IsUpgrade answers a different
// question - whether one reference should replace another.
func CompareTags(a, b string) (int, bool) {
	return compareTags(a, b)
}
```

Append to `hivepaas_app/pkg/imageref/imageref_test.go`:

```go
func TestCompareTagsOrdersWhatItCan(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
		ok   bool
	}{
		{"18.6", "18.7", -1, true},
		{"18.7-alpine3.24", "18.6-alpine3.24", 1, true},
		{"18.6", "18.6", 0, true},
		{"18.6-alpine", "18.6", 0, false},
		{"stable", "18.6", 0, false},
	} {
		got, ok := CompareTags(tc.a, tc.b)
		assert.Equal(t, tc.ok, ok, "%s vs %s", tc.a, tc.b)
		if tc.ok {
			assert.Equal(t, tc.want, got, "%s vs %s", tc.a, tc.b)
		}
	}
}
```

- [ ] **Step 2: Declare the error**

Add to the `var (...)` block in `hivepaas_app/hperrors/errors_app_template.go`:

```go
	ErrAppTemplateImageNotAllowed = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_IMAGE_NOT_ALLOWED")
```

Add to `errors.app_template.en.toml`:

```toml
ERR_APP_TEMPLATE_IMAGE_NOT_ALLOWED = "Image '{{.Image}}' cannot be used here. The template deploys images from '{{.Repository}}'."
```

- [ ] **Step 3: Write the failing tests**

`hivepaas_app/service/apptemplateservice/templatemodel/image_test.go`:

```go
package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const pgImage = "postgres:18.6-alpine3.24"

func TestClassifyImageOverride(t *testing.T) {
	for name, tc := range map[string]struct {
		override string
		want     ImageOverrideClass
	}{
		"newer patch, same line":   {"postgres:18.7-alpine3.24", ImageOverrideSameLine},
		"newer base, same line":    {"postgres:18.7-alpine3.25", ImageOverrideSameLine},
		"another base entirely":    {"postgres:18.7-trixie", ImageOverrideSameLine},
		"older patch, same line":   {"postgres:18.4-alpine3.22", ImageOverrideSameLine},
		"next major line":          {"postgres:19.0-alpine3.24", ImageOverrideOtherMajor},
		"digest beside a tag":      {"postgres:18.7-alpine3.24@sha256:" + hex64, ImageOverrideSameLine},
		"major-only tag moves":     {"postgres:18", ImageOverrideMoving},
		"major and base tag moves": {"postgres:18-alpine", ImageOverrideMoving},
	} {
		t.Run(name, func(t *testing.T) {
			class, err := ClassifyImageOverride(pgImage, tc.override)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, class)
		})
	}
}

func TestClassifyImageOverrideRefuses(t *testing.T) {
	for name, override := range map[string]string{
		"another repository": "mariadb:11.8.9-noble",
		"another registry":   "ghcr.io/library/postgres:18.7",
		"latest":             "postgres:latest",
		"no tag at all":      "postgres",
		"digest on its own":  "postgres@sha256:" + hex64,
		"empty":              "",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ClassifyImageOverride(pgImage, override)
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateImageNotAllowed)
		})
	}
}

// A registry answers with everything it has ever published. What reaches the user
// has to be the builds they would recognize as another build of the same image.
func TestSelectTagCandidates(t *testing.T) {
	tags := []string{
		"latest", "18", "18-alpine", "alpine", "bookworm",
		"18.6-alpine3.24", "18.7-alpine3.24", "18.7-alpine3.25", "18.5-alpine3.22",
		"18.7-trixie", "19.0-alpine3.24", "19beta1-alpine3.24", "17.11-alpine3.24",
	}

	got := SelectTagCandidates(pgImage, tags, 0)

	assert.Equal(t, []TagCandidate{
		{Tag: "19.0-alpine3.24", Class: ImageOverrideOtherMajor, Newer: true},
		{Tag: "18.7-alpine3.25", Class: ImageOverrideSameLine, Newer: true},
		{Tag: "18.7-alpine3.24", Class: ImageOverrideSameLine, Newer: true},
		{Tag: "18.5-alpine3.22", Class: ImageOverrideSameLine, Newer: false},
		{Tag: "17.11-alpine3.24", Class: ImageOverrideOtherMajor, Newer: false},
	}, got)
}

func TestSelectTagCandidatesCapsAndDeduplicates(t *testing.T) {
	tags := []string{"18.7-alpine3.24", "18.7-alpine3.24", "18.8-alpine3.24", "18.9-alpine3.24"}

	got := SelectTagCandidates(pgImage, tags, 2)

	assert.Equal(t, []string{"18.9-alpine3.24", "18.8-alpine3.24"}, []string{got[0].Tag, got[1].Tag})
	assert.Len(t, got, 2)
}

func TestSelectTagCandidatesWithoutASuffixFamily(t *testing.T) {
	got := SelectTagCandidates("mariadb:11.8.9-noble", []string{"11.8.10-noble", "11.8.10-ubi", "latest"}, 0)

	assert.Equal(t, []TagCandidate{{Tag: "11.8.10-noble", Class: ImageOverrideSameLine, Newer: true}}, got)
}
```

Add the fixture beside `pgImage`:

```go
const hex64 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
```

- [ ] **Step 4: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ -run Image`
Expected: FAIL - `undefined: ClassifyImageOverride`.

- [ ] **Step 5: Write the image policy**

`hivepaas_app/service/apptemplateservice/templatemodel/image.go`:

```go
package templatemodel

import (
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

const latestTag = "latest"

// ImageOverrideClass says how far an image a user chose is from the one the
// template pinned. It decides what the dashboard warns about; it does not decide
// whether the override is allowed, which ClassifyImageOverride answers with an
// error instead.
type ImageOverrideClass string

const (
	// ImageOverrideSameLine is another release of the major line the template
	// describes. The template's configuration applies; HivePaaS has not tested
	// this exact release.
	ImageOverrideSameLine ImageOverrideClass = "same-line"
	// ImageOverrideOtherMajor is a line the template says nothing about, so the
	// overrides written for the declared lines may not apply - PostgreSQL 18 moved
	// its data directory, and a template that does not know about a line cannot
	// carry the mount change it needs.
	ImageOverrideOtherMajor ImageOverrideClass = "other-major"
	// ImageOverrideMoving is a tag that changes what it points at, so what runs
	// can change on a redeploy with nothing recorded to say it did.
	ImageOverrideMoving ImageOverrideClass = "moving"
)

var AllImageOverrideClasses = []ImageOverrideClass{
	ImageOverrideSameLine, ImageOverrideOtherMajor, ImageOverrideMoving,
}

// ClassifyImageOverride checks an image a user wants instead of the template's.
//
// The repository has to be the same one: everything else the template says - the
// environment variables, the health check, the mount path, the app kind - is only
// correct for that software, so another repository is not an override but a
// different app wearing this template's configuration.
//
// Within that repository the rule is deliberately loose, because the user is
// accepting the risk for one app of theirs. What is refused is only what cannot be
// recorded: a reference with no tag - a digest on its own included - has no version
// to show in the app's binding, and latest names whatever is newest today.
func ClassifyImageOverride(templateImage, override string) (ImageOverrideClass, error) {
	templateRef, overrideRef := imageref.Parse(templateImage), imageref.Parse(override)
	switch {
	case override == "" || overrideRef.Repository != templateRef.Repository:
		return "", imageNotAllowed(override, templateRef.Repository,
			"the image must come from the same repository as the template's")
	case overrideRef.Tag == "":
		return "", imageNotAllowed(override, templateRef.Repository,
			"name a tag: a reference without one has no version to record")
	case overrideRef.Tag == latestTag:
		return "", imageNotAllowed(override, templateRef.Repository,
			"latest names whatever is newest at the time, which is not a version")
	case !pinnedTagPattern.MatchString(overrideRef.Tag):
		return ImageOverrideMoving, nil
	}

	templateMajor, templateOK := imageref.MajorVersion(templateImage)
	overrideMajor, overrideOK := imageref.MajorVersion(override)
	if !templateOK || !overrideOK || templateMajor != overrideMajor {
		return ImageOverrideOtherMajor, nil
	}
	return ImageOverrideSameLine, nil
}

func imageNotAllowed(image, repository, reason string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateImageNotAllowed).
		WithParam("Image", image).WithParam("Repository", repository).
		WithExtraDetail("%s: %s", image, reason)
}

// TagCandidate is one tag a repository publishes that a user could move to.
type TagCandidate struct {
	Tag   string
	Class ImageOverrideClass
	Newer bool
}

// SelectTagCandidates picks, out of everything a repository has ever published,
// the tags a person would read as another build of the image in front of them.
//
// A registry answers with thousands of tags: every release, every base image,
// every alias. Three rules cut that down. A tag that does not name a release is
// dropped, because it moves. A tag from another base family is dropped - 18.7-trixie
// is not a newer build of an app running alpine, however much larger the number is.
// What remains is ordered newest first, so the top of the list is the answer to the
// question the user asked by opening it.
//
// maxCandidates of 0 means all of them.
func SelectTagCandidates(templateImage string, tags []string, maxCandidates int) []TagCandidate {
	current := imageref.Parse(templateImage)
	family := tagFamily(current.Tag)

	seen := map[string]bool{}
	candidates := make([]TagCandidate, 0, len(tags))
	for _, tag := range tags {
		if tag == "" || tag == current.Tag || seen[tag] || tagFamily(tag) != family {
			continue
		}
		seen[tag] = true

		candidate := current.Repository + ":" + tag
		class, err := ClassifyImageOverride(templateImage, candidate)
		if err != nil || class == ImageOverrideMoving {
			continue
		}
		newer, _ := imageref.IsUpgrade(templateImage, candidate)
		candidates = append(candidates, TagCandidate{Tag: tag, Class: class, Newer: newer})
	}

	slices.SortFunc(candidates, func(a, b TagCandidate) int {
		if order, ok := imageref.CompareTags(b.Tag, a.Tag); ok {
			return order
		}
		return strings.Compare(b.Tag, a.Tag)
	})
	if maxCandidates > 0 && len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	return candidates
}

// tagFamily is the non-numeric part of a tag's suffix: 18.6-alpine3.24 and
// 17.11-alpine3.22 are both "alpine", 11.8.9-noble is "noble", and 18.6 has no
// family at all. The numbers go because a base image moves on its own schedule -
// alpine3.24 succeeds alpine3.22 without the app's version changing.
func tagFamily(tag string) string {
	start := strings.IndexAny(tag, "-_+")
	if start < 0 {
		return ""
	}
	parts := strings.FieldsFunc(tag[start+1:], func(r rune) bool {
		return r == '-' || r == '_' || r == '+'
	})
	families := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimRight(part, "0123456789."); trimmed != "" {
			families = append(families, trimmed)
		}
	}
	return strings.Join(families, "-")
}
```

- [ ] **Step 6: Run the tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/templatemodel/ ./hivepaas_app/pkg/imageref/`
Expected: PASS.

- [ ] **Step 7: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice/templatemodel hivepaas_app/pkg/imageref \
  hivepaas_app/hperrors/errors_app_template.go hivepaas_app/pkg/translation/messages/en/errors.app_template.en.toml
git commit -m "feat(apptemplate): classify an image override and pick tag candidates"
```

---

## Task 2: Reading tags from a registry - `services/registry`

**Files:**
- Create: `services/registry/reference.go`
- Create: `services/registry/registry.go`
- Create: `hivepaas_app/hperrors/errors_registry.go`
- Create: `hivepaas_app/pkg/translation/messages/en/errors.registry.en.toml`
- Test: `services/registry/reference_test.go`
- Test: `services/registry/registry_test.go`

**Interfaces:**
- Consumes: `imageref.Parse` (Task 1 leaves it unchanged).
- Produces:
  - `type Reference struct { Host, Name string }` with `String() string`
  - `func ParseRepository(image string) Reference`
  - `type Client struct { ... }`, `func New() *Client`
  - `type ListTagsResult struct { Tags []string; Truncated bool }`
  - `func (c *Client) ListTags(ctx context.Context, ref Reference, maxTags int) (*ListTagsResult, error)`
  - `hperrors.ErrRegistryUnavailable`, `hperrors.ErrRegistryRateLimited`

Nothing in HivePaaS talks to a registry over HTTP today - the daemon does the pulling - so this is the first such client. It reads and nothing else.

- [ ] **Step 1: Declare the errors**

`hivepaas_app/hperrors/errors_registry.go`:

```go
package hperrors

// Errors for reading an image registry
var (
	ErrRegistryUnavailable = NewErr(ErrUnavailable, "ERR_REGISTRY_UNAVAILABLE")
	ErrRegistryRateLimited = NewErr(ErrUnavailable, "ERR_REGISTRY_RATE_LIMITED")
)
```

`hivepaas_app/pkg/translation/messages/en/errors.registry.en.toml`:

```toml
# Messages for the errors declared in hivepaas_app/hperrors/errors_registry.go.
# Adding or removing an error there means editing this file.

ERR_REGISTRY_UNAVAILABLE = "The image registry could not be reached"
ERR_REGISTRY_RATE_LIMITED = "The image registry is rate limiting this installation. Try again in a few minutes."
```

- [ ] **Step 2: Write the failing reference tests**

`services/registry/reference_test.go`:

```go
package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Docker's short forms are conveniences of the CLI. The V2 API wants a host and a
// full repository path, and library/ is only implied for single-segment names on
// Docker Hub.
func TestParseRepository(t *testing.T) {
	for image, want := range map[string]Reference{
		"postgres:18.6-alpine3.24":     {Host: "registry-1.docker.io", Name: "library/postgres"},
		"minio/minio:RELEASE.2025":     {Host: "registry-1.docker.io", Name: "minio/minio"},
		"ghcr.io/owner/app:1.2.3":      {Host: "ghcr.io", Name: "owner/app"},
		"registry.example:5000/app:1":  {Host: "registry.example:5000", Name: "app"},
		"localhost:5000/team/app:1.0":  {Host: "localhost:5000", Name: "team/app"},
		"quay.io/org/sub/app:2.0":      {Host: "quay.io", Name: "org/sub/app"},
	} {
		assert.Equal(t, want, ParseRepository(image), image)
	}
}

func TestReferenceString(t *testing.T) {
	assert.Equal(t, "registry-1.docker.io/library/postgres", ParseRepository("postgres:18.6").String())
}
```

- [ ] **Step 3: Write the reference**

`services/registry/reference.go`:

```go
// Package registry reads what an image registry publishes.
//
// It speaks the OCI distribution API directly rather than through the docker
// daemon, because the daemon can pull an image but cannot say which other tags
// exist. It reads and never writes, and it carries no credentials yet: the
// templates HivePaaS ships are public.
//
// TODO: app templates later - authenticate with the registry-auth setting so a
// private registry can be scanned. See
// docs/superpowers/specs/2026-09-17-app-template-image-override-design.md §10.
package registry

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

const (
	dockerHubHost    = "registry-1.docker.io"
	dockerHubLibrary = "library"
	localhostHost    = "localhost"
)

// Reference is a repository as the V2 API addresses it: a host to talk to and a
// path under /v2/.
type Reference struct {
	Host string
	Name string
}

func (r Reference) String() string {
	return r.Host + "/" + r.Name
}

// ParseRepository normalizes a docker image reference into the host and repository
// the registry API uses. `postgres` is `library/postgres` on Docker Hub, while
// `ghcr.io/owner/app` already says where it lives.
//
// The first path segment is a host when it looks like one - it carries a dot or a
// port, or it is localhost. That is docker's own rule, and it is why `minio/minio`
// is a Docker Hub repository and not a host called minio.
func ParseRepository(image string) Reference {
	name := imageref.Parse(image).Repository
	host := dockerHubHost

	if first, rest, found := strings.Cut(name, "/"); found && isHost(first) {
		host, name = first, rest
	}
	if host == dockerHubHost && !strings.Contains(name, "/") {
		name = dockerHubLibrary + "/" + name
	}
	return Reference{Host: host, Name: name}
}

func isHost(segment string) bool {
	return strings.ContainsAny(segment, ".:") || segment == localhostHost
}
```

- [ ] **Step 4: Write the failing client tests**

`services/registry/registry_test.go`:

```go
package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// fakeRegistry answers the way a V2 registry does: a challenge first, then tags in
// pages, with a Link header pointing at the next one.
type fakeRegistry struct {
	mu sync.Mutex

	server    *httptest.Server
	pages     [][]string
	status    int
	tokenSeen []string
	tokens    int
}

func newFakeRegistry(t *testing.T, pages ...[]string) *fakeRegistry {
	t.Helper()
	fake := &fakeRegistry{pages: pages}
	fake.server = httptest.NewServer(fake)
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.URL.Path == "/token" {
		f.tokens++
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"token":"test-token"}`)
		return
	}
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	if r.Header.Get("Authorization") == "" {
		w.Header().Set("WWW-Authenticate",
			fmt.Sprintf(`Bearer realm="%s/token",service="fake",scope="repository:library/postgres:pull"`,
				f.server.URL))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	f.tokenSeen = append(f.tokenSeen, r.Header.Get("Authorization"))

	page := 0
	if last := r.URL.Query().Get("last"); last != "" {
		_, _ = fmt.Sscanf(last, "page-%d", &page)
	}
	if page >= len(f.pages) {
		_, _ = fmt.Fprint(w, `{"name":"library/postgres","tags":[]}`)
		return
	}
	if page+1 < len(f.pages) {
		w.Header().Set("Link",
			fmt.Sprintf(`</v2/library/postgres/tags/list?n=100&last=page-%d>; rel="next"`, page+1))
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"name":"library/postgres","tags":["%s"]}`, strings.Join(f.pages[page], `","`))
}

func (f *fakeRegistry) client() *Client {
	client := New()
	client.endpoint = func(string) string { return f.server.URL }
	return client
}

func testRef() Reference {
	return Reference{Host: "registry-1.docker.io", Name: "library/postgres"}
}

func TestListTagsFollowsTheTokenChallengeAndPages(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6", "18.7"}, []string{"19.0"})

	result, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.NoError(t, err)
	assert.Equal(t, []string{"18.6", "18.7", "19.0"}, result.Tags)
	assert.False(t, result.Truncated)
	assert.Equal(t, 1, fake.tokens, "the token is fetched once and reused for the next page")
	assert.Equal(t, []string{"Bearer test-token", "Bearer test-token"}, fake.tokenSeen)
}

func TestListTagsStopsAtTheCap(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6", "18.7"}, []string{"19.0", "19.1"})

	result, err := fake.client().ListTags(context.Background(), testRef(), 3)

	assert.NoError(t, err)
	assert.Len(t, result.Tags, 3)
	assert.True(t, result.Truncated, "the caller has to know the answer is partial")
}

func TestListTagsReportsRateLimiting(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	fake.status = http.StatusTooManyRequests

	_, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrRegistryRateLimited)
}

func TestListTagsReportsAMissingRepository(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	fake.status = http.StatusNotFound

	_, err := fake.client().ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestListTagsReportsAnUnreachableRegistry(t *testing.T) {
	fake := newFakeRegistry(t, []string{"18.6"})
	client := fake.client()
	fake.server.Close()

	_, err := client.ListTags(context.Background(), testRef(), 100)

	assert.ErrorIs(t, err, hperrors.ErrRegistryUnavailable)
}
```

- [ ] **Step 5: Run them to see them fail**

Run: `go test ./services/registry/`
Expected: FAIL - `undefined: New`.

- [ ] **Step 6: Write the client**

`services/registry/registry.go`:

```go
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	requestTimeout = 10 * time.Second
	pageSize       = 100
	maxPages       = 20
	maxBodySize    = 4 << 20
)

// challengeParamPattern reads key="value" out of a WWW-Authenticate header.
var challengeParamPattern = regexp.MustCompile(`([a-z]+)="([^"]*)"`)

type Client struct {
	http *http.Client
	// endpoint is the base URL for a registry host. It is a field so a test can
	// point the client at an httptest server instead of the real internet.
	endpoint func(host string) string
}

func New() *Client {
	return &Client{
		// The transport reads HTTP_PROXY, HTTPS_PROXY and NO_PROXY from the
		// environment, which is where config.Proxy puts them.
		http:     &http.Client{Timeout: requestTimeout},
		endpoint: func(host string) string { return "https://" + host },
	}
}

type ListTagsResult struct {
	Tags []string
	// Truncated says the registry had more tags than maxTags allowed.
	Truncated bool
}

// ListTags reads the tags of one repository, following the token challenge the
// registry answers with and the Link header it pages by.
//
// maxTags bounds the work: `postgres` alone publishes thousands of tags, and the
// caller only shows a handful of them.
func (c *Client) ListTags(ctx context.Context, ref Reference, maxTags int) (*ListTagsResult, error) {
	result := &ListTagsResult{Tags: make([]string, 0, pageSize)}
	url := fmt.Sprintf("%s/v2/%s/tags/list?n=%d", c.endpoint(ref.Host), ref.Name, pageSize)
	token := ""

	for page := 0; page < maxPages && url != ""; page++ {
		body, next, newToken, err := c.readPage(ctx, url, token, ref)
		if err != nil {
			return nil, err
		}
		token = newToken

		decoded := &struct {
			Tags []string `json:"tags"`
		}{}
		if err = json.Unmarshal(body, decoded); err != nil {
			return nil, hperrors.Wrap(hperrors.ErrRegistryUnavailable).
				WithExtraDetail("%s: %s", ref, err.Error())
		}
		for _, tag := range decoded.Tags {
			if len(result.Tags) >= maxTags {
				result.Truncated = true
				return result, nil
			}
			result.Tags = append(result.Tags, tag)
		}
		url = next
	}
	result.Truncated = result.Truncated || url != ""
	return result, nil
}

// readPage fetches one page, answering a 401 challenge once. The token it returns
// is reused for the next page, so a repository costs one token, not one per page.
func (c *Client) readPage(
	ctx context.Context,
	url, token string,
	ref Reference,
) (body []byte, next, newToken string, err error) {
	res, err := c.get(ctx, url, token)
	if err != nil {
		return nil, "", "", err
	}

	if res.StatusCode == http.StatusUnauthorized && token == "" {
		challenge := res.Header.Get("WWW-Authenticate")
		_ = res.Body.Close()
		if token, err = c.fetchToken(ctx, challenge, ref); err != nil {
			return nil, "", "", err
		}
		if res, err = c.get(ctx, url, token); err != nil {
			return nil, "", "", err
		}
	}
	defer func() { _ = res.Body.Close() }()

	if err = statusError(res, ref); err != nil {
		return nil, "", "", err
	}
	body, err = io.ReadAll(io.LimitReader(res.Body, maxBodySize))
	if err != nil {
		return nil, "", "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).
			WithExtraDetail("%s: %s", ref, err.Error())
	}
	return body, nextPageURL(res, c.endpoint(ref.Host)), token, nil
}

func (c *Client) get(ctx context.Context, url, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s", err.Error())
	}
	return res, nil
}

// fetchToken answers a Bearer challenge. Docker Hub issues an anonymous token for
// a public repository, which is why no credentials appear here.
func (c *Client) fetchToken(ctx context.Context, challenge string, ref Reference) (string, error) {
	params := map[string]string{}
	for _, match := range challengeParamPattern.FindAllStringSubmatch(challenge, -1) {
		params[match[1]] = match[2]
	}
	realm := params["realm"]
	if realm == "" {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).
			WithExtraDetail("%s: the registry asked for authentication without naming a realm", ref)
	}

	url := realm + "?scope=" + params["scope"]
	if service := params["service"]; service != "" {
		url += "&service=" + service
	}
	res, err := c.get(ctx, url, "")
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	if err = statusError(res, ref); err != nil {
		return "", err
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxBodySize))
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, err.Error())
	}
	decoded := &struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}{}
	if err = json.Unmarshal(body, decoded); err != nil {
		return "", hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, err.Error())
	}
	if decoded.Token == "" {
		return decoded.AccessToken, nil
	}
	return decoded.Token, nil
}

func statusError(res *http.Response, ref Reference) error {
	switch {
	case res.StatusCode == http.StatusOK:
		return nil
	case res.StatusCode == http.StatusTooManyRequests:
		return hperrors.Wrap(hperrors.ErrRegistryRateLimited).WithExtraDetail("%s", ref)
	case res.StatusCode == http.StatusNotFound:
		return hperrors.NewNotFound("Image repository").WithMsgLog("repository %v not found", ref)
	default:
		return hperrors.Wrap(hperrors.ErrRegistryUnavailable).WithExtraDetail("%s: %s", ref, res.Status)
	}
}

// nextPageURL reads the Link header registries page with. The URL in it is a path,
// so it is joined back onto the host that answered.
func nextPageURL(res *http.Response, base string) string {
	link := res.Header.Get("Link")
	if link == "" || !strings.Contains(link, `rel="next"`) {
		return ""
	}
	start := strings.Index(link, "<")
	end := strings.Index(link, ">")
	if start < 0 || end <= start {
		return ""
	}
	target := link[start+1 : end]
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return base + target
}
```

- [ ] **Step 7: Run the tests**

Run: `go test ./services/registry/`
Expected: PASS.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && go build ./... && make lint-local`
Expected: `0 issues.`

```bash
git add services/registry hivepaas_app/hperrors/errors_registry.go \
  hivepaas_app/pkg/translation/messages/en/errors.registry.en.toml
git commit -m "feat(registry): read repository tags over the V2 API"
```

---

## Task 3: Offering the tags - `apptemplateservice.ImageTags`

**Files:**
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags.go`
- Create: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags_cache.go`
- Modify: `hivepaas_app/service/apptemplateservice/service.go`, `types.go`
- Modify: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go`
- Modify: `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags_test.go`

**Interfaces:**
- Consumes: `templatemodel.SelectTagCandidates`, `TagCandidate`, `ImageOverrideClass` (Task 1); `registry.ParseRepository`, `Reference`, `Client`, `ListTagsResult` (Task 2); the existing `(*service).Template`.
- Produces:
  - `type apptemplateservice.ImageTagsReq struct { Name, Version, Variant string }`
  - `type apptemplateservice.ImageTag struct { Tag string; Class templatemodel.ImageOverrideClass; Newer bool }`
  - `type apptemplateservice.ImageTagsResp struct { Repository, CurrentTag string; Truncated bool; Tags []*ImageTag }`
  - `apptemplateservice.Service.ImageTags(ctx context.Context, req *ImageTagsReq) (*ImageTagsResp, error)`
  - `apptemplateserviceimpl.New(hpAppService, registryClient *registry.Client, logger)`

- [ ] **Step 1: Declare the types and the method**

Append to `hivepaas_app/service/apptemplateservice/types.go`:

```go
type ImageTagsReq struct {
	Name string
	// Version and Variant are the user's choice; empty means the template's default.
	Version string
	Variant string
}

type ImageTag struct {
	Tag   string
	Class templatemodel.ImageOverrideClass
	Newer bool
}

type ImageTagsResp struct {
	// Repository is where the tags came from, host included, so the dashboard can
	// say what it scanned.
	Repository string
	CurrentTag string
	// Truncated says the registry publishes more tags than were read.
	Truncated bool
	Tags      []*ImageTag
}
```

Add to the `Service` interface in `service.go`:

```go
	// ImageTags lists the tags a user could use instead of the one this template
	// version pins. It reads the registry, so it is slow, it can fail, and it is
	// called only when somebody asks for it.
	ImageTags(ctx context.Context, req *ImageTagsReq) (*ImageTagsResp, error)
```

- [ ] **Step 2: Write the failing tests**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags_test.go`:

```go
package apptemplateserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/services/registry"
)

// fakeTagLister stands in for the registry client, and counts calls so the cache
// can be shown to work.
type fakeTagLister struct {
	tags  []string
	calls int
	err   error
}

func (f *fakeTagLister) ListTags(_ context.Context, _ registry.Reference, _ int) (*registry.ListTagsResult, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &registry.ListTagsResult{Tags: f.tags}, nil
}

func newImageTagsTest(t *testing.T, tags ...string) (*service, *fakeTagLister) {
	t.Helper()
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	lister := &fakeTagLister{tags: tags}
	svc.registryClient = lister
	svc.tagCache = newImageTagsCache()
	return svc, lister
}

func TestImageTagsOffersNewerBuildsOfTheSameImage(t *testing.T) {
	// The test repository's demo template pins demo:2.1.0 for version "2".
	svc, lister := newImageTagsTest(t, "2.1.0", "2.2.0", "1.9.3", "latest", "2")

	resp, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})

	assert.NoError(t, err)
	assert.Equal(t, "registry-1.docker.io/library/demo", resp.Repository)
	assert.Equal(t, "2.1.0", resp.CurrentTag)
	assert.Equal(t, []*apptemplateservice.ImageTag{
		{Tag: "2.2.0", Class: templatemodel.ImageOverrideSameLine, Newer: true},
		{Tag: "1.9.3", Class: templatemodel.ImageOverrideOtherMajor, Newer: false},
	}, resp.Tags)
	assert.Equal(t, 1, lister.calls)
}

func TestImageTagsReadsTheRegistryOncePerRepository(t *testing.T) {
	svc, lister := newImageTagsTest(t, "2.2.0")

	for range 3 {
		_, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})
		assert.NoError(t, err)
	}

	assert.Equal(t, 1, lister.calls, "a repository is read once per cache window, not once per click")
}

func TestImageTagsFollowsTheChosenVersion(t *testing.T) {
	svc, _ := newImageTagsTest(t, "1.9.4")

	resp, err := svc.ImageTags(context.Background(),
		&apptemplateservice.ImageTagsReq{Name: "demo", Version: "1"})

	assert.NoError(t, err)
	assert.Equal(t, "1.9.3", resp.CurrentTag, "the deprecated version is still a version to compare against")
	assert.Equal(t, "1.9.4", resp.Tags[0].Tag)
}

func TestImageTagsRefusesAVersionTheTemplateDoesNotHave(t *testing.T) {
	svc, _ := newImageTagsTest(t, "2.2.0")

	_, err := svc.ImageTags(context.Background(),
		&apptemplateservice.ImageTagsReq{Name: "demo", Version: "99"})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVersionNotFound)
}

func TestImageTagsPassesTheRegistryErrorThrough(t *testing.T) {
	svc, lister := newImageTagsTest(t)
	lister.err = hperrors.Wrap(hperrors.ErrRegistryRateLimited)

	_, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})

	assert.ErrorIs(t, err, hperrors.ErrRegistryRateLimited)
	assert.Equal(t, 1, lister.calls, "a failed read is not cached")
}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/ -run ImageTags`
Expected: FAIL - `svc.registryClient undefined`.

- [ ] **Step 4: Write the cache**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags_cache.go`:

```go
package apptemplateserviceimpl

import (
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/services/registry"
)

// imageTagsCacheTTL is how long a repository's tag list is reused. Tags change
// when an image is published - hours or days apart - and the cost of being a few
// minutes behind is one stale row in a list somebody is reading anyway.
const imageTagsCacheTTL = 10 * time.Minute

type imageTagsCacheEntry struct {
	result  *registry.ListTagsResult
	readAt  time.Time
}

// imageTagsCache keeps one tag list per repository. Only a successful read is
// stored: a rate-limited or unreachable registry must not be remembered as an
// empty repository.
type imageTagsCache struct {
	mu      sync.Mutex
	entries map[string]*imageTagsCacheEntry
}

func newImageTagsCache() *imageTagsCache {
	return &imageTagsCache{entries: map[string]*imageTagsCacheEntry{}}
}

func (c *imageTagsCache) get(key string) (*registry.ListTagsResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.entries[key]
	if !found || time.Since(entry.readAt) > imageTagsCacheTTL {
		return nil, false
	}
	return entry.result, true
}

func (c *imageTagsCache) put(key string, result *registry.ListTagsResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &imageTagsCacheEntry{result: result, readAt: time.Now()}
}
```

- [ ] **Step 5: Write the service method**

`hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/image_tags.go`:

```go
package apptemplateserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/services/registry"
)

const (
	// maxScannedTags bounds what is read from a registry; maxOfferedTags bounds
	// what a person is asked to choose from.
	maxScannedTags = 1000
	maxOfferedTags = 50
)

func (s *service) ImageTags(
	ctx context.Context,
	req *apptemplateservice.ImageTagsReq,
) (*apptemplateservice.ImageTagsResp, error) {
	image, err := s.templateImage(ctx, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	ref := registry.ParseRepository(image)
	tags, found := s.tagCache.get(ref.String())
	if !found {
		if tags, err = s.registryClient.ListTags(ctx, ref, maxScannedTags); err != nil {
			return nil, hperrors.Wrap(err)
		}
		s.tagCache.put(ref.String(), tags)
	}

	candidates := templatemodel.SelectTagCandidates(image, tags.Tags, maxOfferedTags)
	resp := &apptemplateservice.ImageTagsResp{
		Repository: ref.String(),
		CurrentTag: imageref.Parse(image).Tag,
		Truncated:  tags.Truncated,
		Tags:       make([]*apptemplateservice.ImageTag, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		resp.Tags = append(resp.Tags, &apptemplateservice.ImageTag{
			Tag: candidate.Tag, Class: candidate.Class, Newer: candidate.Newer,
		})
	}
	return resp, nil
}

// templateImage is the image the chosen version and variant pin. A deprecated
// version is allowed here: somebody running one is exactly who needs to see what
// else the repository publishes.
func (s *service) templateImage(ctx context.Context, req *apptemplateservice.ImageTagsReq) (string, error) {
	loaded, err := s.Template(ctx, req.Name)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	tmpl := loaded.Template

	version := tmpl.DefaultVersion()
	if req.Version != "" {
		version = tmpl.FindVersion(req.Version)
	}
	if version == nil {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateVersionNotFound).
			WithParam("Template", req.Name).WithParam("Version", req.Version)
	}

	variantName := req.Variant
	if variantName == "" && len(tmpl.Variants) > 0 {
		if variant := tmpl.DefaultVariant(); variant != nil {
			variantName = variant.Name
		}
	}
	image := version.ImageFor(variantName)
	if image == "" {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateVariantUnavailable).
			WithParam("Template", req.Name).WithParam("Version", version.Name).WithParam("Variant", variantName)
	}
	return image, nil
}
```

- [ ] **Step 6: Wire the dependency**

In `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go`, add the interface the service depends on - narrow, so a test does not need an HTTP server - the two fields, and the constructor parameter:

```go
// tagLister is the part of the registry client this service uses. It is an
// interface so a test can answer with a tag list instead of a fake registry.
type tagLister interface {
	ListTags(ctx context.Context, ref registry.Reference, maxTags int) (*registry.ListTagsResult, error)
}

func New(
	hpAppService hpappservice.Service,
	registryClient *registry.Client,
	logger logging.Logger,
) apptemplateservice.Service {
	return &service{
		official:       newOfficialSource(hpAppService),
		registryClient: registryClient,
		tagCache:       newImageTagsCache(),
		logger:         logger,
	}
}

type service struct {
	official       apptemplateservice.Source
	registryClient tagLister
	tagCache       *imageTagsCache
	logger         logging.Logger
	warnOnce       sync.Once
}
```

Add `"context"`, `"github.com/hivepaas/hivepaas/services/registry"` to its imports.

In `hivepaas_app/registry/provides.go`, add `registry.New,` beside the other service constructors, with the import `"github.com/hivepaas/hivepaas/services/registry"`.

- [ ] **Step 7: Run the tests**

Run: `go build ./... && go test ./hivepaas_app/service/apptemplateservice/...`
Expected: PASS.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice hivepaas_app/registry/provides.go
git commit -m "feat(apptemplate): list the tags a template's image repository publishes"
```

---

## Task 4: Applying and remembering the override

**Files:**
- Modify: `hivepaas_app/service/apptemplateservice/templaterender/render.go`
- Modify: `hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/service.go` (pass it through `Render`)
- Modify: `hivepaas_app/service/apptemplateservice/types.go` (`RenderReq`)
- Modify: `hivepaas_app/entity/setting_app_template.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/create.go`, `audit.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/binding_get.go`
- Test: `hivepaas_app/service/apptemplateservice/templaterender/render_test.go`
- Test: `hivepaas_app/entity/setting_app_template_test.go`
- Test: `hivepaas_app/usecase/apptemplateuc/create_test.go`

**Interfaces:**
- Consumes: `templatemodel.ClassifyImageOverride`, `ImageOverrideClass` (Task 1).
- Produces:
  - `templaterender.Request.ImageOverride string`
  - `templaterender.Result.ImageOverride string`, `Result.ImageOverrideClass templatemodel.ImageOverrideClass`
  - `apptemplateservice.RenderReq.ImageOverride string`
  - `entity.AppTemplateSettings.ImageOverride string`
  - `apptemplatedto.AppTemplateBindingResp.ImageOverride string`

The override changes the document that builds the app and nothing else. The base - the render phase 2 merges against - keeps the template's own image, because the base is what the template said, and the override is what the user said.

- [ ] **Step 1: Write the failing render tests**

Append to `hivepaas_app/service/apptemplateservice/templaterender/render_test.go`:

```go
func TestRenderAppliesAnImageOverride(t *testing.T) {
	result, err := render(t, &Request{
		Params:        map[string]any{"dataVolume": "vol-1"},
		ImageOverride: "postgres:17.9-alpine3.22",
	})

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"image": "postgres:17.9-alpine3.22"},
		result.Doc.Deployment.Source["imageSource"])
	assert.Equal(t, "postgres:17.9-alpine3.22", result.ImageOverride)
	assert.Equal(t, templatemodel.ImageOverrideOtherMajor, result.ImageOverrideClass)

	assert.Equal(t, "postgres:17.6-alpine3.22", result.Image, "Image stays what the template pinned")
	assert.Contains(t, string(result.Base), "postgres:17.6-alpine3.22",
		"the base is what the template rendered, so phase 2 can tell the two apart")
	assert.NotContains(t, string(result.Base), "17.9")
}

func TestRenderRefusesAnImageFromAnotherRepository(t *testing.T) {
	_, err := render(t, &Request{
		Params:        map[string]any{"dataVolume": "vol-1"},
		ImageOverride: "mariadb:11.8.9-noble",
	})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateImageNotAllowed)
}

func TestRenderWithoutAnOverrideSaysSo(t *testing.T) {
	result, err := render(t, &Request{Params: map[string]any{"dataVolume": "vol-1"}})

	assert.NoError(t, err)
	assert.Empty(t, result.ImageOverride)
	assert.Empty(t, result.ImageOverrideClass)
}
```

The fixture in that file pins `postgres:17.6-alpine3.22` for the default version, so `17.9-alpine3.22` is the same repository and the same family, and `17.9` against `17.6` is the same major line - the test above asserts `other-major` only where the majors differ. Check the constant at the top of the file before writing the expectations, and use a tag from the version the test renders.

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/service/apptemplateservice/templaterender/ -run Override`
Expected: FAIL - `unknown field ImageOverride`.

- [ ] **Step 3: Apply the override in the renderer**

In `hivepaas_app/service/apptemplateservice/templaterender/render.go`, add to `Request`:

```go
	// ImageOverride replaces the image this version pins, for this app only. It
	// must come from the same repository; ClassifyImageOverride decides.
	ImageOverride string
```

and to `Result`:

```go
	// ImageOverride is the image the user chose instead of Image, empty when the
	// template's own image is in use. Image stays what the template pinned, because
	// the version and release recorded on the app describe the template.
	ImageOverride      string
	ImageOverrideClass templatemodel.ImageOverrideClass
```

In `Render`, after `specmodel.CheckBuildable(doc)` succeeds and before the base render:

```go
	overrideClass, err := applyImageOverride(doc, name, image, req.ImageOverride)
	if err != nil {
		return nil, err
	}
```

and set the two fields on the returned `Result`.

Add the function at the end of the file:

```go
// applyImageOverride swaps the image in the document the app is built from.
//
// It runs after CheckBuildable, so the document has already been accepted as
// something this HivePaaS can build; all that changes is which release of the same
// repository it deploys. It does not touch the tree the base render walks - the
// base is the template's own output, and phase 2 needs to see the override as the
// user's change rather than as something the template did.
func applyImageOverride(
	doc *specmodel.AppDoc,
	templateName, templateImage, override string,
) (templatemodel.ImageOverrideClass, error) {
	if override == "" {
		return "", nil
	}
	class, err := templatemodel.ClassifyImageOverride(templateImage, override)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	var imageSource map[string]any
	if doc.Deployment != nil && doc.Deployment.Source != nil {
		imageSource, _ = doc.Deployment.Source["imageSource"].(map[string]any)
	}
	if imageSource == nil {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: the template deploys no image to replace", templateName)
	}
	imageSource["image"] = override
	return class, nil
}
```

- [ ] **Step 4: Carry it through the service**

In `hivepaas_app/service/apptemplateservice/types.go`, add to `RenderReq`:

```go
	// ImageOverride is an image the user chose instead of the template's, empty to
	// use the template's own.
	ImageOverride string
```

In `apptemplateserviceimpl/service.go`, pass it in `Render`:

```go
	result, err := templaterender.Render(&templaterender.Request{
		Template:      loaded.Template,
		Version:       req.Version,
		Variant:       req.Variant,
		Params:        req.Params,
		ImageOverride: req.ImageOverride,
	})
```

- [ ] **Step 5: Run the render tests**

Run: `go test ./hivepaas_app/service/apptemplateservice/...`
Expected: PASS.

- [ ] **Step 6: Record it on the app**

In `hivepaas_app/entity/setting_app_template.go`, add to `AppTemplateSettings` after `Variant`:

```go
	// ImageOverride is the image the user chose instead of the template's, empty
	// when the template's own image is in use. It is stored so phase 2's update can
	// say "you chose this" rather than guessing from a string comparison.
	ImageOverride string `json:"imageOverride,omitempty"`
```

Append to `hivepaas_app/entity/setting_app_template_test.go`, inside `TestAppTemplateSettingsRoundTrip` after the existing assertions:

```go
	assert.Equal(t, "postgres:18.7-alpine3.24", parsed.ImageOverride)
```

and set it in `newTestAppTemplateSettings`:

```go
		ImageOverride: "postgres:18.7-alpine3.24",
```

In `hivepaas_app/usecase/apptemplateuc/create.go`, inside `newAppTemplateSetting`, beside `Variant`:

```go
	data.ImageOverride = result.ImageOverride
```

In `hivepaas_app/usecase/apptemplateuc/audit.go`, after the `variant` line:

```go
	if result.ImageOverride != "" {
		detail.Set("imageOverride", result.ImageOverride)
	}
```

In `hivepaas_app/usecase/apptemplateuc/apptemplatedto/binding_get.go`, add the field to `AppTemplateBindingResp` after `Variant`:

```go
	// ImageOverride is empty when the app runs the image the template pinned.
	ImageOverride string `json:"imageOverride"`
```

and to `TransformAppTemplateBinding`:

```go
		ImageOverride: settings.ImageOverride,
```

- [ ] **Step 7: Write the failing usecase test**

Append to `hivepaas_app/usecase/apptemplateuc/create_test.go`:

```go
func TestProvisionFromTemplateRecordsAnImageOverride(t *testing.T) {
	uc, fakes := newCreateTest(t)
	rendered := fakes.templates.resp
	rendered.Result.ImageOverride = "postgres:18.7-alpine3.24"
	rendered.Result.ImageOverrideClass = templatemodel.ImageOverrideSameLine

	created := provision(t, uc, fakes)

	var binding *entity.Setting
	for _, setting := range created.app.Settings {
		if setting.Type == base.SettingTypeAppTemplate {
			binding = setting
		}
	}
	stored := &entity.Setting{Type: base.SettingTypeAppTemplate, Data: binding.Data}
	data, err := stored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "postgres:18.7-alpine3.24", data.ImageOverride)
	assert.Equal(t, "18", data.Version, "the version still describes the template, not the override")

	assert.Contains(t, fakes.audit.entries[0].Detail, `"imageOverride":"postgres:18.7-alpine3.24"`)
}
```

The existing `TestTransformAppTemplateBinding` compares a whole struct, so add `ImageOverride: ""` to its expectation, or set a value in the input and expect it back - either keeps the test honest about the new field.

- [ ] **Step 8: Run everything this touched**

Run: `go build ./... && go test ./hivepaas_app/entity/ ./hivepaas_app/service/apptemplateservice/... ./hivepaas_app/usecase/apptemplateuc/...`
Expected: PASS.

- [ ] **Step 9: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/service/apptemplateservice hivepaas_app/entity hivepaas_app/usecase/apptemplateuc
git commit -m "feat(apptemplate): apply and record an image override on the app"
```

---

## Task 5: The endpoints

**Files:**
- Modify: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/create.go`
- Create: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/image_tags_get.go`
- Modify: `hivepaas_app/usecase/apptemplateuc/create.go`
- Create: `hivepaas_app/usecase/apptemplateuc/image_tags.go`
- Create: `hivepaas_app/interface/api/handler/apptemplatehandler/image_tags_get.go`
- Modify: `hivepaas_app/interface/api/server/router_projects.go`
- Modify: `docs/openapi/swagger.json` (generated)
- Test: `hivepaas_app/usecase/apptemplateuc/apptemplatedto/transform_test.go`

**Interfaces:**
- Consumes: `apptemplateservice.ImageTagsReq/ImageTag/ImageTagsResp` (Task 3); `RenderReq.ImageOverride` (Task 4).
- Produces:
  - `apptemplatedto.CreateAppFromTemplateReq.ImageOverride string`
  - `apptemplatedto.GetAppTemplateImageTagsReq/Resp`, `AppTemplateImageTagsResp`, `AppTemplateImageTagResp`, `TransformAppTemplateImageTags`
  - `(*UC).GetAppTemplateImageTags(ctx, auth, req) (*GetAppTemplateImageTagsResp, error)`
  - route `GET /projects/:projectID/app-templates/:templateName/image-tags`

- [ ] **Step 1: Take the override on the create request**

In `hivepaas_app/usecase/apptemplateuc/apptemplatedto/create.go`, add to `CreateAppFromTemplateReq` after `Variant`:

```go
	// ImageOverride replaces the image the template pins, for this app only. It
	// must come from the same repository; the service refuses anything else.
	ImageOverride string `json:"imageOverride"`
```

Trim it in `ModifyRequest`:

```go
	req.ImageOverride = strings.TrimSpace(req.ImageOverride)
```

and bound its length in `Validate`, raising the validator count to 7:

```go
	validators = append(validators,
		basedto.ValidateStr(&req.ImageOverride, false, 1, imageRefMaxLen, "imageOverride")...)
```

with the constant beside the others:

```go
	// imageRefMaxLen is a registry host, a repository path, a tag and a digest.
	imageRefMaxLen = 512
```

In `hivepaas_app/usecase/apptemplateuc/create.go`, pass it to the service:

```go
	rendered, err := uc.appTemplateService.Render(ctx, &apptemplateservice.RenderReq{
		Name:          req.Template,
		Version:       req.Version,
		Variant:       req.Variant,
		Params:        req.Params,
		ImageOverride: req.ImageOverride,
	})
```

- [ ] **Step 2: Write the failing transform test**

Append to `hivepaas_app/usecase/apptemplateuc/apptemplatedto/transform_test.go`:

```go
func TestTransformAppTemplateImageTags(t *testing.T) {
	resp := TransformAppTemplateImageTags(&apptemplateservice.ImageTagsResp{
		Repository: "registry-1.docker.io/library/postgres",
		CurrentTag: "18.6-alpine3.24",
		Truncated:  true,
		Tags: []*apptemplateservice.ImageTag{
			{Tag: "18.7-alpine3.24", Class: templatemodel.ImageOverrideSameLine, Newer: true},
			{Tag: "19.0-alpine3.24", Class: templatemodel.ImageOverrideOtherMajor, Newer: true},
		},
	})

	assert.Equal(t, &AppTemplateImageTagsResp{
		Repository: "registry-1.docker.io/library/postgres",
		CurrentTag: "18.6-alpine3.24",
		Truncated:  true,
		Tags: []*AppTemplateImageTagResp{
			{Tag: "18.7-alpine3.24", Image: "registry-1.docker.io/library/postgres:18.7-alpine3.24",
				Class: "same-line", Newer: true},
			{Tag: "19.0-alpine3.24", Image: "registry-1.docker.io/library/postgres:19.0-alpine3.24",
				Class: "other-major", Newer: true},
		},
	}, resp, "the dashboard posts Image back as imageOverride, so it never builds a reference itself")
}
```

- [ ] **Step 3: Write the DTO**

`hivepaas_app/usecase/apptemplateuc/apptemplatedto/image_tags_get.go`:

```go
package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
)

type GetAppTemplateImageTagsReq struct {
	ProjectID string `json:"-"`
	Name      string `json:"-"`

	Version string `json:"-" mapstructure:"version"`
	Variant string `json:"-" mapstructure:"variant"`
}

func NewGetAppTemplateImageTagsReq() *GetAppTemplateImageTagsReq {
	return &GetAppTemplateImageTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateImageTagsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, templateNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateImageTagsResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *AppTemplateImageTagsResp  `json:"data"`
}

// AppTemplateImageTagsResp is what the registry publishes for the image this
// template version pins.
type AppTemplateImageTagsResp struct {
	Repository string `json:"repository"`
	CurrentTag string `json:"currentTag"`
	// Truncated says the repository has more tags than were read.
	Truncated bool                       `json:"truncated"`
	Tags      []*AppTemplateImageTagResp `json:"tags"`
}

type AppTemplateImageTagResp struct {
	Tag string `json:"tag"`
	// Image is the full reference to post back as imageOverride.
	Image string `json:"image"`
	// Class is one of same-line, other-major or moving, and decides which warning
	// the dashboard shows before creating the app.
	Class string `json:"class"`
	Newer bool   `json:"newer"`
}

func TransformAppTemplateImageTags(tags *apptemplateservice.ImageTagsResp) *AppTemplateImageTagsResp {
	resp := &AppTemplateImageTagsResp{
		Repository: tags.Repository,
		CurrentTag: tags.CurrentTag,
		Truncated:  tags.Truncated,
		Tags:       make([]*AppTemplateImageTagResp, 0, len(tags.Tags)),
	}
	for _, tag := range tags.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateImageTagResp{
			Tag:   tag.Tag,
			Image: tags.Repository + ":" + tag.Tag,
			Class: string(tag.Class),
			Newer: tag.Newer,
		})
	}
	return resp
}
```

- [ ] **Step 4: Write the usecase**

`hivepaas_app/usecase/apptemplateuc/image_tags.go`:

```go
package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateImageTags reads the registry, so it is the one catalog call that
// reaches outside this installation. It runs when somebody asks for it and never
// on a schedule.
func (uc *UC) GetAppTemplateImageTags(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateImageTagsReq,
) (*apptemplatedto.GetAppTemplateImageTagsResp, error) {
	tags, err := uc.appTemplateService.ImageTags(ctx, &apptemplateservice.ImageTagsReq{
		Name:    req.Name,
		Version: req.Version,
		Variant: req.Variant,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateImageTagsResp{
		Data: apptemplatedto.TransformAppTemplateImageTags(tags),
	}, nil
}
```

- [ ] **Step 5: Write the handler and the route**

`hivepaas_app/interface/api/handler/apptemplatehandler/image_tags_get.go`:

```go
package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateImageTags Lists the image tags a template version could use
// @Summary Lists the image tags a template version could use
// @Description Reads the registry on demand. The repository comes from the template, never from the request.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateImageTags
// @Param   projectID path string true "project ID"
// @Param   templateName path string true "template name"
// @Param   version query string false "template version; empty for the default"
// @Param   variant query string false "template variant; empty for the default"
// @Success 200 {object} apptemplatedto.GetAppTemplateImageTagsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-templates/{templateName}/image-tags [get]
func (h *Handler) GetAppTemplateImageTags(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "templateName")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateImageTagsReq()
	req.ProjectID = projectID
	req.Name = name
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateImageTags(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
```

In `hivepaas_app/interface/api/server/router_projects.go`, under the app templates block:

```go
	projectGroup.GET("/:projectID/app-templates/:templateName/image-tags",
		s.handlerRegistry.appTemplateHandler.GetAppTemplateImageTags)
```

- [ ] **Step 6: Build, generate swagger, run the suite**

Run: `go build ./... && make gen-swag && go test ./...`
Expected: PASS, and `git diff --stat docs/openapi/swagger.json` shows `getAppTemplateImageTags` and the new `imageOverride` field.

- [ ] **Step 7: Try it end to end**

Start a second backend on its own port against the templates checkout, and sign in as `docs/DEVELOPMENT.md` §6 describes:

```bash
HP_HTTP_SERVER_PORT=10077 HP_RUN_MODE=app+worker HP_TEMPLATES_DIR=../app-templates make local-app-run &

USER_ID=$(PGPASSWORD=abc123 psql -h localhost -p 35432 -U hivepaas -d hivepaas \
  -tAc "SELECT id FROM users WHERE deleted_at IS NULL ORDER BY created_at LIMIT 1")
TOKEN=$(curl -sS -u hivepaas:abc123 -X POST \
  "http://localhost:10077/_/internal/dev-helper/dev-mode-login?userId=$USER_ID" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])')

curl -sS "http://localhost:10077/_/projects/<projectID>/app-templates/postgres/image-tags" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head -20
```

Expected: the repository reads `registry-1.docker.io/library/postgres`, `currentTag` is what the template pins, and the first entries are real alpine tags newer than it. This call goes to Docker Hub, so it needs the machine to be online; a `412` with `ERR_REGISTRY_RATE_LIMITED` means Docker Hub is throttling and the check should be retried later, not that the code is wrong.

Then create an app with an override one patch ahead of the template and confirm what was recorded:

```bash
curl -sS -X POST "http://localhost:10077/_/projects/<projectID>/<env>/apps/from-template" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"override-check","template":"postgres","params":{"dataVolume":"<volumeID>"},
       "imageOverride":"postgres:18.7-alpine3.24"}'

curl -sS "http://localhost:10077/_/projects/<projectID>/<env>/apps/<appID>/template" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
docker service inspect <project>_<env>_override_check --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}'
```

Expected: `201`; the binding reports `imageOverride` with `version` and `release` still describing the template; once the deployment finishes, the service runs the overridden image and the container reports healthy. Then confirm a refusal:

```bash
curl -sS -X POST ".../apps/from-template" ... -d '{"name":"x","template":"postgres",
  "params":{"dataVolume":"<volumeID>"},"imageOverride":"mariadb:11.8.9-noble"}'
```

Expected: `403` with `ERR_APP_TEMPLATE_IMAGE_NOT_ALLOWED`. Delete the app and stop the second backend.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && make lint-local`
Expected: `0 issues.`

```bash
git add hivepaas_app/usecase/apptemplateuc hivepaas_app/interface/api docs/openapi/swagger.json
git commit -m "feat(apptemplate): image override on create, and a registry tag scan endpoint"
```

---

## Task 6: Bumping the templates - `apptemplate bump`

**Files:**
- Create: `tools/apptemplate/bump.go`
- Modify: `tools/apptemplate/main.go` (the subcommand switch and the usage line)
- Test: `tools/apptemplate/bump_test.go`
- Create: `../app-templates/.github/workflows/bump.yml`
- Modify: `docs/DEVELOPMENT.md` §7

**Interfaces:**
- Consumes: `templaterepo.Load`, `IndexFile`, `ErrorText`; `templatemodel.SelectTagCandidates`, `TagCandidate`, `ImageOverrideSameLine`; `registry.ParseRepository`, `New`, `ListTagsResult`; `imageref.CompareTags`, `Parse`; `buildIndex` from `main.go`.
- Produces: `go run ./tools/apptemplate bump [-dry-run] <dir>`.

A version line is bumped only when every variant it declares has the same newer release. Bumping alpine to 18.7 while debian stays on 18.6 would publish a template whose two variants are different software versions, which is exactly the kind of thing nobody notices until a user asks why their two databases disagree.

- [ ] **Step 1: Write the failing tests**

`tools/apptemplate/bump_test.go`:

```go
package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/registry"
)

// stubTags answers with a fixed tag list per repository.
type stubTags struct {
	byRepo map[string][]string
	calls  int
}

func (s *stubTags) ListTags(
	_ context.Context, ref registry.Reference, _ int,
) (*registry.ListTagsResult, error) {
	s.calls++
	return &registry.ListTagsResult{Tags: s.byRepo[ref.String()]}, nil
}

func useStubTags(t *testing.T, tags map[string][]string) *stubTags {
	t.Helper()
	stub := &stubTags{byRepo: tags}
	previous := newTagLister
	newTagLister = func() tagLister { return stub }
	t.Cleanup(func() { newTagLister = previous })
	return stub
}

func TestPlanVersionBumpTakesTheNewestReleaseEveryVariantHas(t *testing.T) {
	plan, ok := planVersionBump(
		map[string]string{"alpine": "postgres:18.6-alpine3.24", "debian": "postgres:18.6-trixie"},
		map[string][]string{
			"registry-1.docker.io/library/postgres": {
				"18.7-alpine3.24", "18.8-alpine3.25", "18.7-trixie", "19.0-alpine3.24", "19.0-trixie",
			},
		},
	)

	assert.True(t, ok)
	assert.Equal(t, "18.7", plan.Release, "18.8 exists for alpine only, so the line moves to 18.7")
	assert.Equal(t, map[string]string{
		"alpine": "postgres:18.7-alpine3.24",
		"debian": "postgres:18.7-trixie",
	}, plan.Images)
}

func TestPlanVersionBumpStaysInItsMajorLine(t *testing.T) {
	_, ok := planVersionBump(
		map[string]string{"": "mariadb:11.8.9-noble"},
		map[string][]string{"registry-1.docker.io/library/mariadb": {"11.9.1-noble", "12.0.1-noble"}},
	)

	assert.False(t, ok, "11.9 and 12.0 are other lines: a new line is a template change, not a bump")
}

func TestPlanVersionBumpDoesNothingWhenCurrent(t *testing.T) {
	_, ok := planVersionBump(
		map[string]string{"": "mariadb:11.8.9-noble"},
		map[string][]string{"registry-1.docker.io/library/mariadb": {"11.8.9-noble", "11.8.8-noble"}},
	)

	assert.False(t, ok)
}

func TestApplyBumpRewritesImagesAndRelease(t *testing.T) {
	content := []byte(`  - name: "18"
    release: "18.6"
    default: true
    images:
      alpine: postgres:18.6-alpine3.24
      debian: postgres:18.6-trixie
`)

	got, err := applyBump(content, &bumpPlan{
		Version: "18", CurrentRelease: "18.6", Release: "18.7",
		Images: map[string]string{"alpine": "postgres:18.7-alpine3.24", "debian": "postgres:18.7-trixie"},
		CurrentImages: map[string]string{
			"alpine": "postgres:18.6-alpine3.24", "debian": "postgres:18.6-trixie",
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, string(got), `release: "18.7"`)
	assert.Contains(t, string(got), "alpine: postgres:18.7-alpine3.24")
	assert.Contains(t, string(got), "debian: postgres:18.7-trixie")
	assert.NotContains(t, string(got), "18.6")
}

func TestApplyBumpRefusesAnAmbiguousReplacement(t *testing.T) {
	content := []byte("images: {alpine: demo:2.1.0}\nother: demo:2.1.0\n")

	_, err := applyBump(content, &bumpPlan{
		Version: "2", CurrentRelease: "2.1", Release: "2.2",
		Images:        map[string]string{"alpine": "demo:2.2.0"},
		CurrentImages: map[string]string{"alpine": "demo:2.1.0"},
	})

	assert.ErrorIs(t, err, errAmbiguousBump)
}

func TestRunBumpRewritesTheRepositoryAndTheIndex(t *testing.T) {
	dir := copyRepo(t)
	useStubTags(t, map[string][]string{"registry-1.docker.io/library/demo": {"2.2.0", "2.1.0", "1.9.3"}})
	var out bytes.Buffer
	assert.NoError(t, runIndex([]string{dir}, &out))

	out.Reset()
	assert.NoError(t, runBump([]string{dir}, &out))

	template, err := os.ReadFile(filepath.Join(dir, "templates", "demo.yaml"))
	assert.NoError(t, err)
	assert.Contains(t, string(template), "demo:2.2.0")
	assert.Contains(t, string(template), `release: "2.2"`)
	assert.Contains(t, out.String(), "demo 2: 2.1 -> 2.2")
	assert.NoError(t, runIndex([]string{"-check", dir}, &out), "bump leaves index.json current")
}

func TestRunBumpDryRunChangesNothing(t *testing.T) {
	dir := copyRepo(t)
	useStubTags(t, map[string][]string{"registry-1.docker.io/library/demo": {"2.2.0"}})
	before, err := os.ReadFile(filepath.Join(dir, "templates", "demo.yaml"))
	assert.NoError(t, err)
	var out bytes.Buffer

	assert.NoError(t, runBump([]string{"-dry-run", dir}, &out))

	after, err := os.ReadFile(filepath.Join(dir, "templates", "demo.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, before, after)
	assert.Contains(t, out.String(), "2.1 -> 2.2")
	assert.True(t, strings.Contains(out.String(), "dry run"))
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./tools/apptemplate/ -run Bump`
Expected: FAIL - `undefined: planVersionBump`.

- [ ] **Step 3: Write the subcommand**

`tools/apptemplate/bump.go`:

```go
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/services/registry"
)

// maxBumpTags is what a single repository contributes. A repository publishing
// more than this has aliases and old lines far outnumbering what a bump needs.
const maxBumpTags = 1000

var errAmbiguousBump = errors.New("the image to replace appears more than once in the file")

// tagLister is what bump needs from a registry client, as an interface so the
// tests answer without a network.
type tagLister interface {
	ListTags(ctx context.Context, ref registry.Reference, maxTags int) (*registry.ListTagsResult, error)
}

var newTagLister = func() tagLister { return registry.New() }

// bumpPlan is one version line moving to a newer release of the same major.
type bumpPlan struct {
	Template       string
	Version        string
	CurrentRelease string
	Release        string
	// Images and CurrentImages are keyed by variant name, or by "" for a template
	// without variants.
	CurrentImages map[string]string
	Images        map[string]string
}

func (p *bumpPlan) String() string {
	return fmt.Sprintf("%s %s: %s -> %s", p.Template, p.Version, p.CurrentRelease, p.Release)
}

// runBump moves every version line it can to the newest release its major line
// publishes, then rewrites index.json so the repository stays consistent.
func runBump(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("bump", flag.ContinueOnError)
	dryRun := flags.Bool("dry-run", false, "print what would change without writing")
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	repo, _, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}

	lister := newTagLister()
	tagsByRepo := map[string][]string{}
	changed := 0
	for _, file := range repo.Templates {
		plans, err := planTemplateBumps(context.Background(), lister, file, tagsByRepo)
		if err != nil {
			return err
		}
		if len(plans) == 0 {
			continue
		}
		content := file.Content
		for _, plan := range plans {
			fmt.Fprintln(out, plan)
			if content, err = applyBump(content, plan); err != nil {
				return fmt.Errorf("%s: %w", file.Path, err)
			}
		}
		changed++
		if *dryRun {
			continue
		}
		if err = os.WriteFile(filepath.Join(dir, file.Path), content, 0o644); err != nil { //nolint:gosec
			return err
		}
	}

	if changed == 0 {
		fmt.Fprintln(out, "every template is current")
		return nil
	}
	if *dryRun {
		fmt.Fprintln(out, "dry run: nothing written")
		return nil
	}
	return runIndex([]string{dir}, out)
}

// planTemplateBumps plans every version line of one template, reading each
// repository once into tagsByRepo.
func planTemplateBumps(
	ctx context.Context,
	lister tagLister,
	file *templaterepo.TemplateFile,
	tagsByRepo map[string][]string,
) ([]*bumpPlan, error) {
	tmpl := file.Template
	variants := []string{""}
	if len(tmpl.Variants) > 0 {
		variants = variants[:0]
		for _, variant := range tmpl.Variants {
			variants = append(variants, variant.Name)
		}
	}

	plans := make([]*bumpPlan, 0, len(tmpl.Versions))
	for _, version := range tmpl.Versions {
		if version.Deprecated {
			continue
		}
		current := map[string]string{}
		for _, variant := range variants {
			if image := version.ImageFor(variant); image != "" {
				current[variant] = image
			}
		}
		for _, image := range current {
			ref := registry.ParseRepository(image)
			if _, read := tagsByRepo[ref.String()]; read {
				continue
			}
			result, err := lister.ListTags(ctx, ref, maxBumpTags)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}
			tagsByRepo[ref.String()] = result.Tags
		}

		plan, ok := planVersionBump(current, tagsByRepo)
		if !ok {
			continue
		}
		plan.Template, plan.Version, plan.CurrentRelease = tmpl.Metadata.Name, version.Name, version.Release
		plans = append(plans, plan)
	}
	return plans, nil
}

// planVersionBump picks the newest release every variant of one version line
// publishes.
//
// Taking each variant's own newest would let alpine run 18.8 while debian stayed
// on 18.6 - one template describing two different versions of the software. So the
// releases are intersected, and the line moves only as far as all of them reach.
func planVersionBump(current map[string]string, tagsByRepo map[string][]string) (*bumpPlan, bool) {
	var shared map[string]bool
	byVariant := map[string]map[string]string{} // variant → release → tag

	for variant, image := range current {
		tags := tagsByRepo[registry.ParseRepository(image).String()]
		releases := map[string]string{}
		for _, candidate := range templatemodel.SelectTagCandidates(image, tags, 0) {
			if candidate.Class != templatemodel.ImageOverrideSameLine || !candidate.Newer {
				continue
			}
			release := releaseOfTag(candidate.Tag)
			// A release can publish several bases - 18.7-alpine3.24 and
			// 18.7-alpine3.25. The newest of them is the one to move to.
			if existing, found := releases[release]; !found || isNewerTag(existing, candidate.Tag) {
				releases[release] = candidate.Tag
			}
		}
		byVariant[variant] = releases

		names := map[string]bool{}
		for release := range releases {
			names[release] = true
		}
		if shared == nil {
			shared = names
			continue
		}
		for release := range shared {
			if !names[release] {
				delete(shared, release)
			}
		}
	}
	if len(shared) == 0 {
		return nil, false
	}

	best := ""
	for _, release := range slices.Sorted(maps.Keys(shared)) {
		if best == "" || isNewerTag(best, release) {
			best = release
		}
	}

	plan := &bumpPlan{Release: best, CurrentImages: current, Images: map[string]string{}}
	for variant, image := range current {
		plan.Images[variant] = imageref.Parse(image).Repository + ":" + byVariant[variant][best]
	}
	return plan, true
}

// applyBump rewrites the file as text rather than re-marshalling it, so comments
// and formatting survive. Every image it replaces has to appear exactly once, or
// the tool would be guessing which occurrence belongs to this version line.
func applyBump(content []byte, plan *bumpPlan) ([]byte, error) {
	for variant, image := range plan.CurrentImages {
		if bytes.Count(content, []byte(image)) != 1 {
			return nil, fmt.Errorf("%w: %s", errAmbiguousBump, image)
		}
		content = bytes.Replace(content, []byte(image), []byte(plan.Images[variant]), 1)
	}

	currentRelease := []byte(fmt.Sprintf("release: %q", plan.CurrentRelease))
	if bytes.Count(content, currentRelease) != 1 {
		return nil, fmt.Errorf("%w: %s", errAmbiguousBump, currentRelease)
	}
	return bytes.Replace(content, currentRelease, []byte(fmt.Sprintf("release: %q", plan.Release)), 1), nil
}

// releaseOfTag is the version part of a tag: 18.7-alpine3.24 is release 18.7, and
// 11.8.10-noble is 11.8.10 - the same value a template's `release` field carries.
func releaseOfTag(tag string) string {
	if i := strings.IndexAny(tag, "-_+"); i >= 0 {
		return tag[:i]
	}
	return tag
}

func isNewerTag(current, candidate string) bool {
	order, ok := imageref.CompareTags(current, candidate)
	return ok && order < 0
}
```

In `main.go`, add the case and the usage line:

```go
	case "bump":
		err = runBump(args, os.Stdout)
```

```go
	fmt.Fprintln(os.Stderr, "usage: apptemplate lint|index|render|pin|bump [flags] <dir> ...")
```

and the package comment gains:

```go
//	go run ./tools/apptemplate bump   [-dry-run] <dir>
```

- [ ] **Step 4: Run the tests**

Run: `go test ./tools/apptemplate/`
Expected: PASS.

- [ ] **Step 5: Try it against the real repository**

Run: `go run ./tools/apptemplate bump -dry-run ../app-templates`

Expected: either `every template is current`, or one line per version line that moved - `postgres 18: 18.6 -> 18.7`. This call reaches Docker Hub. Read the proposed releases against the registry before writing anything: a bump the tool proposes is still a template change somebody is responsible for.

- [ ] **Step 6: Add the scheduled workflow**

`../app-templates/.github/workflows/bump.yml`:

```yaml
name: bump

on:
  schedule:
    - cron: "0 3 * * 1"
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          path: app-templates

      - uses: actions/checkout@v4
        with:
          repository: hivepaas/hivepaas
          ref: HIVEPAAS_COMMIT_SHA
          path: hivepaas

      - uses: actions/setup-go@v5
        with:
          go-version-file: hivepaas/go.mod
          cache-dependency-path: hivepaas/go.sum

      - name: Bump pinned images
        working-directory: hivepaas
        run: go run ./tools/apptemplate bump ../app-templates

      # The pull request is opened with GITHUB_TOKEN, and GitHub does not run
      # workflows on what that token pushes - so lint runs here instead of waiting
      # for a check that will never start.
      - name: Lint what was produced
        working-directory: hivepaas
        run: |
          go run ./tools/apptemplate lint ../app-templates
          go run ./tools/apptemplate index -check ../app-templates

      - name: Open a pull request
        working-directory: app-templates
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          if [ -z "$(git status --porcelain)" ]; then
            echo "every template is current"
            exit 0
          fi
          branch="bump/$(date -u +%Y-%m-%d)"
          git config user.name "hivepaas-bot"
          git config user.email "bot@hivepaas.com"
          git checkout -b "$branch"
          git commit -am "Bump pinned images"
          git push -u origin "$branch"
          gh pr create --fill --title "Bump pinned images" \
            --body "Opened by the scheduled bump job. Each line moved to the newest release every variant publishes."
```

Replace `HIVEPAAS_COMMIT_SHA` with the commit that contains `tools/apptemplate/bump.go`, the same way `lint.yml` is pinned.

- [ ] **Step 7: Document it**

In `docs/DEVELOPMENT.md` §7, under "### The tool", add the command:

```bash
go run ./tools/apptemplate bump -dry-run ../app-templates   # what the weekly job would change
```

and under "### Publishing", after step 1:

> A scheduled job in `app-templates` runs `bump` every Monday and opens a pull request when a
> version line has a newer release. New major lines are still a person's job: they need a
> `versions` entry, and often an `override`.

- [ ] **Step 8: Lint and commit**

Run: `make fmt && go build ./... && make lint-local && go test ./...`
Expected: `0 issues.` and a green suite.

```bash
git add tools/apptemplate docs/DEVELOPMENT.md
git commit -m "feat(apptemplate): bump pinned images from the registry"

cd ../app-templates
git add .github/workflows/bump.yml
git commit -m "Add the scheduled image bump workflow"
```

---

## After all tasks

- [ ] `go build ./... && make lint-local && go test ./...` in hivepaas - all green.
- [ ] `grep -rn "TODO: app templates" hivepaas_app services | grep -i registry` shows the deferred private-registry work at `services/registry/reference.go`.
- [ ] The spec's §9 live checks have been run: an override one patch ahead reaches the running service, the binding and the audit entry record it, a foreign repository is refused, and the scan lists real tags.
- [ ] Phase 2's update flow is not part of this plan. §6 of the spec says what it must do with `ImageOverride` when it is written.
