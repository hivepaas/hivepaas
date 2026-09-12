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
