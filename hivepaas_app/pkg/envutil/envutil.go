// Package envutil helps hand a safe environment to child processes.
package envutil

import (
	"os"
	"strings"
)

// appEnvPrefix is the prefix of every environment variable this app reads for
// itself, including the secret that decrypts the database.
const appEnvPrefix = "HP_"

// SafeEnviron returns the process environment without this app's own variables.
//
// Commands run on behalf of a user must not be able to read the app's secrets
// with a plain `env`. Config loading already clears these from the process, but
// filtering here too means one missed step is not enough to leak them.
func SafeEnviron() []string {
	environ := os.Environ()
	safe := make([]string, 0, len(environ))
	for _, entry := range environ {
		if !strings.HasPrefix(entry, appEnvPrefix) {
			safe = append(safe, entry)
		}
	}
	return safe
}
