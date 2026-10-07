package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const indexAtCommit = `{"templates":[]}`

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// raw is GitHub's raw files: what each URL answers.
type raw map[string]string

func (r raw) fetch(url string) ([]byte, error) {
	if body, ok := r[url]; ok {
		return []byte(body), nil
	}
	return nil, errors.New("404 Not Found: " + url)
}

const commit = "32c7af0821c421760f70ec33ce5b943841254d20"

var github = raw{
	"https://raw.githubusercontent.com/hivepaas/app-templates/" + commit + "/index.json": indexAtCommit,
}

func releaseWithTemplates(indexSHA string) []byte {
	return []byte(`{
  "beta": {
    "appVersion": "v1.0.0-beta3",
    "templates": {
      "repo": "hivepaas/app-templates",
      "commit": "` + commit + `",
      "indexSha256": "` + indexSHA + `"
    }
  },
  "stable": {
    "appVersion": "v0.9.0"
  }
}`)
}

// The pin is what a server checks the index it fetches against: the index at
// the pinned commit hashes to it.
func TestATemplatesPinThatMatchesItsIndexPasses(t *testing.T) {
	checked, err := checkTemplates(releaseWithTemplates(sha256Hex(indexAtCommit)), github.fetch)

	assert.NoError(t, err)
	assert.Equal(t, []string{"beta: hivepaas/app-templates@" + commit}, checked)
}

// A pin whose hash is not the index's would leave every server that updates
// with no templates: "index.json does not match its sha256".
func TestATemplatesPinThatDoesNotMatchItsIndexFails(t *testing.T) {
	_, err := checkTemplates(releaseWithTemplates(sha256Hex("another index")), github.fetch)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "beta")
		assert.Contains(t, err.Error(), commit)
		assert.Contains(t, err.Error(), sha256Hex(indexAtCommit))
	}
}

// A commit GitHub does not have - mistyped, or never pushed - cannot be checked,
// and a server could not fetch it either.
func TestATemplatesPinOfACommitThatCannotBeFetchedFails(t *testing.T) {
	_, err := checkTemplates(releaseWithTemplates(sha256Hex(indexAtCommit)), raw{}.fetch)

	assert.ErrorContains(t, err, "404")
}

// A release that offers no templates has nothing to check.
func TestAReleaseWithoutTemplatesHasNothingToCheck(t *testing.T) {
	checked, err := checkTemplates([]byte(release), github.fetch)

	assert.NoError(t, err)
	assert.Empty(t, checked)
}
