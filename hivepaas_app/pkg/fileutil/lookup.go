package fileutil

import "path/filepath"

func Lookup(filename string, lookupDirs []string) (filePath, dirPath string) {
	for _, dir := range lookupDirs {
		path := filepath.Join(dir, filename)
		if exists, err := FileExists(path, true); err == nil && exists {
			return path, dir
		}
	}
	return "", ""
}
