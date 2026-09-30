package fileservice

import (
	"path/filepath"
)

const (
	// ProjectFilesDir is HivePaaS's own directory in a project's default volume:
	// no app mounts it, since an app's volumes are made in <project>/<env>/<app>.
	ProjectFilesDir = ".hivepaas"

	// The kinds of file HivePaaS keeps there.
	FilesDirUploads   = "files"
	FilesDirJobOutput = "job-output"
	FilesDirRepoCache = "cache/repos"
)

// AppFilePath is where a file of an app goes in its project's default volume:
// .hivepaas/<dir>/<env>/<app>/<name>.
func AppFilePath(dir, envKey, appKey, name string) string {
	return filepath.Join(ProjectFilesDir, dir, envKey, appKey, name)
}

// ProjectFilePath is where a file of a project goes in its default volume:
// .hivepaas/<dir>/<name>.
func ProjectFilePath(dir, name string) string {
	return filepath.Join(ProjectFilesDir, dir, name)
}
