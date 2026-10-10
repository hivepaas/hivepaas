package gittool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// gitCalls puts a git ahead of the real one on PATH, which writes down the
// arguments it is called with, runs the real one, and answers `git lfs` with
// nothing: what a checkout asks git for, without git-lfs installed.
func gitCalls(t *testing.T) func() []string {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	logFile := filepath.Join(dir, "calls")
	script := "#!/bin/sh\necho \"$*\" >> " + logFile + "\n" +
		"if [ \"$1\" = lfs ]; then exit 0; fi\nexec " + realGit + " \"$@\"\n"
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755)) //nolint:gosec
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return func() []string {
		out, _ := os.ReadFile(logFile)
		return strings.Split(strings.TrimSpace(string(out)), "\n")
	}
}

// A checkout with LFS asked for pulls the LFS files of the commit checked out:
// one from a cached repository finds the pointers an earlier checkout without
// LFS left, and a checkout of the same commit does not replace them.
func TestCheckoutWithGitCli_PullsLFSFilesWhenAsked(t *testing.T) {
	repoDir := t.TempDir()
	run := exec.Command("sh", "-c", "git init -q -b main && echo one > f.txt && git add f.txt && "+
		"git -c user.name=t -c user.email=t@t commit -q -m one")
	run.Dir = repoDir
	run.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("making the repository: %v %s", err, out)
	}
	calls := gitCalls(t)
	ctx := context.Background()

	for _, cached := range []bool{false, true} {
		checkoutDir := filepath.Join(t.TempDir(), "checkout")
		if cached {
			_, err := CheckoutWithGitCli(ctx, &CheckoutOptions{URL: repoDir, ReferenceName: "refs/heads/main",
				TempDir: t.TempDir(), CheckoutDir: checkoutDir})
			assert.NoError(t, err)
		}
		_, err := CheckoutWithGitCli(ctx, &CheckoutOptions{URL: repoDir, ReferenceName: "refs/heads/main",
			LFSEnabled: true, CacheLoaded: cached, TempDir: t.TempDir(), CheckoutDir: checkoutDir})
		assert.NoError(t, err)
	}
	pulls := 0
	for _, call := range calls() {
		if call == "lfs pull" {
			pulls++
		}
	}
	assert.Equal(t, 2, pulls, "an LFS pull after each checkout asking for it, and none after the one that did not")
}
