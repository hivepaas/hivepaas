package gittool

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) checkoutTargetCommit(
	ctx context.Context,
	repo *git.Repository,
) (commit *object.Commit, err error) {
	commitHash := cli.opts.CommitHash
	if commitHash != "" { //nolint:nestif
		// Fetch the commit
		cmd := exec.CommandContext(ctx, "git", "fetch", "--depth=1", "origin", commitHash)
		cmd.Dir = cli.opts.CheckoutDir
		cmd.Env = cli.sharedEnv

		out, err := cmd.CombinedOutput()
		addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
		if err != nil {
			return nil, apperrors.New(err)
		}

		// Make sure the commit belongs to the branch (skip for Pull Requests)
		if !cli.opts.refType.IsPull() {
			cmd = exec.CommandContext(ctx, "git", "merge-base", "--is-ancestor", commitHash,
				fmt.Sprintf("%s/%s", cli.opts.RemoteName, cli.opts.refShort))
			cmd.Dir = cli.opts.CheckoutDir
			cmd.Env = []string{}
			out, err = cmd.CombinedOutput()
			addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
			if err != nil {
				return nil, apperrors.New(err)
			}
		}
	} else {
		//nolint:gosec
		cmd := exec.CommandContext(ctx, "git", "fetch", "--depth=1",
			cli.opts.RemoteName, cli.opts.refShort)
		cmd.Dir = cli.opts.CheckoutDir
		cmd.Env = cli.sharedEnv

		out, err := cmd.CombinedOutput()
		addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	// Hard reset the branch to make it point to the last fetched commit
	var cmd *exec.Cmd
	if cli.opts.refType.IsPull() {
		cmd = exec.CommandContext(ctx, "git", "checkout", "--detach", "FETCH_HEAD")
	} else {
		cmd = exec.CommandContext(ctx, "git", "checkout", "-B", cli.opts.refShort, "FETCH_HEAD") //nolint:gosec
	}
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = cli.sharedEnv

	out, err := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
	if err != nil {
		return nil, apperrors.New(err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, apperrors.New(err)
	}
	commit, err = repo.CommitObject(head.Hash())
	if err != nil {
		return nil, apperrors.New(err)
	}
	return commit, nil
}
