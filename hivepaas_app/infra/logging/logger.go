package logging

import (
	"sync"
)

const LoggerCtxKey string = "logger"

var (
	globalLogger Logger
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
		globalLogger = log
	})
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
