package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/registry"
)

// stubTags answers with a fixed tag list per repository, or a fixed error.
type stubTags struct {
	byRepo map[string][]string
	failed map[string]error
	calls  int
}

func (s *stubTags) ListTags(
	_ context.Context, ref registry.Reference, _ int,
) (*registry.ListTagsResult, error) {
	s.calls++
	if err := s.failed[ref.String()]; err != nil {
		return nil, err
	}
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
	plan, ok := planVersionBump("18",
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
	_, ok := planVersionBump("11.8",
		map[string]string{"": "mariadb:11.8.9-noble"},
		map[string][]string{"registry-1.docker.io/library/mariadb": {"11.9.1-noble", "12.0.1-noble"}},
	)

	assert.False(t, ok, "11.9 and 12.0 are other lines: a new line is a template change, not a bump")
}

func TestPlanVersionBumpDoesNothingWhenCurrent(t *testing.T) {
	_, ok := planVersionBump("11.8",
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

// A variant whose tag extends another's is a different image, not a second
// occurrence of it - OpenClaw's 2026.9.5 and 2026.9.5-browser.
func TestApplyBumpTellsAnImageFromOneItStarts(t *testing.T) {
	content := []byte(`  - name: "2026.9"
    release: "2026.9.5"
    images:
      standard: openclaw/openclaw:2026.9.5
      browser: openclaw/openclaw:2026.9.5-browser
`)

	got, err := applyBump(content, &bumpPlan{
		Version: "2026.9", CurrentRelease: "2026.9.5", Release: "2026.9.8",
		Images: map[string]string{
			"standard": "openclaw/openclaw:2026.9.8", "browser": "openclaw/openclaw:2026.9.8-browser",
		},
		CurrentImages: map[string]string{
			"standard": "openclaw/openclaw:2026.9.5", "browser": "openclaw/openclaw:2026.9.5-browser",
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, string(got), "standard: openclaw/openclaw:2026.9.8\n")
	assert.Contains(t, string(got), "browser: openclaw/openclaw:2026.9.8-browser\n")
	assert.NotContains(t, string(got), "2026.9.5")
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
	assert.Contains(t, string(template), `release: "2.2.0"`)
	assert.Contains(t, out.String(), "demo 2: 2.1 -> 2.2.0")
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
	assert.Contains(t, out.String(), "2.1 -> 2.2.0")
	assert.True(t, strings.Contains(out.String(), "dry run"))
}

func TestReleaseInStyleOf(t *testing.T) {
	assert.Equal(t, "0.26.15", releaseInStyleOf("0.26.12", "v0.26.15"), "tags with v, template without")
	assert.Equal(t, "v2.4.2", releaseInStyleOf("v2.4.1", "v2.4.2"))
	assert.Equal(t, "v2.4.2", releaseInStyleOf("v2.4.1", "2.4.2"), "tags without v, template with")
	assert.Equal(t, "18.7", releaseInStyleOf("18.6", "18.7"))
}

// A registry refusing one repository costs that template its bump, not the
// whole run: the others are bumped and the index written.
func TestRunBumpSkipsATemplateItCannotRead(t *testing.T) {
	dir := copyRepo(t)
	stub := useStubTags(t, map[string][]string{"registry-1.docker.io/library/demo": {"2.2.0", "2.1.0", "1.9.3"}})
	stub.failed = map[string]error{"registry-1.docker.io/library/demoweb": errors.New("rate limited")}
	var out bytes.Buffer
	assert.NoError(t, runIndex([]string{dir}, &out))

	out.Reset()
	assert.NoError(t, runBump([]string{dir}, &out))

	assert.Contains(t, out.String(), "::warning::templates/demoweb.yaml: not bumped: ")
	assert.Contains(t, out.String(), "1 template(s) not bumped: demoweb")
	assert.Contains(t, out.String(), "demo 2: 2.1 -> 2.2.0")
	assert.NoError(t, runIndex([]string{"-check", dir}, &out))
}
