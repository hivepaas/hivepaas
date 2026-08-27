package volumeserviceimpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	if subpath != "" {
		subpath = strings.TrimPrefix(subpath, "/")
		_, _ = fmt.Fprintf(&cmdBuilder, "mkdir -p '/mnt/data/%s' && chmod -R 777 '/mnt/data/%s'",
			subpath, subpath)
	} else {
		cmdBuilder.WriteString("chmod 777 /mnt/data")
	}
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
