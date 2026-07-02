package gittool

import (
	"context"
	"os/exec"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

//nolint:unused
func (cli *checkoutCli) gitCliPullLfs(
	ctx context.Context,
) (err error) {
	if !cli.opts.LFSEnabled {
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "lfs", "pull")
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = cli.sharedEnv

	out, err := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
