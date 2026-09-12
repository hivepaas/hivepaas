package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanMessageFilesCollectsIDsWithTheirFile(t *testing.T) {
	entries, err := scanMessageFiles("testdata/messages")
	if err != nil {
		t.Fatalf("scanMessageFiles: %v", err)
	}

	got := map[string]string{}
	for _, e := range entries {
		got[e.ID] = filepath.Base(e.File)
	}

	assert.Equal(t, map[string]string{
		"ERR_BASE":            "errors.sample.en.toml",
		"ERR_DERIVED":         "errors.sample.en.toml",
		"ERR_REFERENCED_ONLY": "errors.other.en.toml",
		"ERR_ORPHAN":          "errors.other.en.toml",
	}, got)
}

func TestLanguageOfFile(t *testing.T) {
	cases := []struct {
		name   string
		lang   string
		wantOK bool
	}{
		{"errors.logging.en.toml", "en", true},
		{"errors.en.toml", "en", true},
		{"validation_errors.en.toml", "en", true},
		// The trap: no language segment, so go-i18n reads "errors" as the tag
		// and language.Make quietly returns the undefined tag.
		{"logging.errors.toml", "errors", false},
		{"errors.toml", "", false},
	}
	for _, c := range cases {
		lang, ok := languageOfFile(c.name)
		assert.Equal(t, c.wantOK, ok, c.name)
		if c.wantOK {
			assert.Equal(t, c.lang, lang, c.name)
		}
	}
}
