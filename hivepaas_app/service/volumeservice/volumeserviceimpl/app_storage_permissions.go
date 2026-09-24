package volumeserviceimpl

import (
	"context"
	"io/fs"
	"os"
	"strconv"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) ResetAppStoragePermissions(
	ctx context.Context,
	db database.IDB,
	req *volumeservice.ResetAppStoragePermissionsReq,
) (*volumeservice.ResetAppStoragePermissionsResp, error) {
	found, err := s.findAppStorage(ctx, db, req.App, []mount.Mount{req.Mount})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(found.targets) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrStoragePermissionsNotResettable).
			WithParam("Target", req.Mount.Target)
	}

	target := found.targets[0]
	err = s.runOnStorageTarget(ctx, &target, found, &storageTargetAction{
		onHost: func(dir string) error { return resetPermissionsOnHost(dir, req.Owner) },
		cmd:    func(dir string) []string { return resetPermissionsCmd(dir, req.Owner) },
		scoped: true,
		failure: func(statusCode int64) error {
			return hperrors.Wrap(hperrors.ErrStoragePermissionsNotReset).WithParam("Path", target.subpath).
				WithMsgLog("resetting permissions of %q exited with status code %d", target.subpath, statusCode)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &volumeservice.ResetAppStoragePermissionsResp{Path: target.subpath}, nil
}

// resetPermissionsCmd is what a helper runs to reset a directory, which it is
// given on its own (storageTargetAction.scoped).
//
// Only directories and regular files are changed, the directory itself
// included. A symlink is never followed: the app wrote it, and it may point
// anywhere. And a socket, a fifo or a device keeps what it has, because its
// permissions mean nothing to the app that reopens it, and on a Docker Desktop
// bind mount a socket cannot be touched at all.
//
// The app may be running while this runs, and could turn a directory into a
// symlink between find seeing it and chmod following it. That is why the helper
// sees the directory alone: whatever such a link points at is inside the helper,
// which has nothing of anybody's.
//
// find goes on past a file it cannot change and then exits non-zero, so one such
// file costs the whole reset its success but not the rest of the files.
func resetPermissionsCmd(dir string, owner *volumeservice.StorageOwner) []string {
	change := []string{"chmod", "a+rwX"}
	if owner != nil {
		change = []string{"chown", strconv.Itoa(owner.UID) + ":" + strconv.Itoa(owner.GID)}
	}
	cmd := []string{"find", dir, "(", "-type", "d", "-o", "-type", "f", ")", "-exec"}
	cmd = append(cmd, change...)
	return append(cmd, "{}", "+")
}

// resetPermissionsOnHost is resetPermissionsCmd for a directory this process
// reaches itself, by the same rules. Everything goes through an os.Root on the
// directory, which is what keeps a link the app swaps in mid-walk from reaching
// anything outside it.
func resetPermissionsOnHost(dir string, owner *volumeservice.StorageOwner) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = root.Close() }()

	failed := 0
	err = fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			failed++
			return nil //nolint:nilerr // counted, and the walk goes on as find's does
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return nil
		}
		if owner != nil {
			if root.Lchown(name, owner.UID, owner.GID) != nil {
				failed++
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			failed++
			return nil //nolint:nilerr // counted, and the walk goes on as find's does
		}
		if root.Chmod(name, openedMode(info.Mode().Perm(), entry.IsDir())) != nil {
			failed++
		}
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if failed > 0 {
		return hperrors.Wrap(hperrors.ErrStoragePermissionsNotReset).WithParam("Path", dir)
	}
	return nil
}

// openedMode is chmod a+rwX: everyone may read and write, and may execute what
// somebody already could - a directory, or a script - but a data file does not
// become a program.
func openedMode(mode os.FileMode, isDir bool) os.FileMode {
	mode |= 0o666
	if isDir || mode&0o111 != 0 {
		mode |= 0o111
	}
	return mode
}
