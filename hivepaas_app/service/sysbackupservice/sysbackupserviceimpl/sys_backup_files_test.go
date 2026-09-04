package sysbackupserviceimpl

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

// A backup carries the database dump, whose secrets are encrypted with the app
// secret. That secret lives in the app path, so putting the app path in the
// archive would ship the lock and the key together and make the encryption
// pointless. Anything added here must stay outside it.
func TestSysBackupExcludesTheAppPath(t *testing.T) {
	config.Current = &config.Config{AppPath: "/var/lib/hivepaas"}
	appPath := filepath.Clean(config.Current.AppPath)

	for _, model := range sysBackupFileModels {
		if model.DirPath == nil {
			t.Errorf("backup file model %q has no DirPath", model.Type)
			continue
		}
		dirPath := filepath.Clean(model.DirPath())
		if dirPath == appPath || strings.HasPrefix(dirPath, appPath+string(filepath.Separator)) {
			t.Errorf("backup file model %q reads %q, inside the app path %q: the app secret"+
				" would be shipped in the same archive as the data it encrypts",
				model.Type, dirPath, appPath)
		}
	}
}
