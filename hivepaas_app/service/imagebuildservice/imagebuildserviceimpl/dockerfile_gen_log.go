package imagebuildserviceimpl

import (
	"context"
	"log/slog"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

// genLogHandler writes what the Dockerfile generator detects - the runtime, its
// version, the install, build and start commands - into the deployment's log,
// where its user reads it, rather than the agent's own output.
type genLogHandler struct {
	ctx      context.Context //nolint:containedctx // the build's, which the generator does not pass on
	logStore *tasklog.Store
}

func (h *genLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

func (h *genLogHandler) Handle(_ context.Context, record slog.Record) error {
	frame := tasklog.NewOutFrame(record.Message, tasklog.TsNow)
	switch {
	case record.Level >= slog.LevelError:
		frame = tasklog.NewErrFrame(record.Message, tasklog.TsNow)
	case record.Level >= slog.LevelWarn:
		frame = tasklog.NewWarnFrame(record.Message, tasklog.TsNow)
	}
	// The build's context: slog passes the generator's, a background one
	if err := h.logStore.AddRedacted(h.ctx, frame); err != nil { //nolint:contextcheck
		return hperrors.Wrap(err)
	}
	return nil
}

func (h *genLogHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *genLogHandler) WithGroup(string) slog.Handler { return h }
