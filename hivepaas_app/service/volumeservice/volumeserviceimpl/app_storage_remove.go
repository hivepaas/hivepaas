package volumeserviceimpl

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// volumeHelperTarget is where the helper container sees the volume, whole.
const volumeHelperTarget = "/mnt/vol"

func (s *service) RemoveAppStorage(ctx context.Context, mounts []mount.Mount) error {
	for _, mnt := range mounts {
		subpath := mountSubpath(&mnt)
		// A mount with no subpath of its own is the whole volume, which belongs to
		// whoever created it and may be shared. Deleting an app is not permission
		// to empty that, so this leaves it alone.
		if subpath == "" {
			continue
		}
		if err := s.removeVolumeSubpath(ctx, &mnt, subpath); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// subpathPattern is what a directory inside a volume is allowed to look like:
// the names HivePaaS gives an app's storage, and nothing else. The path ends up
// in an rm -rf, so the rule is narrow on purpose.
var subpathPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)

// mountSubpath is the directory inside a volume an app was given. It is empty for
// a bind mount, for a volume mounted whole, and for anything that does not read
// as a plain relative path - a traversal, or a name carrying a character a shell
// would act on. Such a subpath is left in place rather than guessed at: what is
// above it belongs to somebody else.
func mountSubpath(mnt *mount.Mount) string {
	switch {
	case mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil:
		return safeSubpath(mnt.VolumeOptions.Subpath)
	case mnt.Type == mount.TypeCluster && mnt.VolumeOptions != nil:
		return safeSubpath(mnt.VolumeOptions.Subpath)
	default:
		return ""
	}
}

func safeSubpath(subpath string) string {
	trimmed := strings.Trim(subpath, "/")
	if !subpathPattern.MatchString(trimmed) {
		return ""
	}
	for _, segment := range strings.Split(trimmed, "/") {
		if segment == "." || segment == ".." {
			return ""
		}
	}
	return trimmed
}

// removeVolumeSubpath deletes one directory inside a volume, the same three ways
// EnsureVolumePermissions creates it: on the host when the volume is a directory
// this node can reach, then a container, then a swarm task for a volume that
// lives on another node.
func (s *service) removeVolumeSubpath(ctx context.Context, volMount *mount.Mount, subpath string) error {
	if hostPath, isDirect := s.getDirectHostPath(ctx, volMount, ""); isDirect && hostPath != "" {
		if err := os.RemoveAll(filepath.Join(hostPath, subpath)); err == nil {
			return nil
		}
	}

	// The helper sees the volume whole: the path to delete is inside it, and
	// mounting with the subpath would put the helper in the directory it is meant
	// to remove.
	targetMnt := *volMount
	targetMnt.Target = volumeHelperTarget
	targetMnt.ReadOnly = false
	if targetMnt.VolumeOptions != nil {
		targetMnt.VolumeOptions = new(*targetMnt.VolumeOptions)
		targetMnt.VolumeOptions.Subpath = ""
	}

	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)
	// No shell: the path is an argument of its own, so nothing in it can be read
	// as a command, and -- keeps a name starting with a dash from becoming a flag.
	rmCmd := []string{"rm", "-rf", "--", path.Join(volumeHelperTarget, subpath)}

	_, statusCode, err := s.dockerManager.ContainerCreateToExec(ctx, image, rmCmd,
		func(opts *client.ContainerCreateOptions) {
			opts.HostConfig.Mounts = []mount.Mount{targetMnt}
		})
	if err == nil && statusCode == 0 {
		return nil
	}

	_, statusCode, err = s.dockerManager.ServiceCreateToExec(ctx, image, rmCmd, 0, 0,
		func(opts *client.ServiceCreateOptions) {
			opts.Spec.TaskTemplate.ContainerSpec.Mounts = []mount.Mount{targetMnt}
		})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if statusCode != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
			"removing volume directory %q exited with status code %d", subpath, statusCode)
	}
	return nil
}
