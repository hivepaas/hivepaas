package gittool

import (
	"context"
	"os"
	"os/exec"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) clone(
	ctx context.Context,
) (err error) {
	err = os.MkdirAll(cli.opts.CheckoutDir, base.DirModeDefault)
	if err != nil {
		return apperrors.New(err)
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "git", "clone", "--single-branch", "--depth=1",
		"--branch="+cli.opts.refShort, "--", cli.opts.URL, cli.opts.CheckoutDir)
	cmd.Env = cli.sharedEnv
	out, err := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, cli.opts.LogStore)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
