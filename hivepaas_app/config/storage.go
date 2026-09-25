package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// ErrStorageInvalid is a storage directory Docker cannot bind, or one that would
// make the whole filesystem project data.
var ErrStorageInvalid = errors.New("storage directory is invalid")

const (
	// defaultProjectDataDir is where project data goes under HostDir when no
	// directory of its own is configured.
	defaultProjectDataDir = "project_data"
	// ProjectDataSettings names the settings a missing project data directory is
	// configured with, for the error that says so.
	ProjectDataSettings = "HP_STORAGE_PROJECT_DATA_HOST_DIR or HP_STORAGE_HOST_DIR"
)

type Storage struct {
	// HostDir is HivePaaS's data directory as the host sees it: the directory the
	// install creates, which the app's container sees at HP_APP_PATH. Docker is
	// given host paths, so directories it binds are named from here.
	HostDir string `toml:"host_dir" env:"HP_STORAGE_HOST_DIR"`
	// ProjectDataHostDir puts project data in a directory of the operator's own,
	// on the host, instead of project_data under HostDir. It must exist. Projects
	// already created keep their data where it was: each volume records its
	// directory.
	ProjectDataHostDir string `toml:"project_data_host_dir" env:"HP_STORAGE_PROJECT_DATA_HOST_DIR"`
}

// ProjectDataDirs is where project data goes on the host, as the directory that
// must already exist and the path under it that HivePaaS creates. A directory of
// the operator's own is used as it is; the default, project_data under
// HostDir, is created on first use, as the directories inside it are. Both
// are empty when neither is configured.
func (storage Storage) ProjectDataDirs() (base, prefix string) {
	if storage.ProjectDataHostDir != "" {
		return filepath.Clean(storage.ProjectDataHostDir), ""
	}
	if storage.HostDir != "" {
		return storage.HostDir, defaultProjectDataDir
	}
	return "", ""
}

// ProjectDataRoot is the directory project data lives in, or "" when none is
// configured.
func (storage Storage) ProjectDataRoot() string {
	base, prefix := storage.ProjectDataDirs()
	if base == "" {
		return ""
	}
	return filepath.Join(base, prefix)
}

// Validate refuses a project data directory Docker cannot bind - a relative
// path - or the root of the filesystem.
func (storage Storage) Validate() error {
	dir := storage.ProjectDataHostDir
	if dir == "" {
		return nil
	}
	if !filepath.IsAbs(dir) || filepath.Clean(dir) == "/" {
		return fmt.Errorf("%w: HP_STORAGE_PROJECT_DATA_HOST_DIR must be an absolute path below /, not %q",
			ErrStorageInvalid, dir)
	}
	return nil
}
