package artifactservice

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// downloadPrefix starts the temp files an agent downloads into, beside where
// the file will be.
const downloadPrefix = ".download-"

// PruneCache removes from a node's artifacts cache, dir as this process sees
// it, every file no kept hash names and nothing has touched since olderThan -
// temp files an agent killed mid-download left among them. The retention keeps
// a file downloaded for an app still being created, whose service does not bind
// it yet.
func PruneCache(dir string, keep []string, olderThan time.Time) (deleted int, reclaimed uint64, err error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || slices.Contains(keep, name) {
			continue
		}
		if strings.HasPrefix(name, downloadPrefix) {
			sha, _, _ := strings.Cut(strings.TrimPrefix(name, downloadPrefix), "-")
			if slices.Contains(keep, sha) {
				continue
			}
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.ModTime().Before(olderThan) {
			continue
		}
		if removeErr := os.Remove(filepath.Join(dir, name)); removeErr != nil {
			err = removeErr
			continue
		}
		deleted++
		reclaimed += uint64(info.Size()) //nolint:gosec // a file's size is not negative
	}
	return deleted, reclaimed, hperrors.Wrap(err)
}
