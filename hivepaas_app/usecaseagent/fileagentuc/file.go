package fileagentuc

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// dirModeOwn is for the directories a file's path needs: HivePaaS's own.
const dirModeOwn = 0o700

// Open reads a file inside root, a volume's directory on the host, and gives its size.
func (uc *UC) Open(root, path string) (io.ReadCloser, int64, error) {
	target, err := uc.existingTarget(root, path)
	if err != nil {
		return nil, 0, hperrors.Wrap(err)
	}
	file, err := os.Open(target) //nolint:gosec // held inside root above
	if err != nil {
		return nil, 0, notFoundOr(err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, 0, hperrors.Wrap(err)
	}
	return file, info.Size(), nil
}

// Stat is the size of a file inside root.
func (uc *UC) Stat(root, path string) (int64, error) {
	target, err := uc.existingTarget(root, path)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return 0, notFoundOr(err)
	}
	return info.Size(), nil
}

// Remove removes a file inside root. A file already gone is removed.
func (uc *UC) Remove(root, path string) error {
	target, err := uc.existingTarget(root, path)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil
		}
		return hperrors.Wrap(err)
	}
	if err = os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return hperrors.Wrap(err)
	}
	return nil
}

// Write writes content to a file inside root, making the directories it needs
// with mode 0700. It writes a temporary file next to it and renames it into
// place once content has ended without error: the file is replaced whole or not
// at all, and a failed write leaves nothing behind.
func (uc *UC) Write(ctx context.Context, root, path string, content io.Reader) (_ int64, err error) {
	realRoot, rel, err := uc.resolveRoot(root, path)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	target := filepath.Join(realRoot, rel)
	if err = holdInside(realRoot, nearestExisting(filepath.Dir(target)), path); err != nil {
		return 0, hperrors.Wrap(err)
	}
	if err = os.MkdirAll(filepath.Dir(target), dirModeOwn); err != nil {
		return 0, hperrors.Wrap(err)
	}
	if err = holdInside(realRoot, filepath.Dir(target), path); err != nil {
		return 0, hperrors.Wrap(err)
	}

	temp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	defer func() {
		if err != nil {
			_ = temp.Close()
			_ = os.Remove(temp.Name())
		}
	}()

	size, err := io.Copy(temp, &ctxReader{ctx: ctx, r: content})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	if err = temp.Close(); err != nil {
		return 0, hperrors.Wrap(err)
	}
	if err = os.Rename(temp.Name(), target); err != nil {
		return 0, hperrors.Wrap(err)
	}
	return size, nil
}

// existingTarget is path inside root, resolved through its symbolic links, which
// must still be inside root.
func (uc *UC) existingTarget(root, path string) (string, error) {
	realRoot, rel, err := uc.resolveRoot(root, path)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(realRoot, rel))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// A path whose existing part leaves root is refused, not reported missing.
			if err := holdInside(realRoot, nearestExisting(filepath.Join(realRoot, rel)), path); err != nil {
				return "", hperrors.Wrap(err)
			}
		}
		return "", notFoundOr(err)
	}
	if err = holdInside(realRoot, target, path); err != nil {
		return "", hperrors.Wrap(err)
	}
	return target, nil
}

// resolveRoot is root as this agent sees it, with its symbolic links resolved,
// and path cleaned. The root is a volume's absolute directory, never the host's
// own root; the path is relative to it and never climbs out.
func (uc *UC) resolveRoot(root, path string) (realRoot, rel string, err error) {
	outside := hperrors.Wrap(hperrors.ErrFilePathOutsideRoot).WithParam("Path", path).WithParam("Root", root)
	cleanRoot := filepath.Clean(root)
	if !filepath.IsAbs(root) || cleanRoot == "/" {
		return "", "", outside
	}
	rel = filepath.Clean(path)
	if path == "" || filepath.IsAbs(path) || rel == "." || rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", outside
	}
	realRoot, err = filepath.EvalSymlinks(filepath.Join(uc.hostPrefix, cleanRoot))
	if err != nil {
		return "", "", notFoundOr(err)
	}
	return realRoot, rel, nil
}

// holdInside refuses a resolved path that is not realRoot or below it.
func holdInside(realRoot, resolved, path string) error {
	real, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return notFoundOr(err)
	}
	rel, err := filepath.Rel(realRoot, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return hperrors.Wrap(hperrors.ErrFilePathOutsideRoot).WithParam("Path", path).WithParam("Root", realRoot)
	}
	return nil
}

// nearestExisting is path, or the closest of its parents that exists.
func nearestExisting(path string) string {
	for {
		if _, err := os.Lstat(path); err == nil {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return path
		}
		path = parent
	}
}

func notFoundOr(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return hperrors.NewNotFound("File")
	}
	return hperrors.Wrap(err)
}

// ctxReader stops a copy once its context is done.
type ctxReader struct {
	ctx context.Context //nolint:containedctx // bounds one copy
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err //nolint:wrapcheck
	}
	return c.r.Read(p) //nolint:wrapcheck
}
