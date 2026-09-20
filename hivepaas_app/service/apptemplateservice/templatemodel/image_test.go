package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	pgLine  = "18"
	pgImage = "postgres:18.6-alpine3.24"
	hex64   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

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
			class, err := ClassifyImageOverride(pgLine, pgImage, tc.override)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, class)
		})
	}
}

// The line is what the template declares, not a rule about segments: MariaDB's
// 11.8 line does not cover 11.9, while a template declaring line 2 covers 2.2.0.
func TestClassifyImageOverrideReadsTheDeclaredLine(t *testing.T) {
	for _, tc := range []struct {
		line, image, override string
		want                  ImageOverrideClass
	}{
		{"11.8", "mariadb:11.8.9-noble", "mariadb:11.8.10-noble", ImageOverrideSameLine},
		{"11.8", "mariadb:11.8.9-noble", "mariadb:11.9.1-noble", ImageOverrideOtherMajor},
		{"11.8", "mariadb:11.8.9-noble", "mariadb:11.80.1-noble", ImageOverrideOtherMajor},
		{"2", "demo:2.1.0", "demo:2.2.0", ImageOverrideSameLine},
		{"2", "demo:2.1.0", "demo:v2.3.0", ImageOverrideSameLine},
		{"2", "demo:2.1.0", "demo:3.0.0", ImageOverrideOtherMajor},
		{"", "demo:2.1.0", "demo:2.2.0", ImageOverrideOtherMajor},
	} {
		class, err := ClassifyImageOverride(tc.line, tc.image, tc.override)
		assert.NoError(t, err, tc.override)
		assert.Equal(t, tc.want, class, "line %q, %s", tc.line, tc.override)
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
			_, err := ClassifyImageOverride(pgLine, pgImage, override)
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

	got := SelectTagCandidates(pgLine, pgImage, tags, 0)

	assert.Equal(t, []TagCandidate{
		{Tag: "19.0-alpine3.24", Class: ImageOverrideOtherMajor, Newer: true},
		{Tag: "18.7-alpine3.25", Class: ImageOverrideSameLine, Newer: true},
		{Tag: "18.7-alpine3.24", Class: ImageOverrideSameLine, Newer: true},
		{Tag: "18.5-alpine3.22", Class: ImageOverrideSameLine, Newer: false},
		{Tag: "17.11-alpine3.24", Class: ImageOverrideOtherMajor, Newer: false},
	}, got)
}

// Same release, different base image: nothing orders those two, and the list must
// not claim the older base is an upgrade.
func TestSelectTagCandidatesDoesNotCallAnUnorderableTagNewer(t *testing.T) {
	got := SelectTagCandidates(pgLine, pgImage, []string{"18.6-alpine3.23", "18.6-alpine", "18.7-alpine3.24"}, 0)

	newerByTag := map[string]bool{}
	for _, candidate := range got {
		newerByTag[candidate.Tag] = candidate.Newer
	}
	assert.Equal(t, map[string]bool{
		"18.7-alpine3.24": true,
		"18.6-alpine3.23": false,
		"18.6-alpine":     false,
	}, newerByTag)
}

func TestSelectTagCandidatesCapsAndDeduplicates(t *testing.T) {
	tags := []string{"18.7-alpine3.24", "18.7-alpine3.24", "18.8-alpine3.24", "18.9-alpine3.24"}

	got := SelectTagCandidates(pgLine, pgImage, tags, 2)

	assert.Equal(t, []string{"18.9-alpine3.24", "18.8-alpine3.24"}, []string{got[0].Tag, got[1].Tag})
	assert.Len(t, got, 2)
}

func TestSelectTagCandidatesWithoutASuffixFamily(t *testing.T) {
	got := SelectTagCandidates("11.8", "mariadb:11.8.9-noble", []string{"11.8.10-noble", "11.8.10-ubi", "latest"}, 0)

	assert.Equal(t, []TagCandidate{{Tag: "11.8.10-noble", Class: ImageOverrideSameLine, Newer: true}}, got)
}

// A Postgres extension image names the server it is built for in its tag. Another
// major of PostgreSQL is a different product to the app running on this one - it
// will not open the data directory - so it must not be offered as another build.
func TestSelectTagCandidatesKeepsThePostgresMajorApart(t *testing.T) {
	tags := []string{
		"0.8.7-pg18-bookworm", "0.8.7-pg17-bookworm", "0.8.7-pg18-trixie",
		"0.8.6-pg17-bookworm", "0.9.0-pg18-bookworm", "pg18-bookworm", "pg18",
	}

	got := SelectTagCandidates("0.8", "pgvector/pgvector:0.8.6-pg18-bookworm", tags, 0)

	assert.Equal(t, []TagCandidate{
		{Tag: "0.9.0-pg18-bookworm", Class: ImageOverrideOtherMajor, Newer: true},
		{Tag: "0.8.7-pg18-bookworm", Class: ImageOverrideSameLine, Newer: true},
	}, got)
}

// The rule is narrow on purpose: a base image still moves on its own schedule,
// and alpine3.24 is another build of alpine3.22.
func TestTagFamily(t *testing.T) {
	tests := map[string]string{
		"18.6-alpine3.24":     "alpine",
		"11.8.9-noble":        "noble",
		"18.6":                "",
		"0.8.6-pg18-bookworm": "pg18-bookworm",
		"2.30.1-pg17":         "pg17",
		"2.30.1-pg17-oss":     "pg17-oss",
	}

	for tag, want := range tests {
		assert.Equal(t, want, tagFamily(tag), tag)
	}
}

func TestClassifyImageOverrideAcceptsTheSameRepositoryWrittenDifferently(t *testing.T) {
	// A person pasting from a registry writes the canonical form; a template
	// writes the short one. Refusing that is refusing the template's own image.
	class, err := ClassifyImageOverride("18", "postgres:18.6-alpine3.24",
		"registry-1.docker.io/library/postgres:18.6-alpine3.23")
	assert.NoError(t, err)
	assert.Equal(t, ImageOverrideSameLine, class)

	class, err = ClassifyImageOverride("18", "postgres:18.6-alpine3.24", "docker.io/library/postgres:17.11-alpine3.24")
	assert.NoError(t, err)
	assert.Equal(t, ImageOverrideOtherMajor, class)

	// And a different repository is still refused, canonical or not.
	_, err = ClassifyImageOverride("18", "postgres:18.6-alpine3.24", "docker.io/library/mysql:8.4.11")
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateImageNotAllowed)
	_, err = ClassifyImageOverride("18", "postgres:18.6-alpine3.24", "ghcr.io/library/postgres:18.6")
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateImageNotAllowed)
}
