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

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

// EnsureVolumePermissions guarantees that a volume root and any requested subpaths
// have full read/write/execute permissions (0777) for any container user (Root or Non-Root).
// It uses a 3-tier strategy:
//  1. Fast-Path: Direct host filesystem access (~1ms)
//  2. High-Performance ContainerCreate (~100-500ms, direct standalone container with AutoRemove)
//  3. Swarm Service Fallback (more than 2s, for multi-node distributed cluster volumes)
func (s *service) EnsureVolumePermissions(
	ctx context.Context,
	volMount *mount.Mount,
	subpaths ...string,
) error {
	if volMount == nil || volMount.Source == "" {
		return nil
	}

	// 1. Fast-Path: If the volume points to a directory directly accessible on the host filesystem (~ms)
	if hostPath, isDirect := s.getDirectHostPath(ctx, volMount, ""); isDirect && hostPath != "" {
		if s.ensurePermissionsOnDirectHostPath(hostPath, subpaths...) == nil {
			return nil
		}
	}

	// Determine helper image
	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)

	targetMnt := *volMount
	targetMnt.Target = "/mnt/vol"
	targetMnt.ReadOnly = false
	if targetMnt.BindOptions != nil {
		targetMnt.BindOptions = new(*targetMnt.BindOptions)
	}
	if targetMnt.VolumeOptions != nil {
		targetMnt.VolumeOptions = new(*targetMnt.VolumeOptions)
		targetMnt.VolumeOptions.Subpath = ""
	}

	var cmdBuilder strings.Builder
	cmdBuilder.WriteString("chmod 777 /mnt/vol")
	for _, sub := range subpaths {
		if sub == "" {
			continue
		}
		_, _ = fmt.Fprintf(&cmdBuilder, " && mkdir -p '/mnt/vol/%s' && chmod -R 777 '/mnt/vol/%s'", sub, sub)
	}
	shCmd := []string{"sh", "-c", cmdBuilder.String()}

	// 2. High-Performance ContainerCreate (~100ms, direct container with AutoRemove)
	_, statusCode, err := s.dockerManager.ContainerCreateToExec(ctx, image, shCmd,
		func(opts *client.ContainerCreateOptions) {
			opts.HostConfig.Mounts = []mount.Mount{targetMnt}
		})
	if err == nil && statusCode == 0 {
		return nil
	}

	// 3. Swarm Service Fallback (for multi-node Swarm cluster volumes located on remote nodes)
	_, statusCode, err = s.dockerManager.ServiceCreateToExec(ctx, image, shCmd, 0, 0,
		func(opts *client.ServiceCreateOptions) {
			opts.Spec.TaskTemplate.ContainerSpec.Mounts = []mount.Mount{targetMnt}
		},
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if statusCode != 0 {
		return apperrors.Wrap(apperrors.ErrActionFailed).WithMsgLog(
			"volume init swarm task exited with status code %d", statusCode)
	}
	return nil
}

func (s *service) ensurePermissionsOnDirectHostPath(
	baseHostPath string,
	subpaths ...string,
) (err error) {
	if err = os.MkdirAll(baseHostPath, fullFileMode); err != nil {
		return apperrors.Wrap(err)
	}
	if err = os.Chmod(baseHostPath, fullFileMode); err != nil {
		return apperrors.Wrap(err)
	}
	for _, sub := range subpaths {
		if sub == "" {
			continue
		}
		subDir := filepath.Join(baseHostPath, sub)
		if err = os.MkdirAll(subDir, fullFileMode); err != nil {
			return apperrors.Wrap(err)
		}
		if err = os.Chmod(subDir, fullFileMode); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
