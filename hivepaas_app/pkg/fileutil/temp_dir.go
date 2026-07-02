package fileutil

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// CreateTempDir creates a temp dir.
// Should use "*" for `pattern` value. If empty, only base dir is created. See os.MkdirTemp.
func CreateTempDir(baseDir, pattern string, perm os.FileMode) (dir string, err error) {
	if perm == 0 {
		perm = defaultDirMode
	}
	dir = filepath.Join(baseDir, timeutil.NowUTC().Format(time.DateOnly))

	err = os.MkdirAll(dir, perm)
	if err != nil {
		return "", apperrors.New(err)
	}

	if pattern != "" {
		dir, err = os.MkdirTemp(dir, pattern)
		if err != nil {
			return "", apperrors.New(err)
		}
	}

	return dir, nil
}

func CreateTempDirInAppPath(baseDir, pattern string, perm os.FileMode) (dir string, err error) {
	if perm == 0 {
		perm = defaultDirMode
	}
	dir = filepath.Join(config.Current.AppPath, "tmp", timeutil.NowUTC().Format(time.DateOnly), baseDir)

	err = os.MkdirAll(dir, perm)
	if err != nil {
		return "", apperrors.New(err)
	}

	if pattern != "" {
		dir, err = os.MkdirTemp(dir, pattern)
		if err != nil {
			return "", apperrors.New(err)
		}
	}

	return dir, nil
}
