package nodecleanupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc/nodecleanupagentdto"
)

func NodeCleanup(
	ctx context.Context,
	uc *nodecleanupagentuc.UC,
	req *agentproto.NodeCleanupReq,
) (*agentproto.NodeCleanupResp, error) {
	if req == nil {
		return &agentproto.NodeCleanupResp{}, nil
	}

	var cleanupSettings *entity.SystemClusterCleanup
	if ps := req.GetCleanupSettings(); ps != nil {
		cleanupSettings = &entity.SystemClusterCleanup{
			Enabled:             ps.GetEnabled(),
			GeneralRetention:    timeutil.Duration(ps.GetGeneralRetention()),
			BuildCacheRetention: timeutil.Duration(ps.GetBuildCacheRetention()),
			PruneImages:         ps.GetPruneImages(),
			PruneVolumes:        ps.GetPruneVolumes(),
			PruneNetworks:       ps.GetPruneNetworks(),
			PruneContainers:     ps.GetPruneContainers(),
			PruneBuildCache:     ps.GetPruneBuildCache(),
		}
	}

	dtoReq := &nodecleanupagentdto.NodeCleanupReq{
		TaskID: req.GetTaskId(),
		//nolint:gosec
		ClusterCleanupReq: clustercleanupservice.ClusterCleanupReq{
			CleanupSettings:   cleanupSettings,
			CleanupContainers: base.CleanupFlag(req.GetCleanupContainers()),
			CleanupImages:     base.CleanupFlag(req.GetCleanupImages()),
			CleanupVolumes:    base.CleanupFlag(req.GetCleanupVolumes()),
			CleanupNetworks:   base.CleanupFlag(req.GetCleanupNetworks()),
			CleanupBuildCache: base.CleanupFlag(req.GetCleanupBuildCache()),
		},
	}

	resp, err := uc.NodeCleanup(ctx, dtoReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if resp == nil {
		return &agentproto.NodeCleanupResp{}, nil
	}

	//nolint:gosec
	return &agentproto.NodeCleanupResp{
		NodeId:                resp.NodeID,
		NodeName:              resp.NodeName,
		ImagesDeleted:         int32(resp.ImagesDeleted),
		ImagesPruneError:      resp.ImagesPruneError,
		VolumesDeleted:        int32(resp.VolumesDeleted),
		VolumesPruneError:     resp.VolumesPruneError,
		ContainersDeleted:     int32(resp.ContainersDeleted),
		ContainersPruneError:  resp.ContainersPruneError,
		TempContainersDeleted: int32(resp.TempContainersDeleted),
		TempServicesDeleted:   int32(resp.TempServicesDeleted),
		NetworksDeleted:       int32(resp.NetworksDeleted),
		NetworksPruneError:    resp.NetworksPruneError,
		BuildCachesDeleted:    int32(resp.BuildCachesDeleted),
		BuildCachesPruneError: resp.BuildCachesPruneError,
		SpaceReclaimed:        resp.SpaceReclaimed,
	}, nil
}
