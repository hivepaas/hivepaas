package volumeserviceimpl

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) MakeSubDirInHost(
	ctx context.Context,
	baseDirInHost string,
	subpath string,
	requireBaseDirExist bool,
) error {
	targetMnt := mount.Mount{
		Type:   mount.TypeBind,
		Source: baseDirInHost,
		Target: "/mnt/data",
	}

	var cmdBuilder strings.Builder
	if requireBaseDirExist {
		cmdBuilder.WriteString("test -d /mnt/data && ")
	}
	subpath = strings.TrimPrefix(subpath, "/")
	cmdBuilder.WriteString(volumeservice.MakeDirWritableCmd(path.Join("/mnt/data", subpath)))
	shCmd := []string{"sh", "-c", cmdBuilder.String()}

	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)

	_, statusCode, err := s.dockerManager.ContainerCreateToExec(ctx, image, shCmd,
		func(opts *client.ContainerCreateOptions) {
			opts.HostConfig.Mounts = []mount.Mount{targetMnt}
		})
	if err != nil || statusCode != 0 {
		return hperrors.Wrap(hperrors.ErrDirNotCreated).
			WithParam("Name", filepath.Join(baseDirInHost, subpath))
	}
	return nil
}
