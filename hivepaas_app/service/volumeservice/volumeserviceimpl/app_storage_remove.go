package volumeserviceimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

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

// mountSubpath is the directory inside a volume an app was given. It is empty for
// a bind mount and for a volume mounted whole.
func mountSubpath(mnt *mount.Mount) string {
	switch {
	case mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil:
		return strings.Trim(mnt.VolumeOptions.Subpath, "/")
	case mnt.Type == mount.TypeCluster && mnt.VolumeOptions != nil:
		return strings.Trim(mnt.VolumeOptions.Subpath, "/")
	default:
		return ""
	}
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
	targetMnt.Target = "/mnt/vol"
	targetMnt.ReadOnly = false
	if targetMnt.VolumeOptions != nil {
		targetMnt.VolumeOptions = new(*targetMnt.VolumeOptions)
		targetMnt.VolumeOptions.Subpath = ""
	}

	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)
	shCmd := []string{"sh", "-c", fmt.Sprintf("rm -rf '/mnt/vol/%s'", subpath)}

	_, statusCode, err := s.dockerManager.ContainerCreateToExec(ctx, image, shCmd,
		func(opts *client.ContainerCreateOptions) {
			opts.HostConfig.Mounts = []mount.Mount{targetMnt}
		})
	if err == nil && statusCode == 0 {
		return nil
	}

	_, statusCode, err = s.dockerManager.ServiceCreateToExec(ctx, image, shCmd, 0, 0,
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
