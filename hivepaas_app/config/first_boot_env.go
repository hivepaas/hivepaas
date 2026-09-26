package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// firstBootEnvFileName is the settings the installer hands the first boot
	// without putting them in a service's environment, where `docker service
	// inspect` would show them for as long as the service exists.
	firstBootEnvFileName = "first-boot.env"
	// firstBootEnvMaxPerm is the widest permission the file may carry: it holds
	// the admin's password until the first boot has used it.
	firstBootEnvMaxPerm fs.FileMode = 0o600
)

var (
	ErrFirstBootEnvPermissive = errors.New("first-boot env file is too permissive")
	ErrFirstBootEnvInvalid    = errors.New("first-boot env file is invalid")
)

// FirstBootEnvPath returns where the first-boot settings live for an app path.
func FirstBootEnvPath(appPath string) string {
	return filepath.Join(appPath, firstBootEnvFileName)
}

// readFirstBootEnv reads the first-boot settings: HP_* variables, one KEY=VALUE
// a line, the value as it stands after the first '=' - no quotes, no escapes,
// so a password may hold any character but a line break. Comments and blank
// lines are skipped. A missing file is the normal case: it exists only until
// the first boot is done.
//
// It is for values the first boot uses once, such as the admin's account.
// Whatever every boot needs belongs in the managed settings (hivepaas.toml):
// this file is deleted once the installation's data exists.
func readFirstBootEnv(path string) (map[string]string, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil //nolint:nilnil // no file is the normal case
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat first-boot env: %w", err)
	}
	if perm := info.Mode().Perm(); perm&^firstBootEnvMaxPerm != 0 {
		return nil, fmt.Errorf("%w: %s has mode %#o, want %#o or stricter",
			ErrFirstBootEnvPermissive, path, perm, firstBootEnvMaxPerm)
	}
	content, err := os.ReadFile(path) // #nosec G304 - the app's own directory
	if err != nil {
		return nil, fmt.Errorf("failed to read first-boot env: %w", err)
	}

	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("%w: %s line %d is not KEY=VALUE", ErrFirstBootEnvInvalid, path, n)
		}
		if !strings.HasPrefix(key, envVarPrefix) {
			return nil, fmt.Errorf("%w: %s line %d: %s is not a setting of this app (%s*)",
				ErrFirstBootEnvInvalid, path, n, key, envVarPrefix)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read first-boot env: %w", err)
	}
	return values, nil
}

// applyFirstBootEnv puts the first-boot settings into the process environment
// for the config to read, under whatever the service's own environment already
// says: a value set there is set on purpose.
func applyFirstBootEnv(appPath string) error {
	values, err := readFirstBootEnv(FirstBootEnvPath(appPath))
	if err != nil {
		return err
	}
	for key, value := range values {
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to apply first-boot env %s: %w", key, err)
		}
	}
	return nil
}

// RemoveFirstBootEnv deletes the first-boot settings once the first boot has
// used them. Nothing to delete is not an error.
func RemoveFirstBootEnv() error {
	err := os.Remove(FirstBootEnvPath(resolveAppPath()))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to remove first-boot env: %w", err)
	}
	return nil
}
