package gittool

import (
	"context"
	"os"
	"os/exec"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) clone(
	ctx context.Context,
) (err error) {
	err = os.MkdirAll(cli.opts.CheckoutDir, base.DirModeDefault)
	if err != nil {
		return apperrors.New(err)
	}

	var cmd *exec.Cmd
	if cli.opts.refType.IsPull() {
		// For Pull Requests / Merge Requests, we clone the default branch (without specifying --branch)
		// because cloning custom refs directly with --branch is not supported on all git servers/clients.
		//nolint:gosec
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth=1", "--",
			cli.opts.URL, cli.opts.CheckoutDir)
	} else {
		//nolint:gosec
		cmd = exec.CommandContext(ctx, "git", "clone", "--single-branch", "--depth=1",
			"--branch="+cli.opts.refShort, "--", cli.opts.URL, cli.opts.CheckoutDir)
	}
	cmd.Env = cli.sharedEnv
	out, err := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
