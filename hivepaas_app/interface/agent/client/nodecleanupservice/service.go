package nodecleanupservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc/nodecleanupagentdto"
)

type NodeCleanupServiceClient interface {
	NodeCleanup(ctx context.Context, req *nodecleanupagentdto.NodeCleanupReq) (*nodecleanupagentdto.NodeCleanupResp, error)
	Close() error
}

type grpcNodeCleanupServiceClient struct {
	protoClient agentproto.NodeCleanupServiceClient
	conn        *grpc.ClientConn
}

func NewNodeCleanupServiceClient(agentAddr string) (NodeCleanupServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &grpcNodeCleanupServiceClient{
		conn:        conn,
		protoClient: agentproto.NewNodeCleanupServiceClient(conn),
	}, nil
}

func (c *grpcNodeCleanupServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcNodeCleanupServiceClient) NodeCleanup(
	ctx context.Context,
	req *nodecleanupagentdto.NodeCleanupReq,
) (*nodecleanupagentdto.NodeCleanupResp, error) {
	authCtx := client.CreateAuthCtx(ctx)

	var protoSettings *agentproto.ClusterCleanupSettings
	if req.CleanupSettings != nil {
		protoSettings = &agentproto.ClusterCleanupSettings{
			Enabled:             req.CleanupSettings.Enabled,
			GeneralRetention:    int64(req.CleanupSettings.GeneralRetention),
			BuildCacheRetention: int64(req.CleanupSettings.BuildCacheRetention),
			PruneImages:         req.CleanupSettings.PruneImages,
			PruneVolumes:        req.CleanupSettings.PruneVolumes,
			PruneNetworks:       req.CleanupSettings.PruneNetworks,
			PruneContainers:     req.CleanupSettings.PruneContainers,
			PruneBuildCache:     req.CleanupSettings.PruneBuildCache,
		}
	}

	protoReq := &agentproto.NodeCleanupReq{
		TaskId:            req.TaskID,
		CleanupSettings:   protoSettings,
		CleanupContainers: int32(req.CleanupContainers),
		CleanupImages:     int32(req.CleanupImages),
		CleanupVolumes:    int32(req.CleanupVolumes),
		CleanupNetworks:   int32(req.CleanupNetworks),
		CleanupBuildCache: int32(req.CleanupBuildCache),
	}

	resp, err := c.protoClient.NodeCleanup(authCtx, protoReq)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodecleanupagentdto.NodeCleanupResp{
		ClusterNodeCleanupOutput: entity.ClusterNodeCleanupOutput{
			NodeID:                resp.GetNodeId(),
			NodeName:              resp.GetNodeName(),
			ImagesDeleted:         int(resp.GetImagesDeleted()),
			ImagesPruneError:      resp.GetImagesPruneError(),
			VolumesDeleted:        int(resp.GetVolumesDeleted()),
			VolumesPruneError:     resp.GetVolumesPruneError(),
			ContainersDeleted:     int(resp.GetContainersDeleted()),
			ContainersPruneError:  resp.GetContainersPruneError(),
			TempContainersDeleted: int(resp.GetTempContainersDeleted()),
			TempServicesDeleted:   int(resp.GetTempServicesDeleted()),
			NetworksDeleted:       int(resp.GetNetworksDeleted()),
			NetworksPruneError:    resp.GetNetworksPruneError(),
			BuildCachesDeleted:    int(resp.GetBuildCachesDeleted()),
			BuildCachesPruneError: resp.GetBuildCachesPruneError(),
			SpaceReclaimed:        resp.GetSpaceReclaimed(),
		},
	}, nil
}
