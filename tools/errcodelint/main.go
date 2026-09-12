// Command errcodelint keeps error codes and their translations in sync.
//
// Error codes are plain strings by the time they reach the translation layer,
// so nothing in the compiler connects hperrors.NewErr(_, "ERR_X") to the
// message file that gives ERR_X its text. Codes are also declared per package
// now rather than centrally, which makes a forgotten message file entry easy.
//
// It reports four things:
//
//  1. a code used in Go with no translation anywhere
//  2. an id defined in two different message files - the loader's bundle is a
//     plain map, so the file loaded last wins and nothing says so. A repeat
//     inside one file is not checked: TOML rejects it and the translation
//     package panics at startup.
//  3. a translation no Go source mentions, except the ERR_VLD_ prefix, which
//     the validation library generates at runtime
//  5. a code declared as a constant nothing else refers to. Such a code cannot
//     be raised, and its translation ships for a message no caller produces.
//  4. a message file whose name does not yield a resolvable language. go-i18n
//     reads the segment before the extension as the language tag and
//     language.Make answers the undefined tag rather than failing, so
//     "logging.errors.toml" loads under a language nothing matches while the
//     build stays green.
//
// Usage:
//
//	go run ./tools/errcodelint [-messages dir] [dir...]
package main

import (
	"flag"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
)

const defaultMessagesDir = "hivepaas_app/pkg/translation/messages"

func main() {
	messagesDir := flag.String("messages", defaultMessagesDir, "root of the translation message files")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	fset := token.NewFileSet()
	uses, err := scanGoDirs(fset, roots)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errcodelint: %v\n", err)
		os.Exit(2)
	}

	entries, err := scanMessageFiles(*messagesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errcodelint: %v\n", err)
		os.Exit(2)
	}

	fileNames, err := messageFileNames(*messagesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errcodelint: %v\n", err)
		os.Exit(2)
	}

	declaredIdents := map[string]bool{}
	for _, u := range uses {
		if u.Declared && u.Ident != "" {
			declaredIdents[u.Ident] = true
		}
	}
	identUses, err := scanIdentUses(fset, roots, declaredIdents)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errcodelint: %v\n", err)
		os.Exit(2)
	}

	findings := checkAll(uses, entries, fileNames, identUses)
	for _, f := range findings {
		fmt.Printf("%s: %s\n", f.Where, f.Msg)
	}
	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d problem(s).\n", len(findings))
		os.Exit(1)
	}
}

// messageFileNames lists the .toml files under root, for the name check.
func messageFileNames(root string) ([]string, error) {
	var names []string
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
		if filepath.Ext(path) == ".toml" {
			names = append(names, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}
