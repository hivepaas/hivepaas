package volumeservice

import (
	"context"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// HostDir is where a volume's data is on the host, and the node to reach it on.
type HostDir struct {
	Dir       string
	NodeID    string
	NodeLabel string
	// Shared is a volume every node sees alike, reached here on the node asking.
	Shared bool
}

// ResolveHostDir is where a volume's data is on the host, and the node to reach
// it on. Backup repositories and files on volumes both go through it, so the two
// find a volume the same way.
//
// A pinned volume is on its node. A volume on all nodes is taken at its word that
// every node sees the same data: it is reached on the node asking, and only when
// it is a bind directory - the one kind whose path is the same on every node. A
// docker-managed volume is somewhere under each daemon's own root, and one a
// driver mounts (NFS, CIFS) is on no host path at all.
func ResolveHostDir(ctx context.Context, dockerManager docker.Manager, setting *entity.Setting) (*HostDir, error) {
	clusterVolume, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if clusterVolume.IsPinned() {
		dir, err := resolveVolumeHostPath(ctx, dockerManager, setting, clusterVolume)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return &HostDir{Dir: dir, NodeID: clusterVolume.NodeID, NodeLabel: clusterVolume.NodeLabel}, nil
	}

	dir, isBind := BindDirectory(clusterVolume)
	if !isBind {
		return nil, hperrors.Wrap(hperrors.ErrBackupVolumeSharedNotBind).WithParam("Name", setting.Name)
	}
	nodeID, err := dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &HostDir{Dir: dir, NodeID: nodeID, Shared: true}, nil
}

// BindDirectory is the host directory a local volume binds, if it is a bind volume.
func BindDirectory(clusterVolume *entity.ClusterVolume) (string, bool) {
	device := clusterVolume.DriverOpts[DriverOptDevice]
	if device == "" || !filepath.IsAbs(device) {
		return "", false
	}
	if driver := clusterVolume.Driver; driver != "" && driver != "local" {
		return "", false
	}
	if mountType := clusterVolume.DriverOpts[driverOptType]; mountType != "" && mountType != "none" {
		return "", false
	}
	return device, true
}

// resolveVolumeHostPath maps a pinned volume onto its location on the host.
//
// The setting is asked first, because asking docker cannot always work: volumes
// materialize lazily, so one pinned to another node exists in no daemon this
// process can reach. The recorded `device` is the same string VolumeInspect used
// to report in Options["device"].
//
// Inspecting is kept for a volume whose specification records no device - one
// discovered before backfill ran, or a plain local volume whose data sits under
// the daemon's own volume root. That lookup goes by RefID, the docker-side
// identity.
func resolveVolumeHostPath(
	ctx context.Context,
	dockerManager docker.Manager,
	setting *entity.Setting,
	clusterVolume *entity.ClusterVolume,
) (string, error) {
	if devicePath := clusterVolume.DriverOpts[DriverOptDevice]; devicePath != "" {
		return devicePath, nil
	}

	inspectResp, err := dockerManager.VolumeInspect(ctx, setting.RefID)
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrBackupRepoVolumePathUnresolved).WithParam("Name", setting.Name)
	}
	if devicePath := inspectResp.Volume.Options[DriverOptDevice]; devicePath != "" {
		return devicePath, nil
	}
	if inspectResp.Volume.Mountpoint != "" {
		return inspectResp.Volume.Mountpoint, nil
	}
	return "", hperrors.Wrap(hperrors.ErrBackupRepoVolumePathUnresolved).WithParam("Name", setting.Name)
}
