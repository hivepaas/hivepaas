package volumeserviceimpl

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// EnsureVolumePermissions creates a volume's root and the requested subpaths and
// opens each one up to any container user while it is still empty - see
// volumeservice.MakeDirWritableCmd for why only then. It uses a 3-tier strategy:
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

	cmds := []string{volumeservice.MakeDirWritableCmd("/mnt/vol")}
	for _, sub := range subpaths {
		if sub == "" {
			continue
		}
		cmds = append(cmds, volumeservice.MakeDirWritableCmd(path.Join("/mnt/vol", sub)))
	}
	shCmd := []string{"sh", "-c", strings.Join(cmds, " && ")}

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
		return hperrors.Wrap(err)
	}
	if statusCode != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
			"volume init swarm task exited with status code %d", statusCode)
	}
	return nil
}

func (s *service) ensurePermissionsOnDirectHostPath(
	baseHostPath string,
	subpaths ...string,
) error {
	if err := makeDirWritable(baseHostPath); err != nil {
		return err
	}
	for _, sub := range subpaths {
		if sub == "" {
			continue
		}
		if err := makeDirWritable(filepath.Join(baseHostPath, sub)); err != nil {
			return err
		}
	}
	return nil
}

// makeDirWritable is volumeservice.MakeDirWritableCmd for a directory this
// process reaches itself: it creates the directory and opens it up only while
// it is empty.
func makeDirWritable(dir string) error {
	if err := os.MkdirAll(dir, fullFileMode); err != nil {
		return hperrors.Wrap(err)
	}
	f, err := os.Open(dir)
	if err != nil {
		return hperrors.Wrap(err)
	}
	names, err := f.Readdirnames(1)
	_ = f.Close()
	if len(names) > 0 {
		return nil // the app has written to it
	}
	if !errors.Is(err, io.EOF) {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(os.Chmod(dir, fullFileMode))
}
