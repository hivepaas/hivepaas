package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func lintWithStats(t *testing.T, stats string) string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), StatsFile,
		"apiVersion: hivepaas.com/v1\nkind: TemplateStats\n"+stats))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return strings.Join(messages, "\n")
}

func TestLintAcceptsStats(t *testing.T) {
	assert.Empty(t, lintWithStats(t, `counted: "2026-10-05"
templates:
  demo: {added: "2026-09-01", repo: acme/demo, stars: 12, starsGained: 3}
`))
}

func TestLintChecksStats(t *testing.T) {
	for want, demo := range map[string]string{
		`demo: added "Sept 1" is not a day`:         `{added: Sept 1}`,
		`demo: repo "https://github.com/acme/demo"`: `{added: "2026-09-01", repo: https://github.com/acme/demo}`,
		"demo: a count cannot be negative":          `{added: "2026-09-01", stars: -4}`,
		"field surprise not found":                  `{added: "2026-09-01", surprise: 1}`,
	} {
		assert.Contains(t, lintWithStats(t, "templates:\n  demo: "+demo+"\n"), want)
	}
	for want, stats := range map[string]string{
		`counted "last week" is not a day`: "counted: last week\ntemplates: {}\n",
	} {
		assert.Contains(t, lintWithStats(t, stats), want)
	}
}
