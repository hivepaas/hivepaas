package gittool

import (
	"errors"
	"os/exec"
)

func IsAncestor(sourceDir string, branch, commitHash string) int {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", commitHash, branch)
	cmd.Dir = sourceDir
	cmd.Env = []string{}
	err := cmd.Run()
	if err == nil {
		return 0 // commit belongs to branch
	}
	// Check exit code to distinguish "not ancestor" (exit code 1) from other errors
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}
	return -1
}
