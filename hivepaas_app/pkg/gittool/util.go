package gittool

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

func addLog(
	ctx context.Context,
	msg string,
	isErr bool,
	log *tasklog.Store,
) {
	if log == nil || len(msg) == 0 {
		return
	}
	fn := gofn.If(isErr, tasklog.NewErrFrame, tasklog.NewDebugFrame)
	_ = log.Add(ctx, fn(msg, tasklog.TsNow))
}
