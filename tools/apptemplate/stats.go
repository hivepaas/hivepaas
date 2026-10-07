package main

import (
	"bytes"
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

var (
	errNoGitHubToken   = errors.New("stats needs a GitHub token in STATS_GITHUB_TOKEN, GITHUB_TOKEN or GH_TOKEN")
	errGitHubForbidden = errors.New("GitHub refused this token")
)

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
	// The first of these that is set, named in the output - never its value - so
	// that a run says which token GitHub refused: a secret that did not reach the
	// job leaves the job's own token, and nothing else would tell.
	var token, tokenVar string
	for _, name := range []string{"STATS_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		if token = strings.TrimSpace(os.Getenv(name)); token != "" {
			tokenVar = name
			break
		}
	}
	if token == "" {
		return errNoGitHubToken
	}
	fmt.Fprintf(out, "counting with the token in %s\n", tokenVar)

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
	counter := &starCounter{client: &http.Client{Timeout: statsTimeout}, endpoint: *endpoint, token: token, out: out}
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
	// out is where what could not be counted is said; the count goes on.
	out io.Writer
	// noGained is set once the stargazers cannot be read with this token - the
	// job's own token may read a repository's star count and not who starred it
	// when - so that it is said once, and Trending is left empty.
	noGained bool
}

// The stars, and a page of the newest stargazers, the most GitHub returns at
// once, are asked for apart: a token may read the one and not the other.
const (
	starCountFragment  = `fragment count on Repository { stargazerCount }`
	stargazersFragment = `fragment recent on Repository {
  stargazers(first: 100, after: $after, orderBy: {field: STARRED_AT, direction: DESC}) {
    pageInfo { hasNextPage endCursor }
    edges { starredAt }
  }
}`
)

type repoCount struct {
	StargazerCount int `json:"stargazerCount"`
}

type repoRecent struct {
	Stargazers struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Edges []struct {
			StarredAt time.Time `json:"starredAt"`
		} `json:"edges"`
	} `json:"stargazers"`
}

// graphQLError is one error GitHub answers with. One with a path is about a
// field of the answer - a repository it did not find, a field this token may
// not read - and the rest of the answer stands.
type graphQLError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Path    []any  `json:"path"`
}

func (e graphQLError) forbidden() bool {
	return e.Type == "FORBIDDEN" || strings.Contains(e.Message, "not accessible by integration")
}

// count reads each repository's stars, keyed by its lowercased owner/name as
// asked: GitHub matches names ignoring case. A repository it does not find has
// no key; one whose stargazers it may not read has its stars and no gain.
func (c *starCounter) count(ctx context.Context, repos []string, since time.Time) (map[string]starCount, error) {
	repos = slices.Compact(repos)
	counts := map[string]starCount{}
	for batch := range slices.Chunk(repos, statsBatchSize) {
		query, err := batchQuery(batch, "count", starCountFragment, false)
		if err != nil {
			return nil, err
		}
		found := map[string]*repoCount{}
		fieldErrs, err := c.query(ctx, query, nil, &found)
		if err != nil {
			return nil, err
		}
		c.report(batch, fieldErrs, "NOT_FOUND")
		for i, source := range batch {
			if stars := found[alias(i)]; stars != nil {
				counts[strings.ToLower(source)] = starCount{stars: stars.StargazerCount}
			}
		}
		if err = c.countGained(ctx, batch, since, counts); err != nil {
			return nil, err
		}
	}
	return counts, nil
}

// countGained adds to counts the stars each repository of a batch gained since.
func (c *starCounter) countGained(ctx context.Context, batch []string, since time.Time,
	counts map[string]starCount,
) error {
	if c.noGained {
		return nil
	}
	query, err := batchQuery(batch, "recent", stargazersFragment, true)
	if err != nil {
		return err
	}
	recent := map[string]*repoRecent{}
	fieldErrs, err := c.query(ctx, query, map[string]any{"after": nil}, &recent)
	if errors.Is(err, errGitHubForbidden) || slices.ContainsFunc(fieldErrs, graphQLError.forbidden) {
		c.noGained = true
		why := firstMessage(fieldErrs)
		if err != nil {
			why = err.Error()
		}
		fmt.Fprintf(c.out, "::warning::the stargazers cannot be read with this token, so Trending stays empty: "+
			"give the job a token that can, as STATS_GITHUB_TOKEN (%s)\n", why)
		return nil
	}
	if err != nil {
		return err
	}
	c.report(batch, fieldErrs, "NOT_FOUND")
	for i, source := range batch {
		key := strings.ToLower(source)
		got, counted := counts[key]
		page := recent[alias(i)]
		if !counted || page == nil {
			continue
		}
		if got.gained, err = c.gained(ctx, source, page, since); err != nil {
			return err
		}
		counts[key] = got
	}
	return nil
}

// gained counts the stars that came since, reading further pages of stargazers
// while every one on the last page came since and there are more.
func (c *starCounter) gained(ctx context.Context, source string, stars *repoRecent, since time.Time) (int, error) {
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
			Repository *repoRecent `json:"repository"`
		}
		_, err := c.query(ctx, "query($owner: String!, $name: String!, $after: String) {\n"+
			"  repository(owner: $owner, name: $name) { ...recent }\n}\n"+stargazersFragment,
			map[string]any{"owner": owner, "name": name, "after": stars.Stargazers.PageInfo.EndCursor}, &next)
		if err != nil {
			return 0, err
		}
		if next.Repository == nil {
			// The page could not be read: what was counted so far is a floor.
			return gained, nil
		}
		stars = next.Repository
	}
}

// batchQuery asks about a batch of repositories in one query, each under the
// alias of its position.
func batchQuery(batch []string, fragmentName, fragment string, paged bool) (string, error) {
	var query strings.Builder
	if paged {
		query.WriteString("query($after: String) {\n")
	} else {
		query.WriteString("query {\n")
	}
	for i, source := range batch {
		if !templaterepo.GitHubRepoPattern.MatchString(source) {
			return "", fmt.Errorf("%q is not a GitHub repository written as owner/name", source)
		}
		owner, name, _ := strings.Cut(source, "/")
		fmt.Fprintf(&query, "  %s: repository(owner: %q, name: %q) { ...%s }\n", alias(i), owner, name, fragmentName)
	}
	query.WriteString("}\n" + fragment)
	return query.String(), nil
}

func alias(i int) string { return fmt.Sprintf("r%d", i) }

// report says what GitHub could not answer about a batch's repositories, but
// for errors of the type expected - a repository that is not there, which the
// caller says itself.
func (c *starCounter) report(batch []string, fieldErrs []graphQLError, expected string) {
	for _, problem := range fieldErrs {
		if problem.Type == expected {
			continue
		}
		where := "GitHub"
		if len(problem.Path) > 0 {
			if name, ok := problem.Path[0].(string); ok {
				var i int
				if _, err := fmt.Sscanf(name, "r%d", &i); err == nil && i < len(batch) {
					where = batch[i]
				}
			}
		}
		fmt.Fprintf(c.out, "::warning::%s: %s\n", where, problem.Message)
	}
}

func firstMessage(fieldErrs []graphQLError) string {
	if len(fieldErrs) == 0 {
		return ""
	}
	return fieldErrs[0].Message
}

// query runs one GraphQL query into data, and returns the errors about fields
// of the answer, whose rest stands. An error about the query as a whole - a
// limit, a token - fails it.
func (c *starCounter) query(ctx context.Context, query string, variables map[string]any, data any) (
	[]graphQLError, error,
) {
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 300)) //nolint:mnd
		return nil, fmt.Errorf("GitHub answered %s: %s", resp.Status, strings.TrimSpace(string(snippet)))
	}
	var answer struct {
		Data   json.RawMessage `json:"data"`
		Errors []graphQLError  `json:"errors"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return nil, fmt.Errorf("reading GitHub's answer: %w", err)
	}
	var fieldErrs []graphQLError
	for _, problem := range answer.Errors {
		switch {
		case len(problem.Path) > 0:
			fieldErrs = append(fieldErrs, problem)
		case problem.forbidden():
			return nil, fmt.Errorf("%w: %s", errGitHubForbidden, problem.Message)
		default:
			return nil, fmt.Errorf("GitHub: %s", problem.Message)
		}
	}
	if len(answer.Data) == 0 || string(answer.Data) == "null" {
		if len(fieldErrs) > 0 {
			return fieldErrs, nil
		}
		return nil, errors.New("GitHub answered without data")
	}
	return fieldErrs, json.Unmarshal(answer.Data, data)
}
