package gittool

import (
	"context"
	"os/exec"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) fetchSubmodules(
	ctx context.Context,
) (err error) {
	if !cli.opts.SubmodulesEnabled {
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "submodule", "update", "--init", "--recursive", "--depth=1")
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = cli.sharedEnv

	out, err := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
