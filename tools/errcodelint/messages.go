package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/language"
)

// messageEntry is one translated id and the file it came from.
type messageEntry struct {
	ID   string
	File string
}

// scanMessageFiles reads every .toml under root, the same recursive walk the
// translation loader performs.
//
// A duplicate id inside a single file is not checked for: TOML forbids it and
// toml.DecodeFile below returns an error, which is also what happens to the
// running program, loudly, during package initialisation.
func scanMessageFiles(root string) ([]messageEntry, error) {
	var entries []messageEntry
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".toml" {
			return nil
		}

		var m map[string]any
		if _, err := toml.DecodeFile(path, &m); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		ids := make([]string, 0, len(m))
		for id := range m {
			ids = append(ids, id)
		}
		sort.Strings(ids) // map order is random; keep output stable
		for _, id := range ids {
			entries = append(entries, messageEntry{ID: id, File: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// languageOfFile reports the language tag go-i18n will derive from a message
// file name, and whether that tag is one it can actually resolve.
//
// go-i18n takes the segment before the extension as the tag. language.Make
// answers the undefined tag for anything it cannot parse instead of failing, so
// a misnamed file loads its messages under a language nothing matches.
func languageOfFile(name string) (string, bool) {
	parts := strings.Split(name, ".")
	if len(parts) < 3 {
		// <name>.toml has no language segment at all.
		return "", false
	}
	tag := parts[len(parts)-2]
	parsed, err := language.Parse(tag)
	if err != nil || parsed == language.Und {
		return tag, false
	}
	return tag, true
}
