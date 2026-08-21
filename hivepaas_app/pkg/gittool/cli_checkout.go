package gittool

import (
	"context"
	"os/exec"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) checkoutTargetCommit(
	ctx context.Context,
) (commit *CommitInfo, err error) {
	commitHash := cli.opts.CommitHash
	if commitHash != "" {
		// Fetch the commit
		cmd := exec.CommandContext(ctx, "git", "fetch", "--depth=1", "origin", commitHash)
		cmd.Dir = cli.opts.CheckoutDir
		cmd.Env = cli.sharedEnv

		out, err := cmd.CombinedOutput()
		addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
		if err != nil {
			return nil, apperrors.Wrap(err)
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
			return nil, apperrors.Wrap(err)
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
		return nil, apperrors.Wrap(err)
	}

	commit, err = cli.getHeadCommit(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return commit, nil
}

func (cli *checkoutCli) getHeadCommit(
	ctx context.Context,
) (*CommitInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", "--format=%H%x1f%an%x1f%B")
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = cli.sharedEnv

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	outStr := string(out)
	const noParts = 3
	parts := strings.SplitN(outStr, "\x1f", noParts)
	if len(parts) < noParts {
		return nil, apperrors.Wrap(apperrors.ErrGitLogOutputUnexpected).WithParam("Output", outStr)
	}

	return &CommitInfo{
		Hash:    strings.TrimSpace(parts[0]),
		Author:  strings.TrimSpace(parts[1]),
		Message: strings.TrimRight(parts[2], "\r\n"),
	}, nil
}
