package config

import (
	"sync/atomic"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// SystemInfo is runtime state, not configuration.
//
// It used to be a field on Config, which read naturally but was wrong in one
// specific way: it is written while the process runs - a request handler updates
// the installation step - and Config is read by every goroutine at once. A field
// written in place inside a shared struct is a data race however carefully the
// pointer to that struct is published, so this lives on its own.
type SystemInfo struct {
	NextStep base.InstallationStep
}

var systemInfo atomic.Pointer[SystemInfo]

func init() {
	systemInfo.Store(&SystemInfo{})
}

// CurrentSystemInfo returns a copy of the runtime state. A copy, so a caller
// cannot write through it into what everyone else is reading.
func CurrentSystemInfo() SystemInfo {
	return *systemInfo.Load()
}

// SetInstallationStep records the step the installation is waiting on, or clears
// it with an empty value once there is nothing left to do.
func SetInstallationStep(step base.InstallationStep) {
	systemInfo.Store(&SystemInfo{NextStep: step})
}
