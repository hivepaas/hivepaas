// Package safego provides the helpers for handling panics.
//
// A panic in a bare `go func()` cannot be recovered by whoever started it: gin's
// recovery middleware and RecoverTo only protect the goroutine they run in, and
// sync.WaitGroup.Go does not recover either. Any goroutine that touches external
// input (archives, gRPC/websocket streams, docker events) or runs for the
// lifetime of the process should be started through this package.
//
// The helpers split by who is responsible for the panic:
//
//   - Recover / RecoverWithLogger / RecoverPipe contain a panic where there is
//     nobody to return it to - a goroutine. They log it, they never return it.
//   - RecoverTo turns a panic into an error at a function boundary, so the
//     caller decides what to do with it. It does not log: the caller does.
package safego

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// Go runs fn in a new goroutine, recovering and logging any panic it raises.
// `name` identifies the goroutine in the panic log.
func Go(name string, fn func()) {
	go func() {
		defer Recover(name)
		fn()
	}()
}

// GoWithLogger is Go using an explicit logger instead of the global one.
func GoWithLogger(logger logging.Logger, name string, fn func()) {
	go func() {
		defer RecoverWithLogger(logger, name)
		fn()
	}()
}

// Recover recovers from a panic in the current goroutine and logs it with the
// stack trace. Use it as the first statement of a goroutine body that cannot
// be started via Go, for example inside sync.WaitGroup.Go:
//
//	wg.Go(func() {
//		defer safego.Recover("delete-project-env")
//		...
//	})
func Recover(name string) {
	if r := recover(); r != nil {
		logPanic(logging.GlobalLogger(), name, r, debug.Stack())
	}
}

// RecoverWithLogger is Recover using an explicit logger instead of the global one.
func RecoverWithLogger(logger logging.Logger, name string) {
	if r := recover(); r != nil {
		logPanic(logger, name, r, debug.Stack())
	}
}

// RecoverTo recovers a panic in the current goroutine and stores it in
// *currentErr, joining it with whatever error is already there. Use it at a
// function boundary that returns an error, so the caller decides what happens:
//
//	func (s *service) Backup(ctx context.Context) (err error) {
//		defer safego.RecoverTo(&err)
//		...
//	}
//
// Unlike Recover it does not log - the panic becomes a normal error and the
// caller is expected to handle it.
//
// currentErr must not be nil. Passing nil leaves nowhere to return the panic to,
// so RecoverTo logs it rather than swallowing it silently; use Recover when you
// deliberately want a panic contained and logged. The goroutinelint check
// rejects a literal nil argument.
func RecoverTo(currentErr *error) {
	r := recover()
	if r == nil {
		return
	}
	if currentErr == nil {
		logPanic(logging.GlobalLogger(), "RecoverTo(nil)", r, debug.Stack())
		return
	}

	panicErr := hperrors.NewPanic(r)
	if *currentErr != nil {
		*currentErr = errors.Join(*currentErr, panicErr)
		return
	}
	*currentErr = panicErr
}

// RecoverPipe is Recover for a goroutine that feeds an io.Pipe. On panic it also
// fails the pipe writer, so the reader on the other side gets an error instead
// of blocking forever on a pipe nobody will ever close.
//
// Use it as the first statement of the goroutine body:
//
//	go func() {
//		defer safego.RecoverPipe("containerfile.tarStream", pw)
//		...
//	}()
func RecoverPipe(name string, pw *io.PipeWriter) {
	if r := recover(); r != nil {
		logPanic(logging.GlobalLogger(), name, r, debug.Stack())
		if pw != nil {
			_ = pw.CloseWithError(hperrors.NewPanic(r))
		}
	}
}

// LogPanic logs an already-recovered panic value together with the current
// stack. Use it when the goroutine needs to do its own cleanup with the panic
// value and cannot simply hand everything to Recover.
func LogPanic(name string, r any) {
	logPanic(logging.GlobalLogger(), name, r, debug.Stack())
}

func logPanic(logger logging.Logger, name string, r any, stack []byte) {
	if logger == nil {
		// The global logger is not initialized yet (bootstrap, tests, tools).
		// Never let the reporting itself panic.
		fmt.Fprintf(os.Stderr, "[panic] goroutine %s: %v\n%s\n", name, r, stack)
		return
	}
	logger.Errorf("[panic] goroutine %s: %v\n%s", name, r, stack)
}
