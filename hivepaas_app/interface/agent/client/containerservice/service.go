package containerservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// ContainerServiceClient defines the client interface for ContainerService gRPC methods.
type ContainerServiceClient interface {
	ContainerExec(ctx context.Context) (*ContainerExecStream, error)
	Close() error
}

type grpcContainerServiceClient struct {
	protoClient agentproto.ContainerServiceClient
	conn        *grpc.ClientConn
}

func (c *grpcContainerServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

// ContainerExec starts a container execution session.
func (c *grpcContainerServiceClient) ContainerExec(ctx context.Context) (*ContainerExecStream, error) {
	ctx, cancelFunc := client.CreateAuthCtxWithCancel(ctx)
	stream, err := c.protoClient.ContainerExec(ctx)
	if err != nil {
		cancelFunc()
		return nil, apperrors.Wrap(err)
	}

	execStream := &ContainerExecStream{
		stream:     stream,
		cancelFunc: cancelFunc,
		grpcConn:   c.conn,
	}
	execStream.start()

	return execStream, nil
}

func NewContainerServiceClient(
	agentAddr string,
) (ContainerServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &grpcContainerServiceClient{
		conn:        conn,
		protoClient: agentproto.NewContainerServiceClient(conn),
	}, nil
}
