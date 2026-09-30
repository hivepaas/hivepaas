// Package srcpack carries a checked-out source tree from the node that checked
// it out to the node that builds it: a tar stream compressed with zstd.
package srcpack

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	dirMode  = 0o755
	permMask = 0o777
)

// Stats is what a source tree holds: its regular files and their size.
type Stats struct {
	Files int
	Bytes int64
}

// Pack writes the tree under dir to w: its directories, its regular files with
// their mode, and its symbolic links as links. Anything else is left out.
func Pack(ctx context.Context, dir string, w io.Writer) (stats Stats, err error) {
	zw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		return stats, hperrors.Wrap(err)
	}
	tw := tar.NewWriter(zw)

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err //nolint:wrapcheck
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil || rel == "." {
			return err //nolint:wrapcheck
		}
		info, err := d.Info()
		if err != nil {
			return err //nolint:wrapcheck
		}
		header := &tar.Header{
			Name:    filepath.ToSlash(rel),
			Mode:    int64(info.Mode().Perm()),
			ModTime: info.ModTime(),
		}
		switch {
		case d.IsDir():
			header.Typeflag = tar.TypeDir
			header.Name += "/"
			return tw.WriteHeader(header) //nolint:wrapcheck
		case info.Mode()&fs.ModeSymlink != 0:
			header.Typeflag = tar.TypeSymlink
			if header.Linkname, err = os.Readlink(path); err != nil {
				return err //nolint:wrapcheck
			}
			return tw.WriteHeader(header) //nolint:wrapcheck
		case info.Mode().IsRegular():
			header.Typeflag = tar.TypeReg
			header.Size = info.Size()
			if err = tw.WriteHeader(header); err != nil {
				return err //nolint:wrapcheck
			}
			if err = copyFile(ctx, tw, path); err != nil {
				return err
			}
			stats.Files++
			stats.Bytes += info.Size()
		}
		return nil
	})
	if err != nil {
		_ = zw.Close()
		return stats, hperrors.Wrap(err)
	}
	if err = tw.Close(); err != nil {
		_ = zw.Close()
		return stats, hperrors.Wrap(err)
	}
	return stats, hperrors.Wrap(zw.Close())
}

func copyFile(ctx context.Context, w io.Writer, path string) error {
	file, err := os.Open(path) //nolint:gosec // a file of the tree being packed
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer file.Close()
	_, err = io.Copy(w, &ctxReader{ctx: ctx, r: file})
	return hperrors.Wrap(err)
}

// Unpack reads a tree written by Pack into dir, which must exist. It trusts
// nothing in the stream: an entry that would be written outside dir - by its
// path, or through a symbolic link unpacked before it - is refused, and so is
// any entry that is not a file, a directory or a symbolic link.
func Unpack(ctx context.Context, r io.Reader, dir string) (stats Stats, err error) {
	zr, err := zstd.NewReader(r)
	if err != nil {
		return stats, hperrors.Wrap(err)
	}
	defer zr.Close()
	tr := tar.NewReader(zr)
	u := &unpacker{root: dir, dirs: map[string]bool{".": true}}

	for {
		if err = ctx.Err(); err != nil {
			return stats, hperrors.Wrap(err)
		}
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return stats, hperrors.Wrap(err)
		}
		size, err := u.entry(ctx, header, tr)
		if err != nil {
			return stats, hperrors.Wrap(err)
		}
		if header.Typeflag == tar.TypeReg {
			stats.Files++
			stats.Bytes += size
		}
	}
	// To the end of the stream, so a stream cut after the archive is an error too.
	if _, err = io.Copy(io.Discard, zr); err != nil {
		return stats, hperrors.Wrap(err)
	}
	return stats, nil
}

type unpacker struct {
	root string
	// dirs are the directories known to be real directories inside root.
	dirs map[string]bool
}

func (u *unpacker) entry(ctx context.Context, header *tar.Header, content io.Reader) (int64, error) {
	rel := filepath.Clean(filepath.FromSlash(header.Name))
	if header.Name == "" || filepath.IsAbs(filepath.FromSlash(header.Name)) || rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return 0, u.outside(header.Name)
	}
	if rel == "." {
		return 0, nil
	}
	target := filepath.Join(u.root, rel)

	switch header.Typeflag {
	case tar.TypeDir:
		return 0, u.ensureDir(rel, header.Name)
	case tar.TypeSymlink:
		if err := u.ensureFree(rel, header.Name); err != nil {
			return 0, err
		}
		return 0, hperrors.Wrap(os.Symlink(header.Linkname, target))
	case tar.TypeReg:
		if err := u.ensureFree(rel, header.Name); err != nil {
			return 0, err
		}
		return u.writeFile(ctx, target, header, content)
	default:
		return 0, hperrors.NewUnsupported("Source entry '" + header.Name + "'")
	}
}

func (u *unpacker) writeFile(
	ctx context.Context, target string, header *tar.Header, content io.Reader,
) (_ int64, err error) {
	mode := os.FileMode(header.Mode & permMask) //nolint:gosec // masked to the permission bits
	// O_EXCL: never through something already there, a symbolic link above all.
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode) //nolint:gosec // held inside root
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	size, err := io.Copy(file, &ctxReader{ctx: ctx, r: content})
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	// MkdirTemp and umask do not decide a source file's mode: the sender does.
	if err = os.Chmod(target, mode); err != nil {
		return 0, hperrors.Wrap(err)
	}
	if !header.ModTime.IsZero() {
		_ = os.Chtimes(target, header.ModTime, header.ModTime)
	}
	return size, nil
}

// ensureFree makes the directory an entry goes in, and refuses an entry whose
// place is taken: by a symbolic link, which writing would follow out of root,
// or by anything else.
func (u *unpacker) ensureFree(rel, name string) error {
	if err := u.ensureDir(filepath.Dir(rel), name); err != nil {
		return err
	}
	info, err := os.Lstat(filepath.Join(u.root, rel))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return hperrors.Wrap(err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return u.outside(name)
	}
	return hperrors.NewAlreadyExist("Source entry '" + name + "'")
}

// ensureDir makes rel a directory inside root, one component at a time, and
// refuses to pass through anything that is not a real directory: a symbolic
// link would lead out of root.
func (u *unpacker) ensureDir(rel, name string) error {
	if u.dirs[rel] {
		return nil
	}
	if err := u.ensureDir(filepath.Dir(rel), name); err != nil {
		return err
	}
	path := filepath.Join(u.root, rel)
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err = os.Mkdir(path, dirMode); err != nil {
			return hperrors.Wrap(err)
		}
	case err != nil:
		return hperrors.Wrap(err)
	case !info.IsDir():
		return u.outside(name)
	}
	u.dirs[rel] = true
	return nil
}

func (u *unpacker) outside(name string) error {
	return hperrors.Wrap(hperrors.ErrFilePathOutsideRoot).WithParam("Path", name).WithParam("Root", u.root)
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
