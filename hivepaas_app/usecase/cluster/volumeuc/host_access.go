package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
)

// checkVolumeHostAccess gates a volume that reaches the filesystem of the nodes
// it is mounted on.
//
// A volume is ordinarily docker's own storage. One that names a directory of a
// node - through bind options, or through the raw driver options that say the
// same thing - hands whatever app mounts it that path of the host, which is the
// decision Write on the Cluster module stands for. Holding Write on a project is
// not that decision: the project's own storage is what that permission covers.
//
// The docker socket is refused to everyone, whatever they hold. What an app may
// do with the Docker API is written down in its Docker API settings, and the
// proxy holds it to that; a mount of the socket is the unbounded version of the
// same thing, and would go around the settings that bound it.
//
// It runs before anything touches a node, so a refusal leaves no directory
// behind.
func (uc *UC) checkVolumeHostAccess(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.VolumeBaseReq,
) error {
	access := req.HostAccess()
	if access.ReachesDockerSocket {
		return hperrors.Wrap(hperrors.ErrArgumentInvalid).
			WithExtraDetail("%s holds the Docker socket of the node. An app is given the Docker API through "+
				"its own Docker API settings, which say what it may do with it, and never through a volume.",
				access.Directory).
			WithMsgLog("refused a volume reaching the docker socket at %s", access.Directory)
	}
	if !access.NamesHostPath {
		return nil
	}

	hasPerm, err := uc.PermissionManager.CheckAccess(ctx, uc.DB, auth, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleCluster,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !hasPerm {
		return hperrors.Wrap(hperrors.ErrUnauthorized).
			WithExtraDetail("a volume that reaches a directory of the node needs Write permission on the "+
				"Cluster module. A volume whose directory HivePaaS chooses needs no such permission.").
			WithMsgLog("volume naming the host path %s requires Write on the Cluster module", access.Directory)
	}
	return nil
}
