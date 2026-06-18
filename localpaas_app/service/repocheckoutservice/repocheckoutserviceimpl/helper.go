package repocheckoutserviceimpl

import (
	"context"
	"os"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

func (s *service) resetCheckoutDir(
	data *repoCheckoutData,
) error {
	if err := os.RemoveAll(data.CheckoutDir); err != nil {
		return apperrors.Wrap(err)
	}
	if err := os.MkdirAll(data.CheckoutDir, base.DirModeDefault); err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) addCmdOutToLogs(
	ctx context.Context,
	msg string,
	isErr bool,
	logStore *tasklog.Store,
) {
	if logStore == nil || len(msg) == 0 {
		return
	}
	fn := gofn.If(isErr, tasklog.NewErrFrame, tasklog.NewDebugFrame)
	_ = logStore.Add(ctx, fn(msg, tasklog.TsNow))
}
