package fileutil

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return "", hperrors.Wrap(err)
	}

	if pattern != "" {
		dir, err = os.MkdirTemp(dir, pattern)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
	}

	return dir, nil
}

// CreateTempDirInAppPath creates a temp dir under the day's directory of
// AppTempDir. It is where the app and the worker keep what they write for a
// while: their data directory outlives their container - one whose process
// died is replaced, and what its /tmp held stays with the old one, out of
// reach - and the system cleanup removes its days after TempDirRetentionDays.
// Should use "*" for `pattern` value. If empty, only the day's dir is created.
func CreateTempDirInAppPath(baseDir, pattern string, perm os.FileMode) (dir string, err error) {
	if perm == 0 {
		perm = defaultDirMode
	}
	dir = filepath.Join(AppTempDir(), timeutil.NowUTC().Format(time.DateOnly), baseDir)

	err = os.MkdirAll(dir, perm)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	if pattern != "" {
		dir, err = os.MkdirTemp(dir, pattern)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
	}

	return dir, nil
}

// AppTempDir is tmp in the app's data directory; with none configured, as in
// a test, HivePaaS's directory under the OS's temporary one.
func AppTempDir() string {
	if cfg := config.Current(); cfg != nil && cfg.AppPath != "" {
		return filepath.Join(cfg.AppPath, "tmp")
	}
	return filepath.Join(os.TempDir(), "hivepaas")
}

// TempDirRetentionDays is how long a day's temporary directories are kept by the
// cleanups, in case whatever made them is still running.
const TempDirRetentionDays = 3

// RemoveDatedTempDirs removes the directories of baseDir named after a day
// (YYYY-MM-DD, as CreateTempDir makes them) earlier than before, and says how
// many went. Anything else in baseDir is left alone.
func RemoveDatedTempDirs(baseDir string, before time.Time) (removed int, err error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, hperrors.Wrap(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		day, err := time.Parse(time.DateOnly, entry.Name())
		if err != nil || !day.Before(before) {
			continue
		}
		if os.RemoveAll(filepath.Join(baseDir, entry.Name())) == nil {
			removed++
		}
	}
	return removed, nil
}
