package client

import (
	"context"

	"google.golang.org/grpc"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// AgentServiceClient defines the client interface for AgentService gRPC methods.
type AgentServiceClient interface {
	Ping(ctx context.Context, req *PingReq) (*PingResp, error)
}

// PingReq represents the ping request payload.
type PingReq struct {
	Message string
}

// PingResp represents the ping response payload.
type PingResp struct {
	Message string
}

type grpcAgentServiceClient struct {
	protoClient agentproto.AgentServiceClient
}

// NewAgentServiceClient creates a new AgentServiceClient wrapping gRPC bindings.
func NewAgentServiceClient(conn grpc.ClientConnInterface) AgentServiceClient {
	return &grpcAgentServiceClient{
		protoClient: agentproto.NewAgentServiceClient(conn),
	}
}

// Ping pings the remote agent server.
func (c *grpcAgentServiceClient) Ping(ctx context.Context, req *PingReq) (*PingResp, error) {
	resp, err := c.protoClient.Ping(ctx, &agentproto.PingReq{
		Message: req.Message,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return &PingResp{
		Message: resp.GetMessage(),
	}, nil
}
