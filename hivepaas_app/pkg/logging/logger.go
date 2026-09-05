package logging

import (
	"sync"
)

const LoggerCtxKey string = "logger"

var (
	// globalLogger starts as the fallback so the package functions are always safe to
	// call. Before InitGlobalLogger runs they are reached anyway - error rendering
	// logs a missing translation, tools and tests call in without a bootstrap - and a
	// nil here would turn each of those log lines into a crash.
	globalLogger Logger = fallbackLogger
	once         sync.Once
)

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Infof(template string, args ...any)
	Errorf(template string, args ...any)
	Warnf(template string, args ...any)
	Debugf(template string, args ...any)
	Fatal(keysAndValues ...any)
	Panic(keysAndValues ...any)
	Fatalf(template string, args ...any)
	Panicf(template string, args ...any)
}

// InitGlobalLogger sets singleton instance of Logger
func InitGlobalLogger(log Logger) {
	once.Do(func() {
		if log != nil {
			globalLogger = log
		}
	})
}

// GlobalLogger returns the singleton instance of Logger. It is never nil: before
// InitGlobalLogger has run it is a fallback that writes to stderr, so a caller
// outside the normal app bootstrap (tests, tools) still gets its output somewhere.
func GlobalLogger() Logger {
	return globalLogger
}

func Info(msg string, keysAndValues ...any) {
	globalLogger.Info(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...any) {
	globalLogger.Error(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...any) {
	globalLogger.Debug(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...any) {
	globalLogger.Warn(msg, keysAndValues...)
}

func Infof(template string, args ...any) {
	globalLogger.Infof(template, args...)
}

func Errorf(template string, args ...any) {
	globalLogger.Errorf(template, args...)
}

func Warnf(template string, args ...any) {
	globalLogger.Warnf(template, args...)
}

func Debugf(template string, args ...any) {
	globalLogger.Debugf(template, args...)
}

func Fatal(keysAndValues ...any) {
	globalLogger.Fatal(keysAndValues...)
}

func Panic(keysAndValues ...any) {
	globalLogger.Panic(keysAndValues...)
}

func Fatalf(template string, args ...any) {
	globalLogger.Fatalf(template, args...)
}

func Panicf(template string, args ...any) {
	globalLogger.Panicf(template, args...)
}
