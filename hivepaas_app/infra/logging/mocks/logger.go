package mocks

import (
	"fmt"
)

type Logger struct {
	Errors []string
}

func (m *Logger) Info(msg string, keysAndValues ...any)  {}
func (m *Logger) Error(msg string, keysAndValues ...any) { m.Errors = append(m.Errors, msg) }
func (m *Logger) Debug(msg string, keysAndValues ...any) {}
func (m *Logger) Warn(msg string, keysAndValues ...any)  {}
func (m *Logger) Infof(template string, args ...any)     {}
func (m *Logger) Errorf(template string, args ...any) {
	m.Errors = append(m.Errors, fmt.Sprintf(template, args...))
}
func (m *Logger) Warnf(template string, args ...any)  {}
func (m *Logger) Debugf(template string, args ...any) {}
func (m *Logger) Fatal(keysAndValues ...any)          {}
func (m *Logger) Panic(keysAndValues ...any)          {}
func (m *Logger) Fatalf(template string, args ...any) {}
func (m *Logger) Panicf(template string, args ...any) {}
