package gittool

import (
	"context"
	"errors"
	"os/exec"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (cli *checkoutCli) cleanup(
	ctx context.Context,
) (err error) {
	// git reflog expire --expire=now --all
	cmd := exec.CommandContext(ctx, "git", "reflog", "expire", "--expire=now", "--all")
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = []string{}

	out, err1 := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err1 != nil, cli.opts.LogStore)
	err = errors.Join(err, err1)

	// git gc --prune=now
	cmd = exec.CommandContext(ctx, "git", "gc", "--prune=now")
	cmd.Dir = cli.opts.CheckoutDir
	cmd.Env = []string{}

	out, err2 := cmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err2 != nil, cli.opts.LogStore)
	err = errors.Join(err, err2)

	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
