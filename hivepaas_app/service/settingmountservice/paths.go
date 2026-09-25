package settingmountservice

import (
	"path"
)

const (
	secretsDir = "/run/secrets"
	configsDir = "/"
)

// ValidPath reports whether a file may be mounted at p: absolute, clean, not the
// root.
func ValidPath(p string) bool {
	return p != "" && p != "/" && path.IsAbs(p) && path.Clean(p) == p
}

// SecretTarget is where a secret reference's file lands: a relative name is
// under /run/secrets.
func SecretTarget(name string) string {
	if path.IsAbs(name) {
		return path.Clean(name)
	}
	return path.Join(secretsDir, name)
}

// ConfigTarget is where a config reference's file lands: a relative name is
// under the root.
func ConfigTarget(name string) string {
	if path.IsAbs(name) {
		return path.Clean(name)
	}
	return path.Join(configsDir, name)
}
