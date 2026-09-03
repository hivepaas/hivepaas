package main

import (
	"go/token"
	"strings"
	"testing"
)

func TestCheckDir(t *testing.T) {
	findings, err := checkDir(token.NewFileSet(), "testdata/sample")
	if err != nil {
		t.Fatalf("checkDir: %v", err)
	}

	got := map[int]string{}
	for _, f := range findings {
		got[f.pos.Line] = f.msg
	}

	// The fixture flags every bad case with a `// want:` comment; the ok cases
	// and the //safego:allow case must stay silent.
	wantLines := []int{13, 20, 26, 34, 39, 47, 105}
	for _, line := range wantLines {
		if _, ok := got[line]; !ok {
			t.Errorf("expected a finding at sample.go:%d, got none", line)
		}
	}
	if len(findings) != len(wantLines) {
		for _, f := range findings {
			t.Logf("finding at line %d: %s", f.pos.Line, f.msg)
		}
		t.Fatalf("expected %d findings, got %d", len(wantLines), len(findings))
	}

	if msg := got[34]; !strings.Contains(msg, "another package") {
		t.Errorf("line 34 should report an unverifiable cross-package callee, got: %s", msg)
	}
	if msg := got[26]; !strings.Contains(msg, "unguardedEntryPoint") {
		t.Errorf("line 26 should name the entry point, got: %s", msg)
	}
	if msg := got[105]; !strings.Contains(msg, "RecoverTo(nil)") {
		t.Errorf("line 105 should reject RecoverTo(nil), got: %s", msg)
	}
}
