package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

// fakeGitHub answers stargazer queries for the repositories it holds, each
// with its stargazers' times newest first, a page of 100 at a time.
type fakeGitHub struct {
	repos   map[string][]time.Time
	queries int
	fail    string
	// forbid refuses the stargazers as GitHub refuses an integration's token:
	// "field" with an error per repository, "query" with one for the query.
	forbid string
}

var aliasPattern = regexp.MustCompile(`(r\d+): repository\(owner: "([^"]+)", name: "([^"]+)"\)`)

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.queries++
	if r.Header.Get("Authorization") != "bearer test-token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var req struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if f.fail != "" {
		_ = json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"type": "RATE_LIMITED", "message": f.fail}}})
		return
	}
	after, _ := req.Variables["after"].(string)
	data := map[string]any{}
	var errs []any
	refused := map[string]any{"type": "FORBIDDEN", "message": "Resource not accessible by integration"}
	if f.forbid == "query" && strings.Contains(req.Query, "stargazers(") {
		_ = json.NewEncoder(w).Encode(map[string]any{"errors": []any{refused}})
		return
	}
	if owner, ok := req.Variables["owner"].(string); ok {
		data["repository"] = f.page(owner+"/"+req.Variables["name"].(string), after)
	} else {
		for _, match := range aliasPattern.FindAllStringSubmatch(req.Query, -1) {
			page := f.page(match[2]+"/"+match[3], after)
			switch {
			case page == nil:
				errs = append(errs, map[string]any{"type": "NOT_FOUND", "message": "not found", "path": []any{match[1]}})
			case f.forbid == "field" && strings.Contains(req.Query, "stargazers("):
				// The connection is not nullable: its error nulls the repository.
				errs = append(errs, map[string]any{"type": refused["type"], "message": refused["message"],
					"path": []any{match[1], "stargazers"}})
				page = nil
			}
			data[match[1]] = page
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "errors": errs})
}

func (f *fakeGitHub) page(repo, after string) any {
	times, ok := f.repos[strings.ToLower(repo)]
	if !ok {
		return nil
	}
	start := 0
	if after != "" {
		_, _ = fmt.Sscanf(after, "cursor-%d", &start)
	}
	end := min(start+100, len(times))
	edges := make([]any, 0, end-start)
	for _, at := range times[start:end] {
		edges = append(edges, map[string]any{"starredAt": at.Format(time.RFC3339)})
	}
	return map[string]any{
		"nameWithOwner":  repo,
		"stargazerCount": len(times) + 5000,
		"stargazers": map[string]any{
			"pageInfo": map[string]any{"hasNextPage": end < len(times), "endCursor": fmt.Sprintf("cursor-%d", end)},
			"edges":    edges,
		},
	}
}

// starTimes is gained stars spread over the last week, newest first, then old
// ones from a year ago.
func starTimes(gained, old int) []time.Time {
	now := time.Now().UTC()
	times := make([]time.Time, 0, gained+old)
	for i := range gained {
		times = append(times, now.Add(-time.Duration(i)*time.Minute))
	}
	for range old {
		times = append(times, now.AddDate(-1, 0, 0))
	}
	return times
}

func setSource(t *testing.T, dir, template, source string) {
	t.Helper()
	path := filepath.Join(dir, "templates", template+".yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	content = []byte(strings.Replace(string(content), "  icon:", "  links: {source: "+source+"}\n  icon:", 1))
	assert.NoError(t, os.WriteFile(path, content, 0o600))
}

func TestStatsCountsStarsAndWhatTheyGained(t *testing.T) {
	dir := copyRepo(t)
	setSource(t, dir, "demo", "https://github.com/Acme/Demo")
	setSource(t, dir, "demoweb", "https://github.com/acme/gone")
	github := &fakeGitHub{repos: map[string][]time.Time{
		// Three pages of stars gained, so the count reads past the first.
		"acme/demo": starTimes(250, 30),
	}}
	server := httptest.NewServer(github)
	defer server.Close()
	t.Setenv("STATS_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "test-token")
	var out bytes.Buffer

	assert.NoError(t, runStats([]string{"-endpoint", server.URL, dir}, &out))

	stats, err := os.ReadFile(filepath.Join(dir, templaterepo.StatsFile))
	assert.NoError(t, err)
	assert.Regexp(t, `  demo: \{added: "[0-9-]+", stars: 5280, starsGained: 250\}`, string(stats))
	assert.Contains(t, out.String(), "demoweb: acme/gone is not on GitHub any more")
	assert.Regexp(t, `  demoweb: \{added: "[0-9-]+"\}`, string(stats))
	assert.Equal(t, 4, github.queries, "the batch's stars, its stargazers, then two more pages of acme/demo")
	assert.Contains(t, out.String(), "counting with the token in GITHUB_TOKEN")
	assert.NoError(t, runIndex([]string{"-check", dir}, &out), "stats writes the index")
}

// The job's own token may read a repository's stars and not its stargazers:
// the stars are counted, Trending is left empty, and the run says why.
func TestStatsCountsStarsWhenTheStargazersAreRefused(t *testing.T) {
	for _, forbid := range []string{"field", "query"} {
		t.Run(forbid, func(t *testing.T) {
			dir := copyRepo(t)
			setSource(t, dir, "demo", "https://github.com/acme/demo")
			server := httptest.NewServer(&fakeGitHub{forbid: forbid, repos: map[string][]time.Time{
				"acme/demo": starTimes(250, 30),
			}})
			defer server.Close()
			t.Setenv("STATS_GITHUB_TOKEN", "test-token")
			var out bytes.Buffer

			assert.NoError(t, runStats([]string{"-endpoint", server.URL, dir}, &out))

			stats, err := os.ReadFile(filepath.Join(dir, templaterepo.StatsFile))
			assert.NoError(t, err)
			assert.Regexp(t, `  demo: \{added: "[0-9-]+", stars: 5280\}`, string(stats))
			assert.Contains(t, out.String(), "::warning::the stargazers cannot be read with this token")
			assert.Contains(t, out.String(), "Resource not accessible by integration")
		})
	}
}

func TestStatsFailsWithoutCountingAnything(t *testing.T) {
	dir := copyRepo(t)
	setSource(t, dir, "demo", "https://github.com/acme/demo")

	t.Setenv("STATS_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	assert.ErrorIs(t, runStats([]string{dir}, &bytes.Buffer{}), errNoGitHubToken)

	server := httptest.NewServer(&fakeGitHub{fail: "API rate limit exceeded"})
	defer server.Close()
	t.Setenv("GITHUB_TOKEN", "test-token")
	err := runStats([]string{"-endpoint", server.URL, dir}, &bytes.Buffer{})
	assert.ErrorContains(t, err, "API rate limit exceeded")
	_, statErr := os.Stat(filepath.Join(dir, templaterepo.StatsFile))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "a failed count writes nothing")
}
