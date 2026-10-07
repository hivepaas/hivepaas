package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

const (
	githubGraphQLURL = "https://api.github.com/graphql"
	// statsBatchSize is how many repositories one query asks about.
	statsBatchSize = 25
	// maxStargazerPages caps how far back one repository is read for the stars it
	// gained: past it the count is a floor, and a project gaining that many in
	// the window is at the top of the order whatever the exact number.
	maxStargazerPages = 20
	statsTimeout      = 30 * time.Second
)

var errNoGitHubToken = errors.New("stats needs a GitHub token in GITHUB_TOKEN (or GH_TOKEN)")

// runStats counts the GitHub stars of every template's project into stats.yaml,
// and how many came in the last templatemodel.TrendingWindowDays days, then
// writes the index. It runs weekly in CI; a template whose project is not on
// GitHub, or no longer there, keeps no counts.
func runStats(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("stats", flag.ContinueOnError)
	endpoint := flags.String("endpoint", githubGraphQLURL, "GitHub's GraphQL endpoint")
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(cmp.Or(os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN")))
	if token == "" {
		return errNoGitHubToken
	}

	repo, _, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	templaterepo.SyncStats(repo, today())

	byTemplate := map[string]string{}
	for _, file := range repo.Templates {
		if source := templaterepo.StatsRepo(repo, file); source != "" && !file.Template.Metadata.Internal {
			byTemplate[file.Template.Metadata.Name] = source
		}
	}
	since := time.Now().UTC().AddDate(0, 0, -templatemodel.TrendingWindowDays)
	counter := &starCounter{client: &http.Client{Timeout: statsTimeout}, endpoint: *endpoint, token: token}
	counts, err := counter.count(context.Background(), slices.Sorted(maps.Values(byTemplate)), since)
	if err != nil {
		return err
	}

	for name, entry := range repo.Stats.Templates {
		source := byTemplate[name]
		got, found := counts[strings.ToLower(source)]
		if source != "" && !found {
			fmt.Fprintf(out, "%s: %s is not on GitHub any more: it keeps no counts\n", name, source)
		}
		entry.Stars, entry.StarsGained = got.stars, got.gained
	}
	repo.Stats.Counted = today()
	statsPath := filepath.Join(dir, templaterepo.StatsFile)
	if err = os.WriteFile(statsPath, templaterepo.MarshalStats(repo.Stats), 0o644); err != nil { //nolint:gosec
		return err
	}
	fmt.Fprintf(out, "counted %d repositories for %d templates\n", len(counts), len(byTemplate))
	return runIndex([]string{dir}, out)
}

type starCount struct {
	stars  int
	gained int
}

type starCounter struct {
	client   *http.Client
	endpoint string
	token    string
}

// stargazersFragment reads a repository's stars and a page of its newest
// stargazers, the most GitHub returns at once.
const stargazersFragment = `fragment stars on Repository {
  nameWithOwner
  stargazerCount
  stargazers(first: 100, after: $after, orderBy: {field: STARRED_AT, direction: DESC}) {
    pageInfo { hasNextPage endCursor }
    edges { starredAt }
  }
}`

type repoStars struct {
	NameWithOwner  string `json:"nameWithOwner"`
	StargazerCount int    `json:"stargazerCount"`
	Stargazers     struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Edges []struct {
			StarredAt time.Time `json:"starredAt"`
		} `json:"edges"`
	} `json:"stargazers"`
}

// count reads each repository's stars, keyed by its lowercased owner/name as
// asked: GitHub matches names ignoring case. A repository it does not find has
// no key.
func (c *starCounter) count(ctx context.Context, repos []string, since time.Time) (map[string]starCount, error) {
	repos = slices.Compact(repos)
	counts := map[string]starCount{}
	for batch := range slices.Chunk(repos, statsBatchSize) {
		found, err := c.firstPages(ctx, batch)
		if err != nil {
			return nil, err
		}
		for i, source := range batch {
			stars := found[fmt.Sprintf("r%d", i)]
			if stars == nil {
				continue
			}
			gained, err := c.gained(ctx, source, stars, since)
			if err != nil {
				return nil, err
			}
			counts[strings.ToLower(source)] = starCount{stars: stars.StargazerCount, gained: gained}
		}
	}
	return counts, nil
}

// firstPages asks about a batch of repositories in one query, each under an
// alias of its position.
func (c *starCounter) firstPages(ctx context.Context, batch []string) (map[string]*repoStars, error) {
	var query strings.Builder
	query.WriteString("query($after: String) {\n")
	for i, source := range batch {
		if !templaterepo.GitHubRepoPattern.MatchString(source) {
			return nil, fmt.Errorf("%q is not a GitHub repository written as owner/name", source)
		}
		owner, name, _ := strings.Cut(source, "/")
		fmt.Fprintf(&query, "  r%d: repository(owner: %q, name: %q) { ...stars }\n", i, owner, name)
	}
	query.WriteString("}\n" + stargazersFragment)
	found := map[string]*repoStars{}
	return found, c.query(ctx, query.String(), map[string]any{"after": nil}, &found)
}

// gained counts the stars that came since, reading further pages of stargazers
// while every one on the last page came since and there are more.
func (c *starCounter) gained(ctx context.Context, source string, stars *repoStars, since time.Time) (int, error) {
	gained := 0
	for page := 1; ; page++ {
		for _, edge := range stars.Stargazers.Edges {
			if edge.StarredAt.Before(since) {
				return gained, nil
			}
			gained++
		}
		if !stars.Stargazers.PageInfo.HasNextPage || page >= maxStargazerPages {
			return gained, nil
		}
		owner, name, _ := strings.Cut(source, "/")
		var next struct {
			Repository *repoStars `json:"repository"`
		}
		err := c.query(ctx, "query($owner: String!, $name: String!, $after: String) {\n"+
			"  repository(owner: $owner, name: $name) { ...stars }\n}\n"+stargazersFragment,
			map[string]any{"owner": owner, "name": name, "after": stars.Stargazers.PageInfo.EndCursor}, &next)
		if err != nil {
			return 0, err
		}
		if next.Repository == nil {
			return gained, nil
		}
		stars = next.Repository
	}
}

// query runs one GraphQL query into data. Errors GitHub returns beside data are
// the repositories it did not find, which data leaves null; errors without any
// data fail the run.
func (c *starCounter) query(ctx context.Context, query string, variables map[string]any, data any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 300)) //nolint:mnd
		return fmt.Errorf("GitHub answered %s: %s", resp.Status, strings.TrimSpace(string(snippet)))
	}
	var answer struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return fmt.Errorf("reading GitHub's answer: %w", err)
	}
	for _, problem := range answer.Errors {
		if problem.Type != "NOT_FOUND" {
			return fmt.Errorf("GitHub: %s", problem.Message)
		}
	}
	if len(answer.Data) == 0 || string(answer.Data) == "null" {
		return errors.New("GitHub answered without data")
	}
	return json.Unmarshal(answer.Data, data)
}
