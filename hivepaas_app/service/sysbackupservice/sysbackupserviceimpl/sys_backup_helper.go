package sysbackupserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

func logCmdOutput(
	ctx context.Context,
	msg string,
	isErr bool,
	logStore *tasklog.Store,
) {
	if logStore == nil || len(msg) == 0 {
		return
	}
	fn := gofn.If(isErr, tasklog.NewErrFrame, tasklog.NewOutFrame)
	_ = logStore.Add(ctx, fn(msg, tasklog.TsNow))
}
