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

// A code bound to a name nothing refers to is one the product cannot raise: the
// declaration is the only thing keeping it alive, and its translation ships for
// a message no caller can produce.
func TestCheckReportsDeclaredButNeverReferenced(t *testing.T) {
	uses := []codeUse{{
		Code: "ERR_ABANDONED", Ident: "ErrAbandoned", Declared: true,
		Pos: token.Position{Filename: "constants.go", Line: 9},
	}}
	entries := []messageEntry{{ID: "ERR_ABANDONED", File: "messages/en/errors.en.toml"}}

	out := msgs(checkAll(uses, entries, nil, map[string]int{"ErrAbandoned": 1}))

	assert.Contains(t, out, "ERR_ABANDONED")
	assert.Contains(t, out, "never referenced")
}

func TestCheckAcceptsDeclaredAndReferenced(t *testing.T) {
	uses := []codeUse{{Code: "ERR_LIVE", Ident: "ErrLive", Declared: true}}
	entries := []messageEntry{{ID: "ERR_LIVE", File: "messages/en/errors.en.toml"}}

	assert.Empty(t, checkAll(uses, entries, nil, map[string]int{"ErrLive": 4}))
}

// A code reached only as a bare string, such as the validation library's, binds
// no identifier and cannot be judged this way.
func TestCheckSkipsCodesThatBindNoIdentifier(t *testing.T) {
	uses := []codeUse{{Code: "ERR_VLD_SOMETHING", Declared: false}}
	entries := []messageEntry{{ID: "ERR_VLD_SOMETHING", File: "messages/en/errors.validation.en.toml"}}

	assert.Empty(t, checkAll(uses, entries, nil, map[string]int{}))
}
