package logging

import "testing"

// The package functions must be safe before InitGlobalLogger runs: they used to
// dereference a nil global and crash the process while trying to log.
func TestPackageFuncsWorkBeforeInit(t *testing.T) {
	if globalLogger == nil {
		t.Fatal("globalLogger must never be nil")
	}
	Errorf("fallback %s", "works")
	Warnf("fallback warn")
	Info("fallback info")
	if GlobalLogger() == nil {
		t.Fatal("GlobalLogger() must never return nil")
	}
}
