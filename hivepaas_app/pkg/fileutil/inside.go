package fileutil

import (
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// DirInside is the directory rel names in root: root as it is given, then the
// path in it with links resolved. It is refused when it leads out of root - by
// its own "..", or through a link to a directory out of it, which a repository
// can hold - and when it is not a directory: a build given it as its context
// could copy whatever it names.
func DirInside(root, rel string) (string, error) {
	if !filepath.IsLocal(rel) {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("path %q leads out of its directory", rel)
	}
	base, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	dir, err := filepath.EvalSymlinks(filepath.Join(base, rel))
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("path %q: %v", rel, err)
	}
	inside, err := filepath.Rel(base, dir)
	if err != nil || !filepath.IsLocal(inside) {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("path %q leads out of its directory", rel)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("path %q is not a directory", rel)
	}
	return filepath.Join(root, inside), nil
}
