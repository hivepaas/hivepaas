package gittool

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func TestValidateWithGitCli(t *testing.T) {
	// Create a temp directory for our mock git repository
	repoDir := t.TempDir()
	tempWorkDir := t.TempDir()

	// Initialize mock git repo
	runGit := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		err := cmd.Run()
		if err != nil {
			t.Fatalf("failed to run git command %v: %v", args, err)
		}
	}

	runGit(repoDir, "init", "-b", "main")
	runGit(repoDir, "config", "user.name", "Test User")
	runGit(repoDir, "config", "user.email", "test@example.com")

	// Commit 1 on branch main
	err := os.WriteFile(filepath.Join(repoDir, "file1.txt"), []byte("commit 1"), 0644)
	if err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	runGit(repoDir, "add", "file1.txt")
	runGit(repoDir, "commit", "-m", "first commit")

	// Commit 2 on branch main
	err = os.WriteFile(filepath.Join(repoDir, "file2.txt"), []byte("commit 2"), 0644)
	if err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	runGit(repoDir, "add", "file2.txt")
	runGit(repoDir, "commit", "-m", "second commit")

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get HEAD commit: %v", err)
	}
	commitHash2 := strings.TrimSpace(string(out))

	// Create branch "feature"
	runGit(repoDir, "checkout", "-b", "feature")

	// Commit 3 on branch feature
	err = os.WriteFile(filepath.Join(repoDir, "file3.txt"), []byte("commit 3"), 0644)
	if err != nil {
		t.Fatalf("failed to write file3: %v", err)
	}
	runGit(repoDir, "add", "file3.txt")
	runGit(repoDir, "commit", "-m", "third commit")

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	out, err = cmd.Output()
	if err != nil {
		t.Fatalf("failed to get feature HEAD commit: %v", err)
	}
	commitHash3 := strings.TrimSpace(string(out))

	// Create a tag "v1.0.0" on main branch
	runGit(repoDir, "tag", "v1.0.0")

	// Checkout main again
	runGit(repoDir, "checkout", "main")

	ctx := context.Background()

	// 1. Success case: URL, main branch, no commit
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/heads/main",
		TempDir:       tempWorkDir,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 2. Success case: URL, main branch, commitHash2 (which belongs to main)
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/heads/main",
		CommitHash:    commitHash2,
		TempDir:       tempWorkDir,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 3. ErrRepoNotFound: invalid repository URL
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           filepath.Join(repoDir, "invalid-repo-path"),
		ReferenceName: "refs/heads/main",
		TempDir:       tempWorkDir,
	})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 4. ErrRepoRefNotFound: invalid reference
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/heads/invalid-branch",
		TempDir:       tempWorkDir,
	})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 5. ErrRepoCommitNotFound: commit does not exist
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/heads/main",
		CommitHash:    "0123456789abcdef0123456789abcdef01234567",
		TempDir:       tempWorkDir,
	})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 6. ErrRepoCommitNotBelongToRef: commitHash3 belongs to feature branch, not main branch
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/heads/main",
		CommitHash:    commitHash3,
		TempDir:       tempWorkDir,
	})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 7. Success case for tags: URL, tag ref, no commit
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "refs/tags/v1.0.0",
		TempDir:       tempWorkDir,
	})
	if err != nil {
		t.Errorf("expected no error for tag validation, got %v", err)
	}

	// 8. Error case for invalid reference format (e.g. HEAD, not starts with refs/heads/ or refs/tags/)
	err = ValidateWithGitCli(ctx, &ValidationOptions{
		URL:           repoDir,
		ReferenceName: "HEAD",
		TempDir:       tempWorkDir,
	})
	if !errors.Is(err, apperrors.ErrUnsupported) {
		t.Errorf("expected ErrUnsupported for HEAD, got %v", err)
	}
}
