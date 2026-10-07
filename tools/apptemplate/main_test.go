package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

const testRepoDir = "../../hivepaas_app/service/apptemplateservice/apptemplateserviceimpl/testdata/repo"

func copyRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	assert.NoError(t, os.CopyFS(dir, os.DirFS(testRepoDir)))
	return dir
}

func TestRepoFromRemote(t *testing.T) {
	for remote, want := range map[string]string{
		"https://github.com/hivepaas/app-templates.git":     "hivepaas/app-templates",
		"https://github.com/hivepaas/app-templates":         "hivepaas/app-templates",
		"https://github.com/hivepaas/app-templates/":        "hivepaas/app-templates",
		"git@github.com:hivepaas/app-templates.git":         "hivepaas/app-templates",
		"ssh://git@github.com/hivepaas/app-templates.git\n": "hivepaas/app-templates",
	} {
		repo, err := repoFromRemote(remote)
		assert.NoError(t, err, remote)
		assert.Equal(t, want, repo, remote)
	}

	_, err := repoFromRemote("https://gitlab.com/hivepaas/app-templates.git")
	assert.Error(t, err, "the official source reads templates from GitHub only")
}

func TestLintAndIndex(t *testing.T) {
	dir := copyRepo(t)
	var out bytes.Buffer

	assert.NoError(t, runLint([]string{dir}, &out))
	assert.Contains(t, out.String(), "OK: 3 template(s)",
		"demo, demoweb which depends on it, and demostack which is several apps")

	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale, "there is no index.json yet")
	assert.NoError(t, runIndex([]string{dir}, &out))
	assert.NoError(t, runIndex([]string{"-check", dir}, &out))

	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path, append(content, []byte("# changed\n")...), 0o600))
	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale, "a changed template changes its hash")
}

// index dates the templates that joined since stats.yaml was written, forgets
// the ones that left, and puts the stats in the index; -check catches a
// repository it has not been run on since.
func TestIndexKeepsStatsInLineWithTheTemplates(t *testing.T) {
	dir := copyRepo(t)
	statsPath := filepath.Join(dir, templaterepo.StatsFile)
	assert.NoError(t, os.WriteFile(statsPath, []byte(`apiVersion: hivepaas.com/v1
kind: TemplateStats
counted: "2026-10-05"
templates:
  demo: {added: "2026-09-01", stars: 1200, starsGained: 40}
  gone: {added: "2026-09-02"}
`), 0o600))
	defer func(original func() string) { today = original }(today)
	today = func() string { return "2026-10-07" }
	var out bytes.Buffer

	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale)
	assert.NoError(t, runIndex([]string{dir}, &out))
	assert.NoError(t, runIndex([]string{"-check", dir}, &out))

	stats, err := os.ReadFile(statsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(stats), `  demo: {added: "2026-09-01", stars: 1200, starsGained: 40}`)
	assert.Contains(t, string(stats), `  demoweb: {added: "2026-10-07"}`, "a template that joined is dated today")
	assert.NotContains(t, string(stats), "gone", "a template that left is forgotten")

	data, err := os.ReadFile(filepath.Join(dir, templaterepo.IndexFile))
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"stats": {
        "added": "2026-09-01",
        "stars": 1200,
        "starsGained": 40
      }`)

	assert.NoError(t, os.WriteFile(statsPath, []byte(strings.Replace(string(stats), "1200", "1300", 1)), 0o600))
	assert.ErrorIs(t, runIndex([]string{"-check", dir}, &out), errIndexStale, "new counts are not in the index yet")
}

func TestLintReportsProblems(t *testing.T) {
	dir := copyRepo(t)
	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path,
		[]byte(strings.Replace(string(content), "databases/sql", "databases/graph", 1)), 0o600))
	var out bytes.Buffer

	err = runLint([]string{dir}, &out)

	assert.ErrorIs(t, err, errProblems)
	assert.Contains(t, out.String(), `category "databases/graph" is not in categories.yaml`)
}

func TestRender(t *testing.T) {
	dir := copyRepo(t)
	var out bytes.Buffer

	assert.NoError(t, runRender([]string{"-param", "dataVolume=vol-1", dir, "demo"}, &out))
	assert.Contains(t, out.String(), "demo:2.1.0")
	assert.Contains(t, out.String(), "source: vol-1")

	out.Reset()
	assert.NoError(t, runRender([]string{"-version", "1", "-param", "dataVolume=vol-1", dir, "demo"}, &out))
	assert.Contains(t, out.String(), "demo:1.9.3", "render includes deprecated versions, for authors")
}

func TestPin(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := copyRepo(t)
	var out bytes.Buffer
	assert.NoError(t, runIndex([]string{dir}, &out))
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "."},
		{"-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-qm", "templates"},
		{"remote", "add", "origin", "https://github.com/hivepaas/app-templates.git"},
	} {
		_, err := gitOutput(dir, args...)
		assert.NoError(t, err)
	}

	out.Reset()
	assert.NoError(t, runPin([]string{dir}, &out))

	var pin struct {
		Repo        string `json:"repo"`
		Commit      string `json:"commit"`
		IndexSHA256 string `json:"indexSha256"`
	}
	assert.NoError(t, json.Unmarshal(out.Bytes(), &pin))
	head, err := gitOutput(dir, "rev-parse", "HEAD")
	assert.NoError(t, err)
	index, err := os.ReadFile(filepath.Join(dir, "index.json"))
	assert.NoError(t, err)
	sum := sha256.Sum256(index)
	assert.Equal(t, "hivepaas/app-templates", pin.Repo)
	assert.Equal(t, strings.TrimSpace(string(head)), pin.Commit)
	assert.Equal(t, hex.EncodeToString(sum[:]), pin.IndexSHA256)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), append(index, '\n'), 0o600))
	assert.Error(t, runPin([]string{dir}, &out), "a pin names the committed index, not a changed one")
}

// An installation refuses an index larger than the limit outright, so the
// warning has to arrive while templates are being added.
func TestIndexSizeWarning(t *testing.T) {
	var out bytes.Buffer
	warnIndexSize(&out, int(templaterepo.MaxIndexSize)/2)
	assert.Empty(t, out.String(), "half of the limit is not worth saying")

	out.Reset()
	warnIndexSize(&out, templaterepo.MaxIndexSize*85/100) //nolint:mnd
	assert.Contains(t, out.String(), "warning: index.json is")
	assert.Contains(t, out.String(), "85%")
}
