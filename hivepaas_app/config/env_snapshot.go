package config

import (
	"os"
	"strings"
	"sync"
)

// envVarPrefix is the prefix of every environment variable this app reads. It
// matches the `env:"HP_..."` tags on the config fields.
const envVarPrefix = envPrefix + "_"

var (
	envSnapshotOnce sync.Once
	// envSnapshot holds the HP_* variables the process started with.
	envSnapshot map[string]string
)

// snapshotEnv records the HP_* variables once, the first time config is loaded.
//
// They are cleared from the process afterwards, so a later reload would otherwise
// see none of them and silently fall back to the config file or the defaults -
// a setting passed through docker compose would apply on boot and vanish on the
// next SIGHUP. Keeping our own copy makes the environment a value we read, not
// process state we depend on.
func snapshotEnv() {
	envSnapshotOnce.Do(func() {
		envSnapshot = map[string]string{}
		for _, entry := range os.Environ() {
			name, value, found := strings.Cut(entry, "=")
			if found && strings.HasPrefix(name, envVarPrefix) {
				envSnapshot[name] = value
			}
		}
	})
}

// restoreEnv puts the snapshot back so configor can read it.
func restoreEnv() {
	for name, value := range envSnapshot {
		_ = os.Setenv(name, value)
	}
}

// clearEnv removes every HP_* variable from the process.
//
// The app runs user-supplied commands and hands them os.Environ(); anything left
// here - the app secret above all - would be readable with a plain `env`. Note
// this does not scrub /proc/<pid>/environ, which is a snapshot taken at exec and
// is not affected by unsetenv.
func clearEnv() {
	for _, entry := range os.Environ() {
		name, _, found := strings.Cut(entry, "=")
		if found && strings.HasPrefix(name, envVarPrefix) {
			_ = os.Unsetenv(name)
		}
	}
}
