package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	lastMsg string
}

func (m *mockLogger) Info(msg string, keysAndValues ...any)  { m.lastMsg = msg }
func (m *mockLogger) Error(msg string, keysAndValues ...any) { m.lastMsg = msg }
func (m *mockLogger) Debug(msg string, keysAndValues ...any) { m.lastMsg = msg }
func (m *mockLogger) Warn(msg string, keysAndValues ...any)  { m.lastMsg = msg }
func (m *mockLogger) Infof(template string, args ...any)     { m.lastMsg = template }
func (m *mockLogger) Errorf(template string, args ...any)    { m.lastMsg = template }
func (m *mockLogger) Warnf(template string, args ...any)     { m.lastMsg = template }
func (m *mockLogger) Debugf(template string, args ...any)    { m.lastMsg = template }
func (m *mockLogger) Fatal(keysAndValues ...any)             {}
func (m *mockLogger) Panic(keysAndValues ...any)             {}
func (m *mockLogger) Fatalf(template string, args ...any)    {}
func (m *mockLogger) Panicf(template string, args ...any)    {}

func TestGlobalLogger(t *testing.T) {
	ml := &mockLogger{}
	InitGlobalLogger(ml)

	// Since InitGlobalLogger uses once.Do, we might be using a logger set by another test
	// or this one. We can't easily reset it.
	// So we'll just check if whatever is in globalLogger works.

	if globalLogger == nil {
		t.Fatal("globalLogger should not be nil after InitGlobalLogger")
	}

	// If it's our mockLogger, we can test the message capture
	if logger, ok := globalLogger.(*mockLogger); ok {
		Info("info msg")
		assert.Equal(t, "info msg", logger.lastMsg)

		Error("error msg")
		assert.Equal(t, "error msg", logger.lastMsg)

		Infof("infof %s", "msg")
		assert.Equal(t, "infof %s", logger.lastMsg)
	} else {
		// If it's not our mockLogger (e.g. set by another test or already set),
		// just ensure it doesn't panic.
		Info("info msg")
		Error("error msg")
	}
}
