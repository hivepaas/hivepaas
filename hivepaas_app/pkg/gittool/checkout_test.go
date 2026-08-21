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

func TestCheckoutWithGitCli(t *testing.T) {
	// Create mock git repo
	repoDir := t.TempDir()
	tempWorkDir := t.TempDir()

	runGit := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test User",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test User",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_CONFIG_GLOBAL=/dev/null",
		)
		err := cmd.Run()
		if err != nil {
			t.Fatalf("failed git command %v: %v", args, err)
		}
	}

	runGit(repoDir, "init", "-b", "main")

	// Commit 1 on branch main
	err := os.WriteFile(filepath.Join(repoDir, "file1.txt"), []byte("commit 1"), 0644)
	assert.NoError(t, err)
	runGit(repoDir, "add", "file1.txt")
	runGit(repoDir, "commit", "-m", "first commit")

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	assert.NoError(t, err)
	commitHash1 := strings.TrimSpace(string(out))

	// Commit 2 on branch main
	err = os.WriteFile(filepath.Join(repoDir, "file2.txt"), []byte("commit 2"), 0644)
	assert.NoError(t, err)
	runGit(repoDir, "add", "file2.txt")
	runGit(repoDir, "commit", "-m", "second commit")

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	out, err = cmd.Output()
	assert.NoError(t, err)
	commitHash2 := strings.TrimSpace(string(out))

	ctx := context.Background()

	t.Run("Checkout latest branch HEAD without commitHash", func(t *testing.T) {
		checkoutDir := t.TempDir()
		commit, err := CheckoutWithGitCli(ctx, &CheckoutOptions{
			URL:           repoDir,
			ReferenceName: "refs/heads/main",
			TempDir:       tempWorkDir,
			CheckoutDir:   checkoutDir,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, commit) {
			assert.Equal(t, commitHash2, commit.Hash)
			assert.Equal(t, "second commit", commit.Message)
			assert.Equal(t, "Test User", commit.Author)
			assert.FileExists(t, filepath.Join(checkoutDir, "file2.txt"))
		}
	})

	t.Run("Checkout specific older commit on branch", func(t *testing.T) {
		checkoutDir := t.TempDir()
		commit, err := CheckoutWithGitCli(ctx, &CheckoutOptions{
			URL:           repoDir,
			ReferenceName: "refs/heads/main",
			CommitHash:    commitHash1,
			TempDir:       tempWorkDir,
			CheckoutDir:   checkoutDir,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, commit) {
			assert.Equal(t, commitHash1, commit.Hash)
			assert.Equal(t, "first commit", commit.Message)
			assert.Equal(t, "Test User", commit.Author)
			assert.FileExists(t, filepath.Join(checkoutDir, "file1.txt"))
			assert.NoFileExists(t, filepath.Join(checkoutDir, "file2.txt"))
		}
	})
}
