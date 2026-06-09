package githelper

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

func addLog(
	ctx context.Context,
	msg string,
	isErr bool,
	checkoutOpts *CheckoutOptions,
) {
	if checkoutOpts.LogStore == nil || len(msg) == 0 {
		return
	}
	fn := gofn.If(isErr, tasklog.NewErrFrame, tasklog.NewDebugFrame)
	_ = checkoutOpts.LogStore.Add(ctx, fn(msg, tasklog.TsNow))
}
