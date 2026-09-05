package logging

import (
	"fmt"
	"os"
)

// fallbackLogger is the logger installed until InitGlobalLogger replaces it. It
// exists so a log line before the bootstrap is a line on stderr rather than a nil
// dereference.
var fallbackLogger Logger = &stderrLogger{}

type stderrLogger struct{}

func (l *stderrLogger) write(level, msg string, keysAndValues ...any) {
	if len(keysAndValues) > 0 {
		fmt.Fprintf(os.Stderr, "[%s] %s %v\n", level, msg, keysAndValues)
		return
	}
	fmt.Fprintf(os.Stderr, "[%s] %s\n", level, msg)
}

func (l *stderrLogger) writef(level, template string, args ...any) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", level, fmt.Sprintf(template, args...))
}

func (l *stderrLogger) Info(msg string, keysAndValues ...any) { l.write("INFO", msg, keysAndValues...) }
func (l *stderrLogger) Error(msg string, keysAndValues ...any) {
	l.write("ERROR", msg, keysAndValues...)
}
func (l *stderrLogger) Debug(msg string, keysAndValues ...any) {
	l.write("DEBUG", msg, keysAndValues...)
}
func (l *stderrLogger) Warn(msg string, keysAndValues ...any) { l.write("WARN", msg, keysAndValues...) }

func (l *stderrLogger) Infof(template string, args ...any)  { l.writef("INFO", template, args...) }
func (l *stderrLogger) Errorf(template string, args ...any) { l.writef("ERROR", template, args...) }
func (l *stderrLogger) Warnf(template string, args ...any)  { l.writef("WARN", template, args...) }
func (l *stderrLogger) Debugf(template string, args ...any) { l.writef("DEBUG", template, args...) }

// Fatal and Panic keep their meaning: a caller that reaches them is saying the process
// cannot continue, and swallowing that because no logger was installed would be worse
// than the crash it replaces.
func (l *stderrLogger) Fatal(keysAndValues ...any) {
	l.write("FATAL", fmt.Sprint(keysAndValues...))
	os.Exit(1)
}

func (l *stderrLogger) Fatalf(template string, args ...any) {
	l.writef("FATAL", template, args...)
	os.Exit(1)
}

func (l *stderrLogger) Panic(keysAndValues ...any) {
	panic(fmt.Sprint(keysAndValues...))
}

func (l *stderrLogger) Panicf(template string, args ...any) {
	panic(fmt.Sprintf(template, args...))
}
