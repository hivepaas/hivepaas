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
		"ERR_BASE":                true,  // errors.New
		"ERR_DERIVED":             true,  // NewErr
		"ERR_DECLARED_NEVER_USED": true,  // NewErr, bound to a name nothing uses
		"ERR_REFERENCED_ONLY":     false, // plain string literal
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

func TestScanGoDirsRecordsTheDeclaringIdentifier(t *testing.T) {
	fset := token.NewFileSet()

	uses, err := scanGoDirs(fset, []string{"testdata/gosample"})
	if err != nil {
		t.Fatalf("scanGoDirs: %v", err)
	}

	idents := map[string]string{}
	for _, u := range uses {
		if u.Declared {
			idents[u.Code] = u.Ident
		}
	}

	assert.Equal(t, "ErrBase", idents["ERR_BASE"])
	assert.Equal(t, "ErrDerived", idents["ERR_DERIVED"])
	assert.Equal(t, "ErrDeclaredNeverUsed", idents["ERR_DECLARED_NEVER_USED"])
}

// A declaration binds its name once. Anything above that count is a reference,
// which is what separates a code in use from one left behind.
func TestScanIdentUsesCountsDeclarationAndReferences(t *testing.T) {
	fset := token.NewFileSet()
	names := map[string]bool{"ErrBase": true, "ErrDerived": true, "ErrDeclaredNeverUsed": true}

	counts, err := scanIdentUses(fset, []string{"testdata/gosample"}, names)
	if err != nil {
		t.Fatalf("scanIdentUses: %v", err)
	}

	assert.Equal(t, 1, counts["ErrDeclaredNeverUsed"], "declared and never referenced")
	assert.Equal(t, 3, counts["ErrBase"], "declared, then used as a base by two others")
	assert.Equal(t, 2, counts["ErrDerived"], "declared, then returned by alsoUse")
}
