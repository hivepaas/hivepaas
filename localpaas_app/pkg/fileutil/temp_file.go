package fileutil

import (
	"os"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

const (
	defaultDirMode  = 0o755
	defaultFileMode = 0o600
)

func WriteTempFile(dir, pattern string, perm os.FileMode, data []byte) (
	path string, cleanup func() error, err error) {
	if perm == 0 {
		perm = defaultFileMode
	}
	if dir == "" {
		dir, err = CreateTempDir("", "", defaultDirMode)
		if err != nil {
			return "", nil, apperrors.Wrap(err)
		}
	}
	fh, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", nil, apperrors.Wrap(err)
	}
	path = fh.Name()

	// 1. Remove the file if an error occurs (runs AFTER fh.Close())
	defer func() {
		if err != nil && path != "" {
			_ = os.Remove(path)
		}
	}()

	// 2. Always close the file descriptor to prevent leaks (runs FIRST)
	defer fh.Close()

	// Set the permissions
	if perm != 0o600 { //nolint:mnd
		if err = os.Chmod(path, perm); err != nil {
			return "", nil, apperrors.Wrap(err)
		}
	}

	_, err = fh.Write(data)
	if err != nil {
		return "", nil, apperrors.Wrap(err)
	}

	return path, func() error { return os.Remove(path) }, nil
}
