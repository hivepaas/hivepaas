package dockerapiservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// RemoveAppResult is what an agent removed of an app on its node.
type RemoveAppResult struct {
	Containers int
	Networks   int
	Volumes    int
}

// DockerAPIServiceClient reaches one node's agent about the Docker API it
// serves apps.
type DockerAPIServiceClient interface {
	// Sync has the agent serve exactly the apps that have access now, and
	// returns how many it serves.
	Sync(ctx context.Context) (int, error)
	// RemoveApp has the agent stop serving an app and remove what its children
	// left on the node.
	RemoveApp(ctx context.Context, appID string) (*RemoveAppResult, error)
	Close() error
}

type grpcDockerAPIServiceClient struct {
	protoClient agentproto.DockerAPIServiceClient
	conn        *grpc.ClientConn
}

func NewDockerAPIServiceClient(agentAddr string) (DockerAPIServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcDockerAPIServiceClient{
		conn:        conn,
		protoClient: agentproto.NewDockerAPIServiceClient(conn),
	}, nil
}

func (c *grpcDockerAPIServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcDockerAPIServiceClient) Sync(ctx context.Context) (int, error) {
	resp, err := c.protoClient.SyncDockerAPI(client.CreateAuthCtx(ctx), &agentproto.DockerAPISyncReq{})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int(resp.GetApps()), nil
}

func (c *grpcDockerAPIServiceClient) RemoveApp(ctx context.Context, appID string) (*RemoveAppResult, error) {
	resp, err := c.protoClient.RemoveDockerAPIApp(client.CreateAuthCtx(ctx),
		&agentproto.DockerAPIRemoveAppReq{AppId: appID})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &RemoveAppResult{
		Containers: int(resp.GetContainers()),
		Networks:   int(resp.GetNetworks()),
		Volumes:    int(resp.GetVolumes()),
	}, nil
}
