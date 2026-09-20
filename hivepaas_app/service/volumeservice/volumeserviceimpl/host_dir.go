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
		cmdBuilder.WriteString(makeDirWritableCmd("/mnt/data/" + subpath))
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

// makeDirWritableCmd is the shell that makes a directory usable by a container
// running as any user: the directory itself, and then what is already inside it.
//
// The directory is the part that has to work. Everything below it is attempted
// and forgiven, because `chmod -R` walks into whatever an app left behind, and a
// unix socket on a Docker Desktop bind mount cannot be touched at all - it is
// listed, and then every operation on it answers "No such file or directory".
// Two sockets left by a deleted GitLab app were enough to make creating a
// project with that key fail forever, which is what this shape prevents. The
// same reasoning is why sockets, fifos and devices are skipped rather than
// chmod'd: the permissions of a socket mean nothing to the app that reopens it.
func makeDirWritableCmd(dir string) string {
	return fmt.Sprintf("mkdir -p '%s' && chmod 777 '%s'"+
		" && { find '%s' -mindepth 1 \\( -type d -o -type f \\) -exec chmod 777 {} + || true; }",
		dir, dir, dir)
}
