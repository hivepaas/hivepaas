package volumeserviceimpl

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentvolumeclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc/volumeagentdto"
)

// RemoveVolumeInCluster removes a volume where its data actually is.
//
// A volume HivePaaS creates exists on one node - the one its pin names, and only
// from the moment a task first mounts it there. The manager's daemon usually
// never hears of it, so a removal issued here says "no such volume" and reports
// success while the volume sits on the other node untouched. So the removal goes
// to that node, through its agent.
//
// When removeData is set the data goes first, and only the part docker's own
// removal does not cover: see volumeStorageTarget.
func (s *service) RemoveVolumeInCluster(
	ctx context.Context,
	setting *entity.Setting,
	removeData bool,
	force bool,
	retryMax int,
	retryDelay time.Duration,
) error {
	if setting == nil || setting.RefID == "" {
		return nil
	}
	clusterVol, err := setting.AsClusterVolume()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if clusterVol == nil {
		clusterVol = &entity.ClusterVolume{}
	}

	pins := []placementservice.VolumePin{{
		VolumeName: setting.Name,
		NodeID:     clusterVol.NodeID,
		NodeLabel:  clusterVol.NodeLabel,
	}}
	// One pin cannot contradict itself, so there is never a conflict to report.
	constraint, _ := placementservice.VolumePinConstraint(pins)
	onThisNode := s.storageIsOnThisNode(ctx, pins)

	if target, ok := volumeStorageTarget(clusterVol); ok && removeData {
		if err := s.removeStorageTarget(ctx, &target, constraint, onThisNode); err != nil {
			return hperrors.Wrap(err)
		}
	}

	if onThisNode {
		return s.RemoveVolume(ctx, setting.RefID, force, retryMax, retryDelay)
	}
	return s.removeVolumeThroughAgent(ctx, setting.RefID, force, clusterVol)
}

// volumeStorageTarget is the part of a volume's data that has to be deleted
// separately, or false when there is none.
//
// Docker's own removal takes the data with it for a plain local volume: the
// directory it keeps under /var/lib/docker/volumes is the volume. A volume made
// of a host directory is not like that - removing it drops the entry and leaves
// every byte where it was - so the directory is deleted here.
//
// Only a directory HivePaaS chose is deleted, which means one under the
// configured storage root. A directory the operator pointed at is theirs: it can
// be a mount point, or hold things HivePaaS never put there, and the apps' own
// directories inside it have already gone with the apps.
func volumeStorageTarget(vol *entity.ClusterVolume) (storageTarget, bool) {
	dir, _, ok := bindMountTarget(vol, "")
	if !ok {
		return storageTarget{}, false
	}

	root := config.Current().Storage.BindSource
	if root == "" {
		return storageTarget{}, false
	}
	root = filepath.Clean(root)
	dir = filepath.Clean(dir)
	if root == "/" || root == "." || !strings.HasPrefix(dir, root+"/") {
		return storageTarget{}, false
	}

	// The helper is given the directory above the one to delete, so the directory
	// itself goes rather than merely being emptied.
	leaf := safeSubpath(filepath.Base(dir))
	if leaf == "" {
		return storageTarget{}, false
	}
	return storageTarget{
		mount:   bindMountWhole(filepath.Dir(dir)),
		subpath: leaf,
	}, true
}

// removeVolumeThroughAgent removes a volume on the node holding it.
//
// Unlike a local removal this does not retry a volume that is still counted as
// in use: the call crosses a network and the agent's error arrives as a message
// rather than as an error this can recognize. The caller removed whatever was
// using the volume before asking, and a volume that outlives that race is
// reported instead of waited out.
func (s *service) removeVolumeThroughAgent(
	ctx context.Context,
	volumeID string,
	force bool,
	vol *entity.ClusterVolume,
) error {
	var (
		agentAddr string
		err       error
	)
	switch {
	case vol.NodeID != "":
		agentAddr, err = s.agentService.GetAgentAddrForNode(ctx, vol.NodeID)
	case vol.NodeLabel != "":
		agentAddr, err = s.agentService.GetAgentAddrForNodeLabel(ctx, vol.NodeLabel)
	default:
		// Nothing pins it, which is why it was not taken as being on this node.
		return nil
	}
	if err != nil {
		return hperrors.Wrap(err)
	}

	agentClient, err := agentvolumeclient.NewVolumeServiceClient(agentAddr)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = agentClient.Close() }()

	_, err = agentClient.RemoveVolume(ctx, &volumeagentdto.RemoveVolumeReq{
		VolumeID: volumeID,
		Force:    force,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
