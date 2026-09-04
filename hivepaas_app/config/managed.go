package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	managedSettingsFileName = "hivepaas.toml"
	// managedSettingsMaxPerm is the widest permission the file may carry. It holds
	// the key that decrypts every stored secret, so anything group- or
	// world-readable is refused rather than quietly accepted.
	managedSettingsMaxPerm fs.FileMode = 0o600
)

var ErrManagedSettingsPermissive = errors.New("managed settings file is too permissive")

// ManagedSettings are the settings HivePaaS writes for itself.
//
// They cannot live in the database - the app secret is what decrypts the database
// rows in the first place - so they are kept in a file the app owns, inside its
// own volume, and applied on top of everything else. That ordering is what lets a
// value the app rotates at runtime win over a stale one still set in the
// environment or the base config file.
//
// This struct is also the allowlist. A setting not declared here cannot be
// overridden by the file at all, so write access to the app volume does not let
// anyone repoint the database or change the listening address.
type ManagedSettings struct {
	Secret string `toml:"secret"`
}

// applyTo overlays the settings that are set. A field the file omits leaves the
// loaded value alone, so a partial file never wipes configuration.
func (s *ManagedSettings) applyTo(config *Config) {
	if s == nil {
		return
	}
	if s.Secret != "" {
		config.Secret = s.Secret
	}
}

// ManagedSettingsPath returns where the managed settings file lives for an app path.
func ManagedSettingsPath(appPath string) string {
	return filepath.Join(appPath, managedSettingsFileName)
}

// loadManagedSettings reads the managed settings file, if there is one.
//
// A missing file is the normal case and not an error: the settings only exist
// once the app has written something for itself.
func loadManagedSettings(appPath string) (*ManagedSettings, error) {
	path := ManagedSettingsPath(appPath)

	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &ManagedSettings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat managed settings: %w", err)
	}
	if perm := info.Mode().Perm(); perm&^managedSettingsMaxPerm != 0 {
		return nil, fmt.Errorf("%w: %s has mode %#o, want %#o or stricter",
			ErrManagedSettingsPermissive, path, perm, managedSettingsMaxPerm)
	}

	settings := &ManagedSettings{}
	// Unknown keys are ignored on purpose: the struct is the allowlist, and a file
	// carrying something we do not manage must not break startup.
	if _, err := toml.DecodeFile(path, settings); err != nil {
		return nil, fmt.Errorf("failed to parse managed settings %s: %w", path, err)
	}
	return settings, nil
}

// saveManagedSettings writes the managed settings, replacing any existing file.
//
// The write is atomic: the file holds the key that decrypts every stored secret,
// so a crash halfway through must never leave a truncated one behind. The
// temporary file is created in the same directory so the rename stays within one
// filesystem, and both it and the directory are synced - without the directory
// sync the rename itself can be lost on power failure.
func saveManagedSettings(appPath string, settings *ManagedSettings) (err error) {
	var buf bytes.Buffer
	// G117 flags marshaling a field named like a secret. That is the whole point
	// of this file: the app secret has to be written somewhere the app owns, and
	// the write below restricts it to 0600.
	if err := toml.NewEncoder(&buf).Encode(settings); err != nil { //nolint:gosec
		return fmt.Errorf("failed to encode managed settings: %w", err)
	}

	tmp, err := os.CreateTemp(appPath, managedSettingsFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp managed settings: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if err = tmp.Chmod(managedSettingsMaxPerm); err != nil {
		return fmt.Errorf("failed to restrict managed settings permission: %w", err)
	}
	if _, err = tmp.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write managed settings: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync managed settings: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("failed to close managed settings: %w", err)
	}

	if err = os.Rename(tmpPath, ManagedSettingsPath(appPath)); err != nil {
		return fmt.Errorf("failed to install managed settings: %w", err)
	}
	return syncDir(appPath)
}

// syncDir flushes a directory entry, so a rename survives a power failure.
func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open app path: %w", err)
	}
	defer dir.Close()

	if err := dir.Sync(); err != nil {
		return fmt.Errorf("failed to sync app path: %w", err)
	}
	return nil
}
