package fileutil

import (
	"path/filepath"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func NormalizePath(path string) string {
	return filepath.Clean(path)
}

func IsSubpath(basepath, targetpath string) (bool, error) {
	// Rel returns a relative path that is lexically equivalent to targetpath when joined to basepath
	rel, err := filepath.Rel(basepath, targetpath)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	// Two paths are the same
	if rel == "." {
		return false, nil
	}
	// A result starting with `..` means the target is outside the base directory
	if strings.HasPrefix(rel, "..") {
		return false, nil
	}
	return true, nil
}

func IsEqualOrSubpath(basepath, targetpath string) (bool, error) {
	// Rel returns a relative path that is lexically equivalent to targetpath when joined to basepath
	rel, err := filepath.Rel(basepath, targetpath)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	if rel == "." {
		return true, nil
	}
	// A result starting with `..` means the target is outside the base directory
	if strings.HasPrefix(rel, "..") {
		return false, nil
	}
	return true, nil
}

func IsSamePath(path1, path2 string) (bool, error) {
	// Rel returns a relative path that is lexically equivalent to path2 when joined to path1
	rel, err := filepath.Rel(path1, path2)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	return rel == ".", nil
}

func PathContain(listPaths []string, aPath string) (bool, error) {
	for _, path := range listPaths {
		same, err := IsSamePath(path, aPath)
		if err != nil {
			return false, apperrors.Wrap(err)
		}
		if same {
			return true, nil
		}
	}
	return false, nil
}
