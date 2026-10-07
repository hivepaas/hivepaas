package templaterepo

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const StatsFile = "stats.yaml"

// GitHubRepoPattern is a repository as stats.yaml names one: owner/name.
var GitHubRepoPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+$`)

// githubSourcePattern is a source link to a GitHub repository, or to a page of it.
var githubSourcePattern = regexp.MustCompile(
	`^https://(?:www\.)?github\.com/([A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+?)(?:\.git)?(?:/.*)?$`)

// loadStats reads stats.yaml. A repository without one has no stats yet, which
// is no problem: the index is built without them, and `apptemplate index`
// writes the file.
func (r *Repo) loadStats(fsys fs.FS) []Problem {
	data, err := readLimited(fsys, StatsFile, MaxIndexSize)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []Problem{{Path: StatsFile, Message: ErrorText(err)}}
	}
	stats, err := templatemodel.DecodeStats(data)
	if err != nil {
		return []Problem{{Path: StatsFile, Message: ErrorText(err)}}
	}
	r.Stats = stats
	return nil
}

// SyncStats brings stats.yaml in line with the templates: a template without an
// entry joins today, and the entry of a template no longer here goes. It
// reports whether anything changed.
func SyncStats(repo *Repo, today string) bool {
	changed := false
	if repo.Stats == nil {
		repo.Stats = &templatemodel.Stats{
			APIVersion: templatemodel.APIVersion,
			Kind:       templatemodel.KindTemplateStats,
			Templates:  map[string]*templatemodel.TemplateStats{},
		}
		changed = true
	}
	names := map[string]bool{}
	for _, file := range repo.Templates {
		name := file.Template.Metadata.Name
		names[name] = true
		entry := repo.Stats.Templates[name]
		if entry == nil {
			entry = &templatemodel.TemplateStats{}
			repo.Stats.Templates[name] = entry
		}
		if entry.Added == "" {
			entry.Added = today
			changed = true
		}
	}
	for name := range repo.Stats.Templates {
		if !names[name] {
			delete(repo.Stats.Templates, name)
			changed = true
		}
	}
	return changed
}

// StatsRepo is the GitHub repository whose stars a template counts, or "" for
// one with none: the repo stats.yaml names for it, else its source link.
func StatsRepo(repo *Repo, file *TemplateFile) string {
	if entry := repo.statsOf(file.Template.Metadata.Name); entry != nil && entry.Repo != "" {
		return entry.Repo
	}
	if links := file.Template.Metadata.Links; links != nil {
		if match := githubSourcePattern.FindStringSubmatch(links.Source); match != nil {
			return match[1]
		}
	}
	return ""
}

func (r *Repo) statsOf(name string) *templatemodel.TemplateStats {
	if r.Stats == nil {
		return nil
	}
	return r.Stats.Templates[name]
}

// indexStats is what the index carries of a template's stats, nil for none.
func (r *Repo) indexStats(name string) *templatemodel.IndexStats {
	entry := r.statsOf(name)
	if entry == nil {
		return nil
	}
	return &templatemodel.IndexStats{Added: entry.Added, Stars: entry.Stars, StarsGained: entry.StarsGained}
}

// MarshalStats writes stats.yaml: one template a line, by name, so that a week's
// counts are a diff a person can read. It is byte-identical for the same stats.
func MarshalStats(stats *templatemodel.Stats) []byte {
	var b strings.Builder
	b.WriteString("# Written by `apptemplate index`, which dates a template the day it joins, and\n" +
		"# `apptemplate stats`, which counts GitHub stars every week. Only `repo` is\n" +
		"# written by hand: the repository to count for a template whose source link\n" +
		"# is not on GitHub.\n")
	fmt.Fprintf(&b, "apiVersion: %s\nkind: %s\n", stats.APIVersion, stats.Kind)
	if stats.Counted != "" {
		fmt.Fprintf(&b, "counted: %q\n", stats.Counted)
	}
	if len(stats.Templates) == 0 {
		b.WriteString("templates: {}\n")
		return []byte(b.String())
	}
	b.WriteString("templates:\n")
	for _, name := range slices.Sorted(maps.Keys(stats.Templates)) {
		entry := stats.Templates[name]
		fields := []string{fmt.Sprintf("added: %q", entry.Added)}
		if entry.Repo != "" {
			fields = append(fields, "repo: "+entry.Repo)
		}
		if entry.Stars != 0 {
			fields = append(fields, fmt.Sprintf("stars: %d", entry.Stars))
		}
		if entry.StarsGained != 0 {
			fields = append(fields, fmt.Sprintf("starsGained: %d", entry.StarsGained))
		}
		fmt.Fprintf(&b, "  %s: {%s}\n", name, strings.Join(fields, ", "))
	}
	return []byte(b.String())
}

// lintStats checks what stats.yaml says. Whether every template has an entry is
// not its question: `apptemplate index` adds them, and its -check is what tells
// a repository it has not been run.
func lintStats(repo *Repo) []Problem {
	if repo.Stats == nil {
		return nil
	}
	var problems []Problem
	add := func(format string, args ...any) {
		problems = append(problems, Problem{Path: StatsFile, Message: fmt.Sprintf(format, args...)})
	}
	if repo.Stats.Counted != "" && !templatemodel.ValidDate(repo.Stats.Counted) {
		add("counted %q is not a day written as YYYY-MM-DD", repo.Stats.Counted)
	}
	for _, name := range slices.Sorted(maps.Keys(repo.Stats.Templates)) {
		entry := repo.Stats.Templates[name]
		switch {
		case entry == nil:
			add("%s: the entry is empty", name)
		case entry.Added != "" && !templatemodel.ValidDate(entry.Added):
			add("%s: added %q is not a day written as YYYY-MM-DD", name, entry.Added)
		case entry.Repo != "" && !GitHubRepoPattern.MatchString(entry.Repo):
			add("%s: repo %q is not a GitHub repository written as owner/name", name, entry.Repo)
		case entry.Stars < 0 || entry.StarsGained < 0:
			add("%s: a count cannot be negative", name)
		}
	}
	return problems
}
