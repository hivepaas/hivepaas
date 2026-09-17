package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
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

// Same release, different base image: nothing orders those two, and the list must
// not claim the older base is an upgrade.
func TestSelectTagCandidatesDoesNotCallAnUnorderableTagNewer(t *testing.T) {
	got := SelectTagCandidates(pgImage, []string{"18.6-alpine3.23", "18.6-alpine", "18.7-alpine3.24"}, 0)

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

	got := SelectTagCandidates(pgImage, tags, 2)

	assert.Equal(t, []string{"18.9-alpine3.24", "18.8-alpine3.24"}, []string{got[0].Tag, got[1].Tag})
	assert.Len(t, got, 2)
}

func TestSelectTagCandidatesWithoutASuffixFamily(t *testing.T) {
	got := SelectTagCandidates("mariadb:11.8.9-noble", []string{"11.8.10-noble", "11.8.10-ubi", "latest"}, 0)

	assert.Equal(t, []TagCandidate{{Tag: "11.8.10-noble", Class: ImageOverrideSameLine, Newer: true}}, got)
}
