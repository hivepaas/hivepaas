# Error code lint and per-domain message files Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a lint that keeps error codes and their translations in sync, then split the translation files per domain so a feature's errors live beside the code that raises them.

**Architecture:** A new `tools/errcodelint` command modelled on the existing `tools/goroutinelint`: walk the repo with `go/ast`, collect every error code declared or referenced in Go, parse the message files with `github.com/BurntSushi/toml`, and report mismatches. Wire it into `make lint`. Only once it is green do the message files get split, so the lint is what proves the move dropped and duplicated nothing.

**Tech Stack:** Go, `go/ast`, `go/parser`, `github.com/BurntSushi/toml` (already vendored), `golang.org/x/text/language` (already vendored).

**Spec:** `docs/superpowers/specs/2026-09-12-logging-design.md`, section 4.

## Global Constraints

- Run every Go command from the repo root, `/Users/tnt/go/src/github.com/hivepaas/hivepaas`.
- The repo vendors its dependencies. Only `github.com/stretchr/testify/assert` is vendored - **`testify/require` is NOT**. Use `assert`, and a bare `t.Fatal`/`t.Fatalf` where execution must stop.
- Never add a dependency. Everything this plan needs is already vendored.
- `make lint` runs `golangci-lint run -v ./...`; new code must pass it. British/American spelling is checked by the `misspell` linter - use American spelling.
- Error code identifiers match `^ERR_[A-Z0-9_]+$`.
- Message files live in `hivepaas_app/pkg/translation/messages/<lang>/` and are loaded by a recursive directory walk, so any new file in that tree is picked up with no code change.

---

## File Structure

| File | Responsibility |
|---|---|
| `tools/errcodelint/main.go` | Command entry, flags, report, exit code |
| `tools/errcodelint/gocodes.go` | Scan Go sources for declared and referenced codes |
| `tools/errcodelint/messages.go` | Parse message files; collect id → file, and validate file names |
| `tools/errcodelint/checks.go` | The four checks, operating on the two collections |
| `tools/errcodelint/testdata/` | Fixture Go package and message files for the tests |
| `Makefile` | Add the tool to the `lint` target |
| `hivepaas_app/pkg/translation/messages/en/errors.*.en.toml` | The split files |

---

### Task 1: Scan Go sources for error codes

**Files:**
- Create: `tools/errcodelint/gocodes.go`
- Create: `tools/errcodelint/gocodes_test.go`
- Create: `tools/errcodelint/testdata/gosample/sample.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type codeUse struct { Code string; Pos token.Position; Declared bool }`
  - `func scanGoDirs(fset *token.FileSet, roots []string) ([]codeUse, error)`
  - `func skipDir(name string) bool`

Context: error codes reach the translation layer as plain strings. They appear in Go three ways, and all three must be found or the lint reports false problems:

1. `hperrors.NewErr(hperrors.ErrBadRequest, "ERR_X")` - a declaration.
2. `errors.New("ERR_INTERNAL")` inside `hivepaas_app/hperrors` - also a declaration; ten base errors are written this way.
3. A bare string literal such as `vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED")` - a reference, not a declaration.

The simplest rule that covers all three: **every string literal matching the code pattern is a use**, and a use is additionally a *declaration* when it is an argument to a call of `NewErr` or `errors.New`. Nothing else needs to know the difference.

- [ ] **Step 1: Write the failing test**

Create `tools/errcodelint/testdata/gosample/sample.go`. It is fixture data, not compiled by the build (`testdata` is ignored by the go tool):

```go
package gosample

import (
	"errors"
)

func NewErr(base error, s string) error { return errors.Join(base, errors.New(s)) }

var (
	ErrBase    = errors.New("ERR_BASE")
	ErrDerived = NewErr(ErrBase, "ERR_DERIVED")
)

func use() string {
	return setCustomKey("ERR_REFERENCED_ONLY")
}

func setCustomKey(s string) string { return s }

// notACode is here to prove the scanner is not matching every string.
const notACode = "hello world"
```

Create `tools/errcodelint/gocodes_test.go`:

```go
package main

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanGoDirsFindsDeclarationsAndReferences(t *testing.T) {
	fset := token.NewFileSet()

	uses, err := scanGoDirs(fset, []string{"testdata/gosample"})
	if err != nil {
		t.Fatalf("scanGoDirs: %v", err)
	}

	got := map[string]bool{} // code -> declared
	for _, u := range uses {
		// A code declared anywhere counts as declared.
		got[u.Code] = got[u.Code] || u.Declared
	}

	assert.Equal(t, map[string]bool{
		"ERR_BASE":            true,  // errors.New
		"ERR_DERIVED":         true,  // NewErr
		"ERR_REFERENCED_ONLY": false, // plain string literal
	}, got)
}

func TestScanGoDirsRecordsPosition(t *testing.T) {
	fset := token.NewFileSet()

	uses, err := scanGoDirs(fset, []string{"testdata/gosample"})
	if err != nil {
		t.Fatalf("scanGoDirs: %v", err)
	}

	for _, u := range uses {
		if u.Code == "ERR_DERIVED" {
			assert.Contains(t, u.Pos.Filename, "sample.go")
			assert.Positive(t, u.Pos.Line)
			return
		}
	}
	t.Fatal("ERR_DERIVED not found")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./tools/errcodelint/ -run TestScanGoDirs -v`
Expected: FAIL to build, `undefined: scanGoDirs`.

- [ ] **Step 3: Write the implementation**

Create `tools/errcodelint/gocodes.go`:

```go
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// codePattern is what an error code identifier looks like. Anything else in a
// string literal is not our business.
var codePattern = regexp.MustCompile(`^ERR_[A-Z0-9_]+$`)

// declaringFuncs are the calls whose string argument declares a code rather
// than merely mentioning one.
var declaringFuncs = map[string]bool{
	"NewErr": true,
	"New":    true, // errors.New, used for the base errors in hperrors
}

// skipDirs are never scanned. Mirrors tools/goroutinelint.
var skipDirs = map[string]bool{
	"vendor":         true,
	"node_modules":   true,
	"tmp":            true,
	"temp":           true,
	"deployment":     true,
	"dist-dashboard": true,
	"test-results":   true,
}

// codeUse is one appearance of an error code in Go source.
type codeUse struct {
	Code     string
	Pos      token.Position
	Declared bool
}

func skipDir(name string) bool {
	return skipDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanGoDirs collects every error code appearing in Go sources under roots.
//
// testdata is scanned here, unlike in goroutinelint: this tool's own fixtures
// live there, and a stray code in someone else's testdata is worth reporting
// rather than hiding.
func scanGoDirs(fset *token.FileSet, roots []string) ([]codeUse, error) {
	var uses []codeUse
	for _, root := range roots {
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
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			uses = append(uses, codesInFile(fset, file)...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return uses, nil
}

// codesInFile walks one parsed file.
//
// Declarations are collected first so that the literal inside a declaring call
// is not also counted as a bare reference.
func codesInFile(fset *token.FileSet, file *ast.File) []codeUse {
	declared := map[*ast.BasicLit]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !declaringFuncs[calleeName(call.Fun)] {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && codePattern.MatchString(litValue(lit)) {
				declared[lit] = true
			}
		}
		return true
	})

	var uses []codeUse
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		code := litValue(lit)
		if !codePattern.MatchString(code) {
			return true
		}
		uses = append(uses, codeUse{
			Code:     code,
			Pos:      fset.Position(lit.Pos()),
			Declared: declared[lit],
		})
		return true
	})
	return uses
}

// calleeName is the function name of a call, ignoring any package qualifier, so
// that NewErr and hperrors.NewErr both answer "NewErr".
func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// litValue is the unquoted contents of a string literal, or "" for anything else.
func litValue(lit *ast.BasicLit) string {
	if lit.Kind != token.STRING {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./tools/errcodelint/ -run TestScanGoDirs -v`
Expected: PASS, both tests.

- [ ] **Step 5: Prove the test can fail**

Temporarily change `codePattern` to `^ERR_BASE$`, re-run, and confirm `TestScanGoDirsFindsDeclarationsAndReferences` fails. Restore the pattern and confirm it passes again.

- [ ] **Step 6: Commit**

```bash
git add tools/errcodelint/gocodes.go tools/errcodelint/gocodes_test.go tools/errcodelint/testdata/gosample/sample.go
git commit -m "feat(errcodelint): scan go sources for error codes"
```

---

### Task 2: Parse message files and validate their names

**Files:**
- Create: `tools/errcodelint/messages.go`
- Create: `tools/errcodelint/messages_test.go`
- Create: `tools/errcodelint/testdata/messages/en/errors.sample.en.toml`
- Create: `tools/errcodelint/testdata/messages/en/errors.other.en.toml`

**Interfaces:**
- Consumes: `skipDir` from Task 1.
- Produces:
  - `type messageEntry struct { ID string; File string }`
  - `func scanMessageFiles(root string) ([]messageEntry, error)`
  - `func languageOfFile(name string) (string, bool)`

Context: the loader (`hivepaas_app/pkg/translation/loading.go`) walks the message directory recursively and hands each file to go-i18n, which derives the language from the file name. `parsePath` scans backwards from the end: the last dot-separated segment is the format, the one before it is the language tag. `language.Make` then turns an unparseable tag into the undefined tag **without reporting an error**, so a file named `logging.errors.toml` would load every message under a language nothing resolves against while the build and the tests stay green. That is what `languageOfFile` guards.

A repeated id inside one file needs no check here: TOML forbids it, `toml.Decode` returns `Key 'X' has already been defined`, and the translation package panics on that during initialisation. Only the cross-file case is silent, and that is Task 3's check.

- [ ] **Step 1: Write the failing test**

Create `tools/errcodelint/testdata/messages/en/errors.sample.en.toml`:

```toml
# Sample errors
ERR_BASE = "base"
ERR_DERIVED = "derived"
```

Create `tools/errcodelint/testdata/messages/en/errors.other.en.toml`:

```toml
# Other errors
ERR_REFERENCED_ONLY = "referenced only"
ERR_ORPHAN = "nothing in go mentions this"
```

Create `tools/errcodelint/messages_test.go`:

```go
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
		name    string
		lang    string
		wantOK  bool
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./tools/errcodelint/ -run 'TestScanMessageFiles|TestLanguageOfFile' -v`
Expected: FAIL to build, `undefined: scanMessageFiles`.

- [ ] **Step 3: Write the implementation**

Create `tools/errcodelint/messages.go`:

```go
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
// toml.Decode below returns an error, which is also what happens to the running
// program, loudly, during package initialisation.
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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./tools/errcodelint/ -run 'TestScanMessageFiles|TestLanguageOfFile' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add tools/errcodelint/messages.go tools/errcodelint/messages_test.go tools/errcodelint/testdata/messages
git commit -m "feat(errcodelint): read message files and validate their names"
```

---

### Task 3: The four checks

**Files:**
- Create: `tools/errcodelint/checks.go`
- Create: `tools/errcodelint/checks_test.go`

**Interfaces:**
- Consumes: `codeUse` and `scanGoDirs` from Task 1; `messageEntry`, `scanMessageFiles`, `languageOfFile` from Task 2.
- Produces:
  - `type finding struct { Where string; Msg string }`
  - `func check(uses []codeUse, entries []messageEntry, fileNames []string) []finding`
  - `var orphanExemptPrefixes = []string{"ERR_VLD_"}`

Context on the exemption: `ERR_VLD_*` ids are produced by the validation library at runtime and reach translation through `vld.SetCustomKey`. Several, such as `ERR_VLD_FIELD_REQUIRED`, appear in no Go source at all - the library generates them from validator rule names. Reporting those 44 ids as orphans would make the check useless on its first run, so the prefix is exempt and the exemption is named, not silent.

- [ ] **Step 1: Write the failing test**

Create `tools/errcodelint/checks_test.go`:

```go
package main

import (
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func msgs(findings []finding) string {
	var b strings.Builder
	for _, f := range findings {
		b.WriteString(f.Where + ": " + f.Msg + "\n")
	}
	return b.String()
}

func TestCheckReportsCodeWithNoTranslation(t *testing.T) {
	uses := []codeUse{{Code: "ERR_MISSING", Pos: token.Position{Filename: "a.go", Line: 3}, Declared: true}}

	out := msgs(check(uses, nil, nil))

	assert.Contains(t, out, "ERR_MISSING")
	assert.Contains(t, out, "no translation")
}

func TestCheckReportsIDDefinedInTwoFiles(t *testing.T) {
	uses := []codeUse{{Code: "ERR_DUP", Declared: true}}
	entries := []messageEntry{
		{ID: "ERR_DUP", File: "messages/en/errors.a.en.toml"},
		{ID: "ERR_DUP", File: "messages/en/errors.b.en.toml"},
	}

	out := msgs(check(uses, entries, nil))

	assert.Contains(t, out, "ERR_DUP")
	assert.Contains(t, out, "errors.a.en.toml")
	assert.Contains(t, out, "errors.b.en.toml")
}

func TestCheckReportsOrphanTranslation(t *testing.T) {
	entries := []messageEntry{{ID: "ERR_ORPHAN", File: "messages/en/errors.a.en.toml"}}

	out := msgs(check(nil, entries, nil))

	assert.Contains(t, out, "ERR_ORPHAN")
	assert.Contains(t, out, "no Go source")
}

// The validation library generates these ids at runtime, so they are translated
// without ever appearing in Go.
func TestCheckExemptsValidationPrefixFromOrphans(t *testing.T) {
	entries := []messageEntry{{ID: "ERR_VLD_FIELD_REQUIRED", File: "messages/en/errors.validation.en.toml"}}

	assert.Empty(t, check(nil, entries, nil))
}

func TestCheckReportsUnparseableFileName(t *testing.T) {
	out := msgs(check(nil, nil, []string{"messages/en/logging.errors.toml"}))

	assert.Contains(t, out, "logging.errors.toml")
	assert.Contains(t, out, "language")
}

// A code merely referenced as a string still has to be translated, but it is
// not an orphan when Go mentions it.
func TestCheckAcceptsReferencedCodeWithTranslation(t *testing.T) {
	uses := []codeUse{{Code: "ERR_REF", Declared: false}}
	entries := []messageEntry{{ID: "ERR_REF", File: "messages/en/errors.a.en.toml"}}

	assert.Empty(t, check(uses, entries, nil))
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./tools/errcodelint/ -run TestCheck -v`
Expected: FAIL to build, `undefined: check`.

- [ ] **Step 3: Write the implementation**

Create `tools/errcodelint/checks.go`:

```go
package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// orphanExemptPrefixes are ids that legitimately have no Go source.
//
// ERR_VLD_* is generated by the validation library from validator rule names
// and reaches translation through vld.SetCustomKey at runtime.
var orphanExemptPrefixes = []string{"ERR_VLD_"}

// finding is one problem to report.
type finding struct {
	Where string
	Msg   string
}

func check(uses []codeUse, entries []messageEntry, fileNames []string) []finding {
	var findings []finding

	translated := map[string][]string{} // id -> files defining it
	for _, e := range entries {
		translated[e.ID] = append(translated[e.ID], e.File)
	}

	// 1. every code used in Go has a translation.
	firstUse := map[string]codeUse{}
	for _, u := range uses {
		if _, seen := firstUse[u.Code]; !seen {
			firstUse[u.Code] = u
		}
	}
	for _, code := range sortedKeys(firstUse) {
		if len(translated[code]) > 0 {
			continue
		}
		u := firstUse[code]
		where := u.Pos.Filename
		if u.Pos.Line > 0 {
			where = fmt.Sprintf("%s:%d", u.Pos.Filename, u.Pos.Line)
		}
		findings = append(findings, finding{
			Where: where,
			Msg:   fmt.Sprintf("%s has no translation in any message file", code),
		})
	}

	// 2. no id is defined in two different files.
	for _, id := range sortedKeys(translated) {
		files := uniqueSorted(translated[id])
		if len(files) < 2 {
			continue
		}
		findings = append(findings, finding{
			Where: files[0],
			Msg: fmt.Sprintf("%s is defined in %d files (%s); the last one loaded wins, silently",
				id, len(files), strings.Join(files, ", ")),
		})
	}

	// 3. no translation is orphaned.
	used := map[string]bool{}
	for _, u := range uses {
		used[u.Code] = true
	}
	for _, id := range sortedKeys(translated) {
		if used[id] || exemptFromOrphanCheck(id) {
			continue
		}
		findings = append(findings, finding{
			Where: translated[id][0],
			Msg:   fmt.Sprintf("%s is translated but appears in no Go source", id),
		})
	}

	// 4. every message file name yields a language go-i18n can resolve.
	for _, name := range fileNames {
		if tag, ok := languageOfFile(filepath.Base(name)); !ok {
			findings = append(findings, finding{
				Where: name,
				Msg: fmt.Sprintf("file name gives language %q, which does not resolve;"+
					" name it <domain>.<lang>.toml", tag),
			})
		}
	}

	return findings
}

func exemptFromOrphanCheck(id string) bool {
	for _, p := range orphanExemptPrefixes {
		if strings.HasPrefix(id, p) {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func uniqueSorted(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./tools/errcodelint/ -run TestCheck -v`
Expected: PASS, all six tests.

- [ ] **Step 5: Commit**

```bash
git add tools/errcodelint/checks.go tools/errcodelint/checks_test.go
git commit -m "feat(errcodelint): the four consistency checks"
```

---

### Task 4: Command entry point, and make it green on this repo

**Files:**
- Create: `tools/errcodelint/main.go`
- Modify: `Makefile:21-23` (the `lint` target)
- Modify: whichever message or Go files the first real run reports

**Interfaces:**
- Consumes: everything from Tasks 1-3.
- Produces: `go run ./tools/errcodelint` exiting 0 when clean, 1 with findings, 2 on an internal error.

- [ ] **Step 1: Write the command**

Create `tools/errcodelint/main.go`:

```go
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
	"sort"
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

	findings := check(uses, entries, fileNames)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Where != findings[j].Where {
			return findings[i].Where < findings[j].Where
		}
		return findings[i].Msg < findings[j].Msg
	})
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
```

- [ ] **Step 2: Run it on the repo and record what it says**

Run: `go run ./tools/errcodelint`

This is the first real run and it is expected to report problems. Save the output - it is the input to Step 3:

```bash
go run ./tools/errcodelint > /tmp/errcodelint-first-run.txt 2>&1; echo "exit=$?"
cat /tmp/errcodelint-first-run.txt
```

Baseline for orientation: 161 codes are declared with `NewErr` in `hivepaas_app/hperrors/constants.go`, 10 more with `errors.New` there, 15 in packages under `services/`, and the message files hold 179 + 44 ids. The counts do not match exactly, so findings are expected.

- [ ] **Step 3: Resolve every finding**

For each line, decide and act:

- **"has no translation"** - add the message. Put it in the section of `errors.en.toml` matching its domain, with the wording style of its neighbours: a sentence, `{{.Name}}` for an interpolated value.
- **"appears in no Go source"** - the code is dead. Delete the message line. If it looks like it should still be used, check for a typo between the Go string and the message id first; a mismatched pair shows up as one of each kind of finding.
- **"defined in N files"** - should not occur before the split; if it does, delete the copy in the file where it does not belong.
- **"does not resolve"** - rename the file to `<domain>.<lang>.toml`.

Do not silence a finding by widening `orphanExemptPrefixes`. That list has one entry and a reason; adding to it needs the same kind of reason.

- [ ] **Step 4: Verify it is clean**

Run: `go run ./tools/errcodelint; echo "exit=$?"`
Expected: no output, `exit=0`.

- [ ] **Step 5: Confirm translation still initialises**

The translation package panics during initialisation on a malformed message file, so any test that imports it proves the files still load:

Run: `go test ./hivepaas_app/pkg/translation/... -count=1`
Expected: PASS.

- [ ] **Step 6: Wire it into make lint**

Modify the `lint` target in `Makefile` so the new tool runs beside `goroutinelint`:

```make
lint:
	$(DEVTOOLS_CMD) go run ./tools/goroutinelint .
	$(DEVTOOLS_CMD) go run ./tools/errcodelint
	$(DEVTOOLS_CMD) golangci-lint --timeout=3m run -v ./...
```

- [ ] **Step 7: Check the tool passes the project's own linter**

Run: `golangci-lint --timeout=3m run ./tools/errcodelint/...`
Expected: `0 issues.`

- [ ] **Step 8: Commit**

```bash
git add tools/errcodelint Makefile hivepaas_app/pkg/translation/messages
git commit -m "feat(errcodelint): command entry, wired into make lint"
```

---

### Task 5: Split the message files per domain

**Files:**
- Modify: `hivepaas_app/pkg/translation/messages/en/errors.en.toml`
- Create: `hivepaas_app/pkg/translation/messages/en/errors.backup.en.toml`
- Rename: `validation_errors.en.toml` → `errors.validation.en.toml`

**Interfaces:**
- Consumes: a clean `go run ./tools/errcodelint` from Task 4.
- Produces: message files split by domain, still loading identically.

Context: the loader walks the directory recursively and parses every `.toml` it finds, so new files need no code change. This task moves lines between files and changes nothing else. The lint from Task 4 is what proves it: a dropped id becomes a "no translation" finding, a copied one becomes "defined in 2 files".

- [ ] **Step 1: Record the before state**

```bash
go run ./tools/errcodelint; echo "lint exit=$?"
grep -chE '^ERR_' hivepaas_app/pkg/translation/messages/en/*.toml | paste -sd+ | bc
```

Write the total down. It must be identical at the end.

- [ ] **Step 2: Move the backup section into its own file**

`errors.en.toml` has a `# Backup errors` section around line 169. Move those lines, in order, into a new `hivepaas_app/pkg/translation/messages/en/errors.backup.en.toml`, with this header:

```toml
# Backup errors
#
# Not declared in hperrors/constants.go. The source is:
#   services/backup/backupmodel/errors.go   (ERR_BACKUP_*)
# Adding or removing an error there means editing this file.
```

Delete the section, and its comment, from `errors.en.toml`.

- [ ] **Step 3: Verify nothing moved twice or got lost**

Run: `go run ./tools/errcodelint; echo "exit=$?"`
Expected: no output, `exit=0`. A missed line reports "has no translation"; a copied line reports "defined in 2 files".

- [ ] **Step 4: Rename the validation file**

```bash
git mv hivepaas_app/pkg/translation/messages/en/validation_errors.en.toml \
       hivepaas_app/pkg/translation/messages/en/errors.validation.en.toml
```

Add this header at the top of the renamed file, above the existing content:

```toml
# Validation errors
#
# These ids are generated by the validation library from validator rule names
# and reach translation through vld.SetCustomKey, so most appear in no Go
# source. tools/errcodelint exempts the ERR_VLD_ prefix from its orphan check
# for that reason.
```

- [ ] **Step 5: Verify the rename changed nothing**

```bash
go run ./tools/errcodelint; echo "lint exit=$?"
go test ./hivepaas_app/pkg/translation/... -count=1
grep -chE '^ERR_' hivepaas_app/pkg/translation/messages/en/*.toml | paste -sd+ | bc
```

Expected: lint `exit=0`, tests PASS, and the total identical to Step 1. The rename is also the first real exercise of check 4 - `errors.validation.en.toml` resolves, where a name like `validation.errors.toml` would not.

- [ ] **Step 6: Run the full suite**

Run: `./scripts/test.sh 2>&1 | grep -E "FAIL|DONE\.|panic"`
Expected: `DONE.` and no `FAIL`.

- [ ] **Step 7: Commit**

```bash
git add hivepaas_app/pkg/translation/messages
git commit -m "refactor(i18n): split message files per domain"
```

---

## Verification

After Task 5, all of these hold:

```bash
go run ./tools/errcodelint; echo $?          # 0, no output
go test ./tools/errcodelint/ -count=1        # PASS
make lint                                     # passes, errcodelint included
./scripts/test.sh                             # DONE., no FAIL
```

And one behavioural check that the split is real: adding `ERR_BACKUP_SNAPSHOT_NOT_FOUND = "x"` to `errors.en.toml` while it also exists in `errors.backup.en.toml` must make `errcodelint` report "defined in 2 files". Undo it afterwards.
