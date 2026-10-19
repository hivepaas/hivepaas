package volumeserviceimpl

import (
	"context"
	"path"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// A volume pinned to a node keeps its data there, and an app's directory in it is
// made there: this node's docker or filesystem reaches another node's no more
// than the node's own. A helper run here would make the directory in a volume of
// the same name on this node, and the app's task, placed on the pinned node,
// would find none.

// onThisNode says whether a volume's directories are made from this node: one
// with no pin - every node reaches its data, by its own claim - and one pinned to
// this node by its id. A pin by label is reached through Swarm, which knows
// which nodes carry it.
func (s *service) onThisNode(ctx context.Context, vol *entity.ClusterVolume) bool {
	if !vol.IsPinned() {
		return true
	}
	if vol.NodeID == "" {
		return false
	}
	current, err := s.dockerManager.NodeCurrentID(ctx)
	return err == nil && current == vol.NodeID
}

// ensureVolumePermissionsOnNode is EnsureVolumePermissions on the volume's node:
// on another one, in a task Swarm places there, whose mount of the volume has
// Docker make it there, with the volume's driver options, when it is not yet.
func (s *service) ensureVolumePermissionsOnNode(
	ctx context.Context,
	vol *entity.ClusterVolume,
	volMount *mount.Mount,
	subpaths ...string,
) error {
	if volMount == nil || volMount.Source == "" {
		return nil
	}
	if s.onThisNode(ctx, vol) {
		return s.EnsureVolumePermissions(ctx, volMount, subpaths...)
	}
	constraint, _ := placementservice.VolumePinConstraint([]placementservice.VolumePin{
		{VolumeName: volMount.Source, NodeID: vol.NodeID, NodeLabel: vol.NodeLabel}})
	targetMnt, shCmd := volumePermissionsCommand(volMount, subpaths)
	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)
	_, statusCode, err := s.dockerManager.ServiceCreateToExec(ctx, image, shCmd, 0, 0,
		func(opts *client.ServiceCreateOptions) {
			opts.Spec.TaskTemplate.ContainerSpec.Mounts = []mount.Mount{targetMnt}
			opts.Spec.TaskTemplate.Placement = &swarm.Placement{Constraints: []string{constraint}}
		},
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if statusCode != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
			"volume init task on %s exited with status code %d", constraint, statusCode)
	}
	return nil
}

// makeSubDirOnNode is MakeSubDirInHost on the volume's node: on another one,
// through its agent, which sees the node's filesystem under HostPathPrefix.
func (s *service) makeSubDirOnNode(
	ctx context.Context,
	vol *entity.ClusterVolume,
	baseDirInHost string,
	subpath string,
	requireBaseDirExist bool,
) error {
	if s.onThisNode(ctx, vol) {
		return s.MakeSubDirInHost(ctx, baseDirInHost, subpath, requireBaseDirExist)
	}
	base := path.Join(volumeservice.HostPathPrefix, baseDirInHost)
	var cmd strings.Builder
	if requireBaseDirExist {
		cmd.WriteString("test -d " + volumeservice.QuoteArg(base) + " && ")
	}
	cmd.WriteString(volumeservice.MakeDirWritableCmd(path.Join(base, strings.TrimPrefix(subpath, "/"))))
	resp, err := s.nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
		NodeID:          vol.NodeID,
		NodeLabel:       vol.NodeLabel,
		CommandExecOpts: &nodeexecservice.CommandExecOpts{Command: []string{"sh", "-c", cmd.String()}},
	})
	if err != nil || resp == nil || resp.ExitCode != 0 {
		return hperrors.Wrap(hperrors.ErrDirNotCreated).WithParam("Name", path.Join(baseDirInHost, subpath))
	}
	return nil
}
