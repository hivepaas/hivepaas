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
	wantLines := []int{13, 20, 26, 34, 39, 47, 48, 105, 120, 126}
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

	// A guard called as a plain statement does nothing: recover() only returns a
	// panic to a deferred function.
	for _, line := range []int{120, 126} {
		if msg := got[line]; !strings.Contains(msg, "without defer") {
			t.Errorf("line %d should report a guard called without defer, got: %s", line, msg)
		}
	}
	if msg := got[126]; !strings.Contains(msg, "safego.RecoverTo") {
		t.Errorf("line 126 should name the helper that was misused, got: %s", msg)
	}
	// The nested-guard trap is reported twice on purpose: once for the goroutine
	// having no working guard, once for the call that was supposed to be one.
	if msg := got[48]; !strings.Contains(msg, "without defer") {
		t.Errorf("line 48 should explain why the nested guard never fires, got: %s", msg)
	}
}

// A method that merely shares a name with a guard belongs to someone else and must
// not be reported, or the check would fire on unrelated APIs.
func TestUnrelatedRecoverMethodIsNotReported(t *testing.T) {
	findings, err := checkDir(token.NewFileSet(), "testdata/sample")
	if err != nil {
		t.Fatalf("checkDir: %v", err)
	}
	for _, f := range findings {
		if f.pos.Line == 133 {
			t.Errorf("an unrelated Recover method was reported: %s", f.msg)
		}
	}
}
